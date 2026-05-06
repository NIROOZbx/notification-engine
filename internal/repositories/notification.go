package repositories

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/conversion"
)

type notificationRepository struct {
	queries    *sqlc.Queries
	configRepo ChannelConfigRepo
	tplRepo    TemplateRepository
}

func NewNotificationRepository(queries *sqlc.Queries, configRepo ChannelConfigRepo, tplRepo TemplateRepository) *notificationRepository {
	return &notificationRepository{
		queries:    queries,
		configRepo: configRepo,
		tplRepo:    tplRepo,
	}
}

func (r *notificationRepository) CreateNotificationLog(ctx context.Context, params core.CreateLogParams) (*core.NotificationLog, error) {
	triggerDataBytes, err := conversion.JSONBFromMap(params.TriggerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
	}

	row, err := r.queries.InsertNotificationLog(ctx, sqlc.InsertNotificationLogParams{
		WorkspaceID:    utils.MustStringToUUID(params.WorkspaceID),
		EnvironmentID:  utils.MustStringToUUID(params.EnvironmentID),
		TemplateID:     utils.MustStringToUUID(params.TemplateID),
		ExternalUserID: params.ExternalUserID,
		EventType:      params.EventType,
		Channel:        params.Channel,
		Recipient:      params.Recipient,
		IdempotencyKey: params.IdempotencyKey,
		IsTest:         params.IsTest,
		ScheduledAt:    conversion.TimestampFromPtr(params.ScheduledAt),
		TriggerData:    triggerDataBytes,
		Status:         params.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert notification log: %w", err)
	}
	return MapToCoreLog(row), nil
}

func (r *notificationRepository) GetNotificationLogByID(ctx context.Context, id string) (*core.NotificationLog, error) {
	row, err := r.queries.GetNotificationLogByID(ctx, utils.MustStringToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get notification log: %w", err)
	}
	return MapToCoreLog(row), nil
}

func (r *notificationRepository) GetNotificationLogByIdempotencyKey(ctx context.Context, key string) (*core.NotificationLog, error) {
	row, err := r.queries.GetNotificationLogByIdempotencyKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification log by idempotency key: %w", err)
	}
	return MapToCoreLog(row), nil
}

func (r *notificationRepository) UpdateNotificationLog(ctx context.Context, params core.UpdateNotificationLogParams) error {
	rendered, err := conversion.JSONBFromMap(params.RenderedContent)
	if err != nil {
		return fmt.Errorf("failed to marshal rendered content: %w", err)
	}

	_, err = r.queries.UpdateNotificationLog(ctx, sqlc.UpdateNotificationLogParams{
		ID:              utils.MustStringToUUID(params.ID),
		Status:          params.Status,
		RenderedContent: rendered,
		AttemptCount:    int32(params.AttemptCount),
		SentAt:          conversion.TimestampFromPtr(params.SentAt),
		NextRetryAt:     conversion.TimestampFromPtr(params.NextRetryAt),
		ErrorMessage:    conversion.TextFromPtr(params.ErrorMessage),
	})
	return err
}

func (r *notificationRepository) UpdateNotificationStatus(ctx context.Context, id string, status string) error {
	_, err := r.queries.UpdateNotificationStatus(ctx, sqlc.UpdateNotificationStatusParams{
		ID:     utils.MustStringToUUID(id),
		Status: status,
	})
	return err
}

func (r *notificationRepository) UpdateProviderMessageID(ctx context.Context, id string, providerMessageID string) error {
	return r.queries.UpdateProviderMessageID(ctx, sqlc.UpdateProviderMessageIDParams{
		ID:                utils.MustStringToUUID(id),
		ProviderMessageID: helpers.Text(providerMessageID),
	})
}

func (r *notificationRepository) UpdateDeliveryStatusByProviderID(ctx context.Context, input domain.UpdateDeliveryStatusInput) error {
	return r.queries.UpdateDeliveryStatusByProviderID(ctx, sqlc.UpdateDeliveryStatusByProviderIDParams{
		ProviderMessageID: helpers.Text(input.ProviderMessageID),
		DeliveryStatus:    helpers.Text(input.DeliveryStatus),
		DeliveredAt:       helpers.Timestamp(input.Timestamp), // always pass — DB CASE WHEN decides
		ProviderResponse:  helpers.Text(input.ErrorMessage),
	})
}

func (r *notificationRepository) InsertNotificationAttempt(ctx context.Context, params core.CreateAttemptParams) error {
	_, err := r.queries.InsertNotificationAttempt(ctx, sqlc.InsertNotificationAttemptParams{
		NotificationLogID: utils.MustStringToUUID(params.NotificationLogID),
		AttemptCount:      int32(params.AttemptCount),
		Provider:          params.Provider,
		Status:            params.Status,
		ErrorMessage:      helpers.Text(params.ErrorMessage),
		DurationMs:        helpers.Int4(int32(params.DurationMs)),
	})
	return err
}

func (r *notificationRepository) GetTemplateByEventType(ctx context.Context, workspaceID, envID, eventType string) (*core.Template, error) {
	tpl, err := r.tplRepo.GetByEventType(ctx, utils.MustStringToUUID(workspaceID), utils.MustStringToUUID(envID), eventType)
	if err != nil {
		return nil, err
	}
	return mapDomainToCoreTemplate(tpl), nil
}

func (r *notificationRepository) GetActiveChannelsByTemplateID(ctx context.Context, templateID string) ([]core.TemplateChannel, error) {
	channels, err := r.tplRepo.ListChannels(ctx, utils.MustStringToUUID(templateID))
	if err != nil {
		return nil, err
	}

	result := make([]core.TemplateChannel, 0, len(channels))
	for _, ch := range channels {
		coreCh, err := mapDomainToCoreTemplateChannel(ch)
		if err != nil {
			return nil, err
		}
		result = append(result, *coreCh)
	}
	return result, nil
}

func (r *notificationRepository) GetTemplateChannel(ctx context.Context, templateID, channel string) (*core.TemplateChannel, error) {
	ch, err := r.tplRepo.GetChannelByTemplateAndChannel(ctx, utils.MustStringToUUID(templateID), channel)
	if err != nil {
		return nil, err
	}
	return mapDomainToCoreTemplateChannel(ch)
}

func (r *notificationRepository) GetTemplateByID(ctx context.Context, workspaceID, templateID string) (*core.Template, error) {
	tpl, err := r.tplRepo.GetByID(ctx, utils.MustStringToUUID(templateID), utils.MustStringToUUID(workspaceID))
	if err != nil {
		return nil, err
	}
	return mapDomainToCoreTemplate(tpl), nil
}

func (r *notificationRepository) GetTemplateWithChannel(ctx context.Context, templateID, workspaceID, channel string) (*core.Template, *core.TemplateChannel, error) {
	row, err := r.queries.GetTemplateWithChannel(ctx, sqlc.GetTemplateWithChannelParams{
		ID:          utils.MustStringToUUID(templateID),
		WorkspaceID: utils.MustStringToUUID(workspaceID),
		Channel:     channel,
	})
	if err != nil {
		return nil, nil, err
	}

	tpl := mapSqlcToCoreTemplate(row.Template)
	ch, err := mapSqlcToCoreTemplateChannel(row.TemplateChannel)
	if err != nil {
		return nil, nil, err
	}

	return tpl, ch, nil
}

func (r *notificationRepository) GetChannelConfigByID(ctx context.Context, channelConfigID, workspaceID string) (*core.ChannelConfig, error) {
	cfg, err := r.configRepo.GetChannelConfigByID(ctx, utils.MustStringToUUID(channelConfigID), utils.MustStringToUUID(workspaceID))
	if err != nil {
		return nil, err
	}

	return mapDomainToCoreChannelConfig(cfg), nil
}

func (r *notificationRepository) GetDefaultChannelConfig(ctx context.Context, workspaceID, channel string) (*core.ChannelConfig, error) {
	cfg, err := r.configRepo.GetDefaultChannelConfig(ctx, utils.MustStringToUUID(workspaceID), channel)
	if err != nil {
		return nil, err
	}

	return mapDomainToCoreChannelConfig(cfg), nil
}

func (r *notificationRepository) GetContactByExternalUserAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*core.Contact, error) {
	row, err := r.queries.GetContactByExternalUserAndChannel(ctx, sqlc.GetContactByExternalUserAndChannelParams{
		WorkspaceID:    utils.MustStringToUUID(workspaceID),
		EnvironmentID:  utils.MustStringToUUID(envID),
		ExternalUserID: externalUserID,
		Channel:        channel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user contact: %w", err)
	}

	return &core.Contact{
		ID:           utils.UUIDToString(row.ID),
		ContactValue: row.ContactValue,
		Channel:      row.Channel,
	}, nil
}

func (r *notificationRepository) GetContactWithPreference(ctx context.Context, params core.GetContactWithPreferenceParams) (*core.Contact, *core.Preference, error) {
	row, err := r.queries.GetContactWithPreference(ctx, sqlc.GetContactWithPreferenceParams{
		ExternalUserID: params.ExternalUserID,
		WorkspaceID:    utils.MustStringToUUID(params.WorkspaceID),
		EnvironmentID:  utils.MustStringToUUID(params.EnvironmentID),
		Channel:        params.Channel,
		EventType:      helpers.Text(params.EventType),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get contact with preference: %w", err)
	}

	contact := &core.Contact{
		ID:           utils.UUIDToString(row.UserInfo.ID),
		ContactValue: row.UserInfo.ContactValue,
		Channel:      row.UserInfo.Channel,
	}

	var preference *core.Preference
	if row.UserPreference.ID.Valid {
		preference = &core.Preference{
			Channel:   row.UserPreference.Channel,
			EventType: row.UserPreference.EventType.String,
			IsEnabled: row.UserPreference.IsEnabled,
		}
	}

	return contact, preference, nil
}

func (r *notificationRepository) GetWorkspaceOwners(ctx context.Context, workspaceID string) ([]core.Contact, error) {
	emails, err := r.queries.GetWorkspaceOwnerEmails(ctx, utils.MustStringToUUID(workspaceID))
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace owners: %w", err)
	}

	var contacts []core.Contact
	for _, email := range emails {
		contacts = append(contacts, core.Contact{
			ContactValue: email,
			Channel:      "email",
		})
	}
	return contacts, nil
}

func (r *notificationRepository) GetProductionEnvironmentID(ctx context.Context, workspaceID string) (string, error) {
	id, err := r.queries.GetProductionEnvironmentByWorkspace(ctx, utils.MustStringToUUID(workspaceID))
	if err != nil {
		return "", fmt.Errorf("failed to get production environment: %w", err)
	}

	return utils.UUIDToString(id), nil
}

func (r *notificationRepository) GetPreferencesBySubscriberAndChannel(ctx context.Context, subscriberID, channel, eventType string) ([]core.Preference, error) {
	rows, err := r.queries.GetPreferencesBySubscriberAndChannel(ctx, sqlc.GetPreferencesBySubscriberAndChannelParams{
		SubscriberID: utils.MustStringToUUID(subscriberID),
		Channel:      channel,
		EventType:    helpers.Text(eventType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	result := make([]core.Preference, 0, len(rows))
	for _, row := range rows {
		result = append(result, core.Preference{
			Channel:   row.Channel,
			EventType: row.EventType.String,
			IsEnabled: row.IsEnabled,
		})
	}
	return result, nil
}

func (r *notificationRepository) GetLayoutByID(ctx context.Context, layoutID, workspaceID string) (*core.Layout, error) {
	layout, err := r.queries.GetLayoutByID(ctx, sqlc.GetLayoutByIDParams{
		ID:          utils.MustStringToUUID(layoutID),
		WorkspaceID: utils.MustStringToUUID(workspaceID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get layout %q: %w", layoutID, err)
	}

	return &core.Layout{
		ID:   utils.UUIDToString(layout.ID),
		HTML: layout.Html,
	}, nil
}

// Mappers

func MapToCoreLog(row sqlc.NotificationLog) *core.NotificationLog {
	triggerData, _ := conversion.MapFromJSONB(row.TriggerData)

	return &core.NotificationLog{
		ID:             utils.UUIDToString(row.ID),
		WorkspaceID:    utils.UUIDToString(row.WorkspaceID),
		EnvironmentID:  utils.UUIDToString(row.EnvironmentID),
		TemplateID:     utils.UUIDToString(row.TemplateID),
		ExternalUserID: row.ExternalUserID,
		EventType:      row.EventType,
		Channel:        row.Channel,
		Status:         row.Status,
		Recipient:      row.Recipient,
		IdempotencyKey: row.IdempotencyKey,
		IsTest:         row.IsTest,
		AttemptCount:   int(row.AttemptCount),
		ScheduledAt:    conversion.TimeFromTimestamp(row.ScheduledAt),
		NextRetryAt:    conversion.TimeFromTimestamp(row.NextRetryAt),
		ErrorMessage:   row.ErrorMessage.String,
		TriggerData:    triggerData,
	}
}

func mapDomainToCoreTemplate(tpl *domain.Template) *core.Template {
	if tpl == nil {
		return nil
	}
	return &core.Template{
		ID:          utils.UUIDToString(tpl.ID),
		WorkspaceID: utils.UUIDToString(tpl.WorkspaceID),
		EnvID:       utils.UUIDToString(tpl.EnvironmentID),
		LayoutID:    utils.UUIDToString(tpl.LayoutID),
		EventType:   tpl.EventType,
		Status:      tpl.Status,
		Name:        tpl.Name,
	}
}

func mapDomainToCoreTemplateChannel(ch *domain.TemplateChannel) (*core.TemplateChannel, error) {
	if ch == nil {
		return nil, nil
	}
	content, err := conversion.MapFromJSONB(ch.Content)
	if err != nil {
		return nil, err
	}
	return &core.TemplateChannel{
		ID:                 utils.UUIDToString(ch.ID),
		TemplateID:         utils.UUIDToString(ch.TemplateID),
		OverrideProviderID: utils.UUIDToString(ch.ChannelConfigID),
		Channel:            ch.Channel,
		Content:            content,
		IsActive:           ch.IsActive,
	}, nil
}

func mapDomainToCoreChannelConfig(row *domain.ChannelConfig) *core.ChannelConfig {
	if row == nil {
		return nil
	}
	return &core.ChannelConfig{
		ID:          utils.UUIDToString(row.ID),
		Channel:     row.Channel,
		Provider:    row.Provider,
		Encrypted:   row.Encrypted,
		Credentials: row.Credentials,
	}
}

func mapSqlcToCoreTemplate(tpl sqlc.Template) *core.Template {
	return &core.Template{
		ID:          utils.UUIDToString(tpl.ID),
		WorkspaceID: utils.UUIDToString(tpl.WorkspaceID),
		EnvID:       utils.UUIDToString(tpl.EnvironmentID),
		LayoutID:    utils.UUIDToString(tpl.LayoutID),
		EventType:   tpl.EventType,
		Status:      tpl.Status,
		Name:        tpl.Name,
	}
}

func mapSqlcToCoreTemplateChannel(ch sqlc.TemplateChannel) (*core.TemplateChannel, error) {
	content, err := conversion.MapFromJSONB(ch.Content)
	if err != nil {
		return nil, err
	}
	return &core.TemplateChannel{
		ID:                 utils.UUIDToString(ch.ID),
		TemplateID:         utils.UUIDToString(ch.TemplateID),
		OverrideProviderID: utils.UUIDToString(ch.ChannelConfigID),
		Channel:            ch.Channel,
		Content:            content,
		IsActive:           ch.IsActive.Bool,
	}, nil
}
