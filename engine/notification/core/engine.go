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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type CreateLogParams struct {
	WorkspaceID    string
	EnvironmentID  string
	TemplateID     string
	ExternalUserID string
	EventType      string
	Channel        string
	Recipient      string
	IdempotencyKey string
	IsTest         bool
}

type UpdateLogParams struct {
	ID              string
	Status          string
	RenderedContent map[string]any
	AttemptCount    int
	SentAt          *time.Time
}

type CreateAttemptParams struct {
	NotificationLogID string
	AttemptCount      int
	Status            string
	ErrorMessage      string
	ErrorCode         string
	Provider          string
	ChannelConfigID   string
	ProviderMessageID string
	DurationMs        int
}

type Template struct {
	ID          string
	WorkspaceID string
	EnvID       string
	LayoutID    string
	EventType   string
	Status      string
	Name        string
}
type TemplateChannel struct {
	ID              string
	TemplateID      string
	ChannelConfigID string
	Channel         string
	Content         map[string]any
	IsActive        bool
}

type Contact struct {
	ID           string
	ContactValue string
	Channel      string
}

type Preference struct {
	Channel   string
	EventType string
	IsEnabled bool
}

type NotificationLog struct {
	ID              string
	WorkspaceID     string
	EnvironmentID   string
	TemplateID      string
	ExternalUserID  string
	EventType       string
	Channel         string
	Status          string
	Recipient       string
	IdempotencyKey  string
	RenderedContent map[string]any
	IsTest          bool
	AttemptCount    int
}
type ChannelConfig struct {
	ID          string
	Channel     string
	Provider    string
	Credentials map[string]any
	IsTest      bool
}

type Layout struct {
	ID   string
	HTML string
}

type Engine struct {
	repo      Repository
	producer  Producer
	providers map[string]provider.Provider
	log       zerolog.Logger
	renderer  Renderer
}

type Renderer interface {
	Render(template string, data map[string]any) (string, error)
}

type Repository interface {
	GetTemplateByEventType(ctx context.Context, workspaceID, envID, eventType string) (*Template, error)
	GetContactByExternalUserAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*Contact, error)
	GetPreferencesBySubscriberAndChannel(ctx context.Context, subscriberID, channel, eventType string) ([]Preference, error)
	CreateNotificationLog(ctx context.Context, params CreateLogParams) (*NotificationLog, error)
	GetNotificationLogByID(ctx context.Context, id string) (*NotificationLog, error)
	GetNotificationLogByIdempotencyKey(ctx context.Context, key string) (*NotificationLog, error)
	UpdateNotificationLogStatus(ctx context.Context, params UpdateLogParams) error
	InsertNotificationAttempt(ctx context.Context, params CreateAttemptParams) error
	GetActiveChannelsByTemplateID(ctx context.Context, templateID string) ([]TemplateChannel, error)
	GetTemplateChannel(ctx context.Context, templateID, channel string) (*TemplateChannel, error)
	GetChannelConfigByID(ctx context.Context, channelConfigID, workspaceID string) (*ChannelConfig, error)
	GetLayoutByID(ctx context.Context, layoutID, workspaceID string) (*Layout, error)
	GetTemplateByID(ctx context.Context, workspaceID, templateID string) (*Template, error)
}

type Producer interface {
	Publish(ctx context.Context, topic string, event any) error
	Close() error
}

func NewEngine(repo Repository, producer Producer, log zerolog.Logger, render Renderer) *Engine {
	return &Engine{
		repo:      repo,
		producer:  producer,
		providers: make(map[string]provider.Provider),
		log:       log,
		renderer:  render,
	}
}

func (e *Engine) Ingest(ctx context.Context, workspaceID string, envID string, payload *models.TriggerPayload) error {
	if err := validatePayload(payload, e.log); err != nil {
		return err
	}

	template, err := e.repo.GetTemplateByEventType(ctx, workspaceID, envID, payload.EventType)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	if template.Status != "live" {
		return fmt.Errorf("template %s is not live (status: %s)", template.ID, template.Status)
	}

	activeChannels, err := e.repo.GetActiveChannelsByTemplateID(ctx, template.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %w", err)
	}
	if len(activeChannels) == 0 {
		return fmt.Errorf("no active channels for template")
	}

	var finalChannel []TemplateChannel
	for _, ch := range activeChannels {
		if _, err := topicByChannel(ch.Channel); err != nil {
			return fmt.Errorf("configuration error: unsupported channel '%s'", ch.Channel)
		}
		if len(payload.Channels) == 0 || slices.Contains(payload.Channels, ch.Channel) {
            finalChannel = append(finalChannel, ch)
        }
	}

	for _, ch := range finalChannel {
		l := e.log.With().
			Str("channel", ch.Channel).
			Str("event_type", payload.EventType).
			Str("external_user_id", payload.ExternalUserID).
			Logger()

		

		channelKey := fmt.Sprintf("%s:%s", payload.IdempotencyKey, ch.Channel)
		_, err := e.repo.GetNotificationLogByIdempotencyKey(ctx, channelKey)
		if err == nil {
			l.Info().Msg("duplicate idempotency key, skipping")
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			l.Error().Err(err).Msg("idempotency check failed")
			continue
		}
		contact, err := e.repo.GetContactByExternalUserAndChannel(ctx, workspaceID, envID, payload.ExternalUserID, ch.Channel)
		if err != nil {
			l.Warn().Err(err).Msg("recipient contact not found, skipping")
			continue
		}

		if e.isOptedOut(ctx, contact.ID, ch.Channel, payload.EventType, l) {
			l.Info().Msg("user has opted out of this channel/event")
			continue
		}

		logParams := CreateLogParams{
			WorkspaceID:    workspaceID,
			EnvironmentID:  envID,
			TemplateID:     template.ID,
			ExternalUserID: payload.ExternalUserID,
			EventType:      payload.EventType,
			Channel:        ch.Channel,
			IdempotencyKey: channelKey,
			Recipient:      contact.ContactValue,
			IsTest:         payload.IsTest,
		}

		notifLog, err := e.repo.CreateNotificationLog(ctx, logParams)
		if err != nil {
			l.Error().Err(err).Str("channel", ch.Channel).Msg("failed to create notification log")
			continue
		}
		event := models.NotificationEvent{
			NotificationLogID: notifLog.ID,
			WorkspaceID:       notifLog.WorkspaceID,
			EnvironmentID:     notifLog.EnvironmentID,
			Channel:           notifLog.Channel,
			Data:              payload.Data,
			AttemptNumber:     0,
			Recipient:         contact.ContactValue,
		}
		topic, _ := topicByChannel(ch.Channel)

		err = e.producer.Publish(ctx, topic, event)

		if err != nil {
			l.Error().Err(err).Str("log_id", notifLog.ID).Msg("failed to publish to kafka")
			params := UpdateLogParams{
				ID:              notifLog.ID,
				Status:          "failed",
				RenderedContent: nil,
				AttemptCount:    0,
			}
			err = e.repo.UpdateNotificationLogStatus(ctx, params)
			if err != nil {
				l.Error().Err(err).Msg("failed to update log status after kafka failure")
			}
			continue
		}
		l.Debug().Str("log_id", notifLog.ID).Msg("notification successfully ingested")
	}
	return nil

}

func (e *Engine) Process(ctx context.Context, event models.NotificationEvent) error {
	notifLogs, err := e.repo.GetNotificationLogByID(ctx, event.NotificationLogID)

	if err != nil {
		return err
	}

	if notifLogs.Recipient == "" {
		e.updateLogStatus(ctx, notifLogs.ID, "failed")
		return fmt.Errorf("recipient is empty for log %q", notifLogs.ID)
	}

	err = e.repo.UpdateNotificationLogStatus(ctx, UpdateLogParams{
		ID:           notifLogs.ID,
		Status:       "processing",
		AttemptCount: notifLogs.AttemptCount,
	})

	if err != nil {
		return err
	}
	templateChannel, err := e.repo.GetTemplateChannel(ctx, notifLogs.TemplateID, notifLogs.Channel)
	if err != nil {
		e.updateLogStatus(ctx, notifLogs.ID, "failed")
		return err
	}
	template, err := e.repo.GetTemplateByID(ctx, event.WorkspaceID, templateChannel.TemplateID)
	if err != nil {
		e.updateLogStatus(ctx, notifLogs.ID, "failed")
		return err
	}

	p, err := e.resolveProvider(notifLogs)
	if err != nil {
		e.updateLogStatus(ctx, notifLogs.ID, "failed")

		return err
	}

	rendered, err := e.renderContent(templateChannel, event.Data, p)
	if err != nil {
		e.updateLogStatus(ctx, notifLogs.ID, "failed")
		return err
	}

	if notifLogs.Channel == "email" {
		rendered, err = e.wrapWithLayout(ctx, rendered, template.LayoutID, notifLogs.WorkspaceID)
		if err != nil {
			e.updateLogStatus(ctx, notifLogs.ID, "failed")
			return err
		}
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	startTime := time.Now()
	sendErr := p.Send(sendCtx, provider.Message{
		To:      notifLogs.Recipient,
		Channel: p.Channel(),
		Content: rendered,
	})

	status := "sent"

	if sendErr != nil {
		status = "failed"
	}

	duration := time.Since(startTime)

	var errMessage string
	if sendErr != nil {
		errMessage = sendErr.Error()
	}

	err = e.repo.InsertNotificationAttempt(ctx, CreateAttemptParams{
		NotificationLogID: notifLogs.ID,
		AttemptCount:      notifLogs.AttemptCount + 1,
		Status:            status,
		Provider:          p.Name(),
		ErrorMessage:      errMessage,
		DurationMs:        int(duration.Milliseconds()),
	})

	if err != nil {
		e.log.Error().Err(err).Msg("failed to record attempt")
	}

	updateParams := UpdateLogParams{
		ID:           notifLogs.ID,
		Status:       status,
		AttemptCount: notifLogs.AttemptCount + 1,
	}
	if sendErr == nil {
		updateParams.SentAt = &startTime
		updateParams.RenderedContent = rendered
	} else {
		updateParams.RenderedContent = nil
	}

	err = e.repo.UpdateNotificationLogStatus(ctx, updateParams)
	if err != nil {
		e.log.Error().Err(err).Msg("failed to update status to processing, continuing")
	}

	if sendErr != nil {
		e.handleSendFailure(ctx, notifLogs, event)

	}

	// channelConfig,err:=e.repo.GetChannelConfigByID(ctx, templateChannel.ChannelConfigID, workspaceID)
	return nil
}

func topicByChannel(channel string) (string, error) {
	switch channel {
	case "email":
		return queue.TopicEmail, nil
	case "sms":
		return queue.TopicSMS, nil
	case "push":
		return queue.TopicPush, nil
	case "slack":
		return queue.TopicSlack, nil
	case "whatsapp":
		return queue.TopicWhatsApp, nil
	case "webhook":
		return queue.TopicWebhook, nil
	case "in_app":
		return queue.TopicInApp, nil
	default:
		return "", fmt.Errorf("unsupported or unknown channel: %s", channel)
	}
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
	e.providers[p.Channel()] = p
}

func (e *Engine) RegisterMockProvider(p provider.Provider) {
	e.providers["mock:"+p.Channel()] = p
}

func (e *Engine) renderContent(templateChannel *TemplateChannel, data map[string]any, p provider.Provider) (map[string]any, error) {
	rendered := map[string]any{}
	for _, key := range p.RequiredFields() {
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
	return rendered, nil
}

func (e *Engine) resolveProvider(notifLog *NotificationLog) (provider.Provider, error) {
	if notifLog.IsTest {
		p := e.providers["mock:"+notifLog.Channel]
		if p == nil {
			return nil, fmt.Errorf("no mock provider for channel %q", notifLog.Channel)
		}
		return p, nil
	}

	p, ok := e.providers[notifLog.Channel]
	if !ok {
		return nil, fmt.Errorf("no provider for channel %q", notifLog.Channel)
	}

	return p, nil
}

func (e *Engine) updateLogStatus(ctx context.Context, id string, status string) {
	if err := e.repo.UpdateNotificationLogStatus(ctx, UpdateLogParams{
		ID:     id,
		Status: status,
	}); err != nil {
		e.log.Error().Err(err).Str("log_id", id).Str("status", status).Msg("failed to update log status")
	}
}

func (e *Engine) handleSendFailure(ctx context.Context, notifLog *NotificationLog, event models.NotificationEvent) {
	newAttemptCount := notifLog.AttemptCount + 1
	event.AttemptNumber = newAttemptCount

	if newAttemptCount >= consts.MaxAttempts {
		e.log.Warn().Str("log_id", notifLog.ID).Msg("max attempts reached, routing to DLQ")
		if err := e.producer.Publish(ctx, queue.TopicDLQ, event); err != nil {
			e.log.Error().Err(err).Msg("failed to publish to DLQ")
		}
		return
	}

	e.log.Info().Str("log_id", notifLog.ID).Int("attempt", newAttemptCount).Msg("routing to Retry topic")
	if err := e.producer.Publish(ctx, queue.TopicRetry, event); err != nil {
		e.log.Error().Err(err).Msg("failed to publish to Retry queue")
	}
}


func validatePayload(payload *models.TriggerPayload, e zerolog.Logger) error {
	if payload == nil {
		return fmt.Errorf("payload is required")
	}
	if payload.ExternalUserID == "" {
		return fmt.Errorf("external_user_id is required")
	}
	if payload.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if payload.IdempotencyKey == "" {
		e.Warn().Msg("idempotency key not provided, generated one")
		payload.IdempotencyKey = uuid.New().String()
	}
	return nil

}
