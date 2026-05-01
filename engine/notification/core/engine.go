package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/NIROOZbx/notification-engine/engine/notification/queue"
	"github.com/NIROOZbx/notification-engine/internal/billing"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/encryptor"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type Engine struct {
	repo          Repository
	producer      Producer
	providers     map[string]provider.Provider
	log           zerolog.Logger
	renderer      Renderer
	secretKey     string
	billingClient billing.Client
}

type EngineConfig struct {
	Repo          Repository
	Producer      Producer
	Log           zerolog.Logger
	Renderer      Renderer
	SecretKey     string
	BillingClient billing.Client
}

func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		repo:          cfg.Repo,
		producer:      cfg.Producer,
		providers:     make(map[string]provider.Provider),
		log:           cfg.Log,
		renderer:      cfg.Renderer,
		secretKey:     cfg.SecretKey,
		billingClient: cfg.BillingClient,
	}
}

func (e *Engine) resolveEnvID(ctx context.Context, workspaceID, envID string) (string, error) {
	if envID != "" && envID != consts.FallBackUUID {
		return envID, nil
	}
	e.log.Debug().Str("workspace_id", workspaceID).Msg("resolving global environment to production")
	return e.repo.GetProductionEnvironmentID(ctx, workspaceID)
}

func (e *Engine) Ingest(ctx context.Context, workspaceID string, envID string, payload *models.TriggerPayload) error {
	if err := validatePayload(payload, e.log); err != nil {
		return err
	}

	if payload.IsSystem {
		return e.ingestSystem(ctx, workspaceID, envID, payload)
	}

	return e.ingestNormal(ctx, &ingestContext{
		workspaceID: workspaceID,
		envID:       envID,
		payload:     payload,
		strategy:    &normalStrategy{},
	})

}
func (e *Engine) ingestSystem(ctx context.Context, workspaceID, envID string, payload *models.TriggerPayload) error {
	resolvedEnvID, err := e.resolveEnvID(ctx, workspaceID, envID)
	if err != nil {
		return err
	}
	owners, err := e.repo.GetWorkspaceOwners(ctx, workspaceID)
	if err != nil {
		return err
	}
	if len(owners) == 0 {
		e.log.Error().Str("workspace_id", workspaceID).Msg("no active owners found, skipping")
		return nil
	}

	for _, contact := range owners {
		ownerPayload := *payload
		ownerPayload.IdempotencyKey = fmt.Sprintf("%s:%s", payload.IdempotencyKey, contact.ContactValue)
		ic := &ingestContext{
			workspaceID: workspaceID,
			envID:       resolvedEnvID,
			payload:     &ownerPayload,
			strategy:    &systemStrategy{recipientEmail: contact.ContactValue},
		}
		if err := e.ingestNormal(ctx, ic); err != nil {
			e.log.Error().Err(err).Str("owner", contact.ContactValue).Msg("system ingest failed")
			continue
		}
	}

	return nil

}

func (e *Engine) ingestNormal(ctx context.Context, ic *ingestContext) error {

	template, err := e.resolveTemplate(ctx, ic.workspaceID, ic.envID, ic.payload.EventType)
	if err != nil {
		return err
	}
	ic.template = template

	channels, err := e.resolveChannels(ctx, template.ID, ic.payload)
	if err != nil {
		return err
	}

	for _, ch := range channels {
		ic.ch = &ch
		if err := e.ingestChannel(ctx, ic); err != nil {
			e.log.Error().Err(err).Str("channel", ch.Channel).Msg("channel ingest failed")
			continue
		}
	}
	return nil

}

func (e *Engine) resolveTemplate(ctx context.Context, workspaceID, envID, eventType string) (*Template, error) {
	template, err := e.repo.GetTemplateByEventType(ctx, workspaceID, envID, eventType)
	if err != nil {
		e.log.Warn().Err(err).Str("event_type", eventType).Msg("INGEST_STOP: Template not found")
		return nil, fmt.Errorf("template not found: %w", err)
	}
	if template.Status != "live" {
		return nil, fmt.Errorf("template %s is not live (status: %s)", template.ID, template.Status)
	}
	return template, nil
}

func (e *Engine) resolveChannels(ctx context.Context, templateID string, payload *models.TriggerPayload) ([]TemplateChannel, error) {
	activeChannels, err := e.repo.GetActiveChannelsByTemplateID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	var final []TemplateChannel
	for _, ch := range activeChannels {
		if !ch.IsActive {
			continue
		}
		if _, err := queue.TopicByChannel(ch.Channel); err != nil {
			return nil, fmt.Errorf("unsupported channel '%s'", ch.Channel)
		}

		if len(payload.Channels) == 0 || slices.Contains(payload.Channels, ch.Channel) {
			final = append(final, ch)
		}
	}

	if len(final) == 0 {
		e.log.Warn().Msg("INGEST_SKIP: No channels matched after filtering")
	}
	return final, nil
}

func (e *Engine) ingestChannel(ctx context.Context, ic *ingestContext) error {
	l := e.log.With().
		Str("channel", ic.ch.Channel).
		Str("event_type", ic.payload.EventType).
		Str("external_user_id", ic.payload.ExternalUserID).
		Logger()

	if !ic.strategy.SkipBillingCheck() {
		resp, err := e.billingClient.CheckLimit(ctx, ic.workspaceID, ic.envID, ic.ch.Channel)

		if err != nil {
			e.log.Warn().Err(err).Str("channel", ic.ch.Channel).Msg("billing check failed, allowing send")
		} else if !resp.Allowed {
			e.log.Warn().
				Str("reason", resp.Reason).
				Str("channel", ic.ch.Channel).
				Msg("INGEST_SKIP: limit exceeded")
			return nil
		}
	}

	channelKey := fmt.Sprintf("%s:%s", ic.payload.IdempotencyKey, ic.ch.Channel)
	ic.channelKey = channelKey
	_, err := e.repo.GetNotificationLogByIdempotencyKey(ctx, channelKey)
	if err == nil {
		l.Info().Msg("duplicate idempotency key, skipping")
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		l.Error().Err(err).Msg("idempotency check failed")
		return err
	}
	contact, err := ic.strategy.ResolveContact(ctx, e.repo, ic)
	if err != nil {
		return err
	}

	ic.contact = contact
	l.Info().Str("recipient", contact.ContactValue).Msg("found recipient contact")

	if !ic.strategy.SkipOptOut() {
		if e.isOptedOut(ctx, contact.ID, ic.ch.Channel, ic.payload.EventType, l) {
			l.Info().Msg("user has opted out of this channel/event")
			return nil
		}
	}

	return e.createAndPublish(ctx, ic)
}

func (e *Engine) createAndPublish(ctx context.Context, ic *ingestContext) error {
	status := "queued"
	if ic.payload.ScheduledAt != nil && ic.payload.ScheduledAt.After(time.Now()) {
		status = "scheduled"
	}

	logParams := CreateLogParams{
		WorkspaceID:    ic.workspaceID,
		EnvironmentID:  ic.envID,
		TemplateID:     ic.template.ID,
		ExternalUserID: ic.payload.ExternalUserID,
		EventType:      ic.payload.EventType,
		Channel:        ic.ch.Channel,
		IdempotencyKey: ic.channelKey,
		Recipient:      ic.contact.ContactValue,
		IsTest:         ic.payload.IsTest,
		ScheduledAt:    ic.payload.ScheduledAt,
		TriggerData:    ic.payload.Data,
		Status:         status,
	}

	notifLog, err := e.repo.CreateNotificationLog(ctx, logParams)
	if err != nil {
		return fmt.Errorf("create notification log: %w", err)
	}
	if status == "scheduled" {
		e.log.Debug().Str("log_id", notifLog.ID).Msg("notification successfully scheduled for future delivery")
		return nil
	}
	return e.publishToKafka(ctx, notifLog, ic)
}

func (e *Engine) publishToKafka(ctx context.Context, notifLog *NotificationLog, ic *ingestContext) error {

	event := &models.NotificationEvent{
		NotificationLogID: notifLog.ID,
		WorkspaceID:       notifLog.WorkspaceID,
		EnvironmentID:     notifLog.EnvironmentID,
		Channel:           notifLog.Channel,
		Data:              ic.payload.Data,
		Recipient:         ic.contact.ContactValue,
	}
	topic, err := queue.TopicByChannel(ic.ch.Channel)
	if err != nil {
		return fmt.Errorf("unsupported channel: %w", err)
	}
	err = e.producer.Publish(ctx, topic, event)

	if err != nil {
		updateErr := e.repo.UpdateNotificationStatus(ctx, notifLog.ID, "failed")
		if updateErr != nil {
			e.log.Error().Err(updateErr).Msg("failed to update log status")
		}
		return fmt.Errorf("publish to kafka: %w", err)
	}
	return nil
}


func (e *Engine) ProcessDLQ(ctx context.Context, event *models.NotificationEvent) error {
	e.log.Error().
		Str("log_id", event.NotificationLogID).
		Str("channel", event.Channel).
		Int("attempts", event.AttemptNumber).
		Msg("TERMINAL FAILURE: Notification reached DLQ")
	return nil
}

func (e *Engine) Process(ctx context.Context, event *models.NotificationEvent) error {
	e.log.Info().
		Str("log_id", event.NotificationLogID).
		Int("attempt", event.AttemptNumber+1).
		Msg("engine picked up message from queue")

	notifLog, err := e.repo.GetNotificationLogByID(ctx, event.NotificationLogID)
	if err != nil {
		e.log.Error().Err(err).Str("log_id", event.NotificationLogID).Msg("failed to fetch notification log")
		return err
	}

	e.log.Info().
		Str("log_id", notifLog.ID).
		Str("channel", notifLog.Channel).
		Str("status", notifLog.Status).
		Msg("processing notification")

	if notifLog.Recipient == "" {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return fmt.Errorf("recipient is empty for log %q", notifLog.ID)
	}

	err = e.repo.UpdateNotificationStatus(ctx, notifLog.ID, "processing")
	if err != nil {
		return err
	}

	templateChannel, err := e.repo.GetTemplateChannel(ctx, notifLog.TemplateID, notifLog.Channel)
	if err != nil {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return err
	}

	template, err := e.repo.GetTemplateByID(ctx, notifLog.WorkspaceID, templateChannel.TemplateID)
	if err != nil {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return err
	}

	params := &resolveProviderParams{
		WorkspaceID:        notifLog.WorkspaceID,
		Channel:            notifLog.Channel,
		IsTest:             notifLog.IsTest,
		OverrideProviderID: templateChannel.OverrideProviderID,
	}

	p, creds, configID, err := e.resolveProvider(ctx, params)
	if err != nil {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return err
	}

	if err := e.validateConfig(p, creds); err != nil {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return err
	}

	rendered, err := e.renderContent(templateChannel, event.Data, p)
	if err != nil {
		e.updateLogStatus(ctx, notifLog.ID, "failed")
		return err
	}

	if notifLog.Channel == "email" {
		e.log.Debug().Str("layout_id", template.LayoutID).Msg("attempting to wrap email with layout")
		rendered, err = e.wrapWithLayout(ctx, rendered, template.LayoutID, notifLog.WorkspaceID)
		e.log.Info().Interface("rendered", rendered).Str("layout id", template.LayoutID).Str("for workspace", notifLog.WorkspaceID).Msg("rendering content with layout wrapping")
		if err != nil {
			e.log.Error().Err(err).Str("log_id", notifLog.ID).Msg("layout wrapping failed")
			e.updateLogStatus(ctx, notifLog.ID, "failed")
			return err
		}
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	startTime := time.Now()
	providerID, sendErr := p.Send(sendCtx, provider.Message{
		To:      notifLog.Recipient,
		Channel: p.Channel(),
		Content: rendered,
	}, creds)

	if providerID != "" {
		updateErr := e.repo.UpdateProviderMessageID(ctx, notifLog.ID, providerID)
		if updateErr != nil {
			e.log.Error().Err(updateErr).Str("log_id", notifLog.ID).Msg("failed to update provider message ID")
		}
	}

	if !notifLog.IsTest {
		err := e.billingClient.RecordUsage(ctx, billing.RecordUsageInput{
			WorkspaceID:     notifLog.WorkspaceID,
			EnvironmentID:   notifLog.EnvironmentID,
			ChannelConfigID: configID,
			ChannelName:     notifLog.Channel,
			Provider:        p.Name(),
			Success:         sendErr == nil,
		})
		if err != nil {
			e.log.Error().Err(err).Str("log_id", notifLog.ID).Msg("failed to record billing usage")
		}
	}

	duration := time.Since(startTime)
	attemptStatus := "sent"
	var errMessage string
	if sendErr != nil {
		attemptStatus = "failed"
		errMessage = sendErr.Error()
	}

	err = e.repo.InsertNotificationAttempt(ctx, CreateAttemptParams{
		NotificationLogID: notifLog.ID,
		AttemptCount:      event.AttemptNumber + 1,
		Status:            attemptStatus,
		Provider:          p.Name(),
		ErrorMessage:      errMessage,
		DurationMs:        int(duration.Milliseconds()),
	})
	if err != nil {
		e.log.Error().Err(err).Msg("failed to record attempt")
	}

	if sendErr == nil {
		err = e.repo.UpdateNotificationLog(ctx, UpdateNotificationLogParams{
			ID:              notifLog.ID,
			Status:          attemptStatus,
			AttemptCount:    event.AttemptNumber + 1,
			SentAt:          &startTime,
			RenderedContent: rendered,
		})
		if err != nil {
			e.log.Error().Err(err).Msg("failed to update status")
		}
	} else {
		e.handleSendFailure(ctx, notifLog, event, sendErr)
	}

	return nil
}

func (e *Engine) isOptedOut(ctx context.Context, userID, channel, eventType string, l zerolog.Logger) bool {
	prefs, err := e.repo.GetPreferencesBySubscriberAndChannel(ctx, userID, channel, eventType)
	if err != nil {
		l.Error().Err(err).Msg("failed to fetch user preferences, defaulting to opt-in")
		return false
	}

	for _, p := range prefs {
		if !p.IsEnabled {
			return true
		}
	}
	return false
}

func (e *Engine) RegisterProvider(p provider.Provider) {
	key := p.Channel() + ":" + p.Name()
	e.providers[key] = p
}

func (e *Engine) RegisterMockProvider(p provider.Provider) {
	e.providers["mock:"+p.Channel()] = p
}

func (e *Engine) renderContent(templateChannel *TemplateChannel, data map[string]any, p provider.Provider) (map[string]any, error) {
	rendered := map[string]any{}
	for _, key := range p.RequiredContent() {
		val, ok := templateChannel.Content[key].(string)
		if !ok {
			return nil, fmt.Errorf("missing key %q in template content", key)
		}
		renderedVal, err := e.renderer.Render(val, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render %q: %w", key, err)
		}
		rendered[key] = renderedVal
	}
	return rendered, nil
}

func (e *Engine) wrapWithLayout(ctx context.Context, rendered map[string]any, layoutID, workspaceID string) (map[string]any, error) {
	if layoutID == "" {
		return rendered, nil
	}

	layout, err := e.repo.GetLayoutByID(ctx, layoutID, workspaceID)
	if err != nil {
		e.log.Warn().Err(err).Msg("layout not found, sending without layout")
		return rendered, nil
	}

	wrappedBody, err := e.renderer.Render(layout.HTML, map[string]any{
		"content": rendered["body"],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render layout: %w", err)
	}

	rendered["body"] = wrappedBody

	e.log.Info().Msg("successfully rendered content")
	return rendered, nil
}

func (e *Engine) resolveProvider(ctx context.Context, params *resolveProviderParams) (provider.Provider, map[string]string, string, error) {
	if params.IsTest {
		e.log.Info().Str("channel", params.Channel).Msg("using mock provider for test notification")
		p := e.providers["mock:"+params.Channel]
		if p == nil {
			return nil, nil, "", fmt.Errorf("no mock provider for channel %q", params.Channel)
		}
		mockCreds := map[string]string{
			"api_key":    "mock_key",
			"from_email": "mock@test.com",
		}
		return p, mockCreds, "", nil
	}

	var config *ChannelConfig
	var err error

	if params.OverrideProviderID != "" {
		e.log.Debug().Str("config_id", params.OverrideProviderID).Msg("attempting to use provider override")
		config, err = e.repo.GetChannelConfigByID(ctx, params.OverrideProviderID, params.WorkspaceID)
	} else {
		e.log.Debug().Str("channel", params.Channel).Interface("params", params).Msg("attempting to fetch default channel config")
		config, err = e.repo.GetDefaultChannelConfig(ctx, params.WorkspaceID, params.Channel)
	}

	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to fetch channel config: %w", err)
	}

	if config == nil {
		return nil, nil, "", fmt.Errorf("no provider config found (override: %s, channel: %s)", params.OverrideProviderID, params.Channel)
	}

	e.log.Info().
		Str("channel", params.Channel).
		Str("provider", config.Provider).
		Str("config_id", config.ID).
		Msg("successfully resolved channel configuration")

	creds, err := encryptor.DecryptToMap(config.Encrypted, e.secretKey)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	key := params.Channel + ":" + config.Provider
	p, ok := e.providers[key]
	if !ok {
		return nil, nil, "", fmt.Errorf("provider driver %q not found for channel %q", config.Provider, params.Channel)
	}

	return p, creds, config.ID, nil
}

func (e *Engine) updateLogStatus(ctx context.Context, id string, status string) {
	if err := e.repo.UpdateNotificationStatus(ctx, id, status); err != nil {
		e.log.Error().Err(err).Str("log_id", id).Str("status", status).Msg("failed to update log status")
	}
}

func (e *Engine) handleSendFailure(ctx context.Context, notifLog *NotificationLog, event *models.NotificationEvent, sendErr error) {
	newAttemptCount := event.AttemptNumber + 1
	event.AttemptNumber = newAttemptCount

	delay := utils.GetNextRetryDelay(newAttemptCount)

	if delay > 0 {
		nextRetryAt := time.Now().Add(delay)
		errStr := sendErr.Error()
		e.log.Debug().
			Str("log_id", notifLog.ID).
			Int("attempt", newAttemptCount).
			Dur("delay", delay).
			Msg("scheduling retry")

		err := e.repo.UpdateNotificationLog(ctx, UpdateNotificationLogParams{
			ID:           notifLog.ID,
			Status:       "retrying",
			NextRetryAt:  &nextRetryAt,
			ErrorMessage: &errStr,
			AttemptCount: newAttemptCount,
		})
		if err != nil {
			e.log.Error().Err(err).Str("log_id", notifLog.ID).Msg("failed to update status for retry")
		}
		return
	}

	e.log.Warn().Str("log_id", notifLog.ID).Msg("max attempts reached or non-retriable error, routing to DLQ")

	finalErr := sendErr.Error()
	_ = e.repo.UpdateNotificationLog(ctx, UpdateNotificationLogParams{
		ID:           notifLog.ID,
		Status:       "failed",
		ErrorMessage: &finalErr,
		AttemptCount: event.AttemptNumber,
	})

	if err := e.producer.Publish(ctx, queue.TopicDLQ, event); err != nil {
		e.log.Error().Err(err).Msg("failed to publish to DLQ")
	}
}

func validatePayload(payload *models.TriggerPayload, e zerolog.Logger) error {
	if payload == nil {
		return fmt.Errorf("payload is required")
	}
	if payload.IdempotencyKey == "" {
		e.Warn().Msg("idempotency key not provided, generated one")
		payload.IdempotencyKey = uuid.New().String()
	}
	if payload.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if payload.IsSystem {
		return nil
	}
	if payload.ExternalUserID == "" {
		return fmt.Errorf("external_user_id is required")
	}
	return nil

}

func (e *Engine) validateConfig(p provider.Provider, config map[string]string) error {
	for _, field := range p.RequiredFields() {
		val, ok := config[field]
		if !ok || val == "" {
			return fmt.Errorf("missing or empty required configuration field: %s", field)
		}
	}
	return nil
}
