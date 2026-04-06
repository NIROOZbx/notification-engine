package repositories

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/conversion"
)

type notificationRepository struct {
	queries *sqlc.Queries
}


func NewNotificationRepository(queries *sqlc.Queries) *notificationRepository {
	return &notificationRepository{queries: queries}
}

func (r *notificationRepository) CreateNotificationLog(ctx context.Context, params core.CreateLogParams) (*core.NotificationLog, error) {
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
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert notification log: %w", err)
	}
	return mapToCoreLog(row), nil
}

func (r *notificationRepository) GetNotificationLogByID(ctx context.Context, id string) (*core.NotificationLog, error) {
	row, err := r.queries.GetNotificationLogByID(ctx, utils.MustStringToUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get notification log: %w", err)
	}
	return mapToCoreLog(row), nil
}

func (r *notificationRepository) GetNotificationLogByIdempotencyKey(ctx context.Context, key string) (*core.NotificationLog, error) {
	row, err := r.queries.GetNotificationLogByIdempotencyKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification log by idempotency key: %w", err)
	}
	return mapToCoreLog(row), nil
}

func (r *notificationRepository) UpdateNotificationLogStatus(ctx context.Context, params core.UpdateLogParams) error {
	var renderedBytes []byte
	if params.RenderedContent != nil {
		b, err := conversion.JSONBFromMap(params.RenderedContent)
		if err != nil {
			return fmt.Errorf("failed to marshal rendered content: %w", err)
		}
		renderedBytes = b
	}

	_, err := r.queries.UpdateNotificationLogStatus(ctx, sqlc.UpdateNotificationLogStatusParams{
		ID:              utils.MustStringToUUID(params.ID),
		Status:          params.Status,
		RenderedContent: renderedBytes,
		AttemptCount:    int32(params.AttemptCount),
		SentAt:          conversion.TimestampFromTime(params.SentAt),
	})
	return err
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
	row, err := r.queries.GetTemplateByEventType(ctx, sqlc.GetTemplateByEventTypeParams{
		WorkspaceID:   utils.MustStringToUUID(workspaceID),
		EnvironmentID: utils.MustStringToUUID(envID),
		EventType:     eventType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &core.Template{
		ID:          utils.UUIDToString(row.ID),
		WorkspaceID: utils.UUIDToString(row.WorkspaceID),
		EnvID:       utils.UUIDToString(row.EnvironmentID),
		LayoutID:    utils.UUIDToString(row.LayoutID),
		EventType:   row.EventType,
		Status:      row.Status,
		Name:        row.Name,
	}, nil
}

func (r *notificationRepository) GetActiveChannelsByTemplateID(ctx context.Context, templateID string) ([]core.TemplateChannel, error) {
	rows, err := r.queries.GetActiveChannelsByTemplateID(ctx, utils.MustStringToUUID(templateID))
	if err != nil {
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}

	result := make([]core.TemplateChannel, 0, len(rows))
	for _, row := range rows {
		content, err := conversion.MapFromJSONB(row.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal content for channel %s: %w", row.Channel, err)
		}
		result = append(result, core.TemplateChannel{
			ID:              utils.UUIDToString(row.ID),
			TemplateID:      utils.UUIDToString(row.TemplateID),
			ChannelConfigID: utils.UUIDToString(row.ChannelConfigID),
			Channel:         row.Channel,
			Content:         content,
			IsActive:        row.IsActive.Bool,
		})
	}
	return result, nil
}

func (r *notificationRepository) GetTemplateChannel(ctx context.Context, templateID, channel string) (*core.TemplateChannel, error) {
	row, err := r.queries.GetTemplateChannelByTemplateAndChannel(ctx, sqlc.GetTemplateChannelByTemplateAndChannelParams{
		TemplateID: utils.MustStringToUUID(templateID),
		Channel:    channel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template channel: %w", err)
	}

	content, err := conversion.MapFromJSONB(row.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal template channel content: %w", err)
	}

	return &core.TemplateChannel{
		ID:              utils.UUIDToString(row.ID),
		TemplateID:      utils.UUIDToString(row.TemplateID),
		ChannelConfigID: utils.UUIDToString(row.ChannelConfigID),
		Channel:         row.Channel,
		Content:         content,
		IsActive:        row.IsActive.Bool,
	}, nil
}

func (r *notificationRepository) GetChannelConfigByID(ctx context.Context, channelConfigID, workspaceID string) (*core.ChannelConfig, error) {
	row, err := r.queries.GetChannelConfigByID(ctx, sqlc.GetChannelConfigByIDParams{
		ID:          utils.MustStringToUUID(channelConfigID),
		WorkspaceID: utils.MustStringToUUID(workspaceID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get channel config: %w", err)
	}

	creds, err := conversion.MapFromJSONB(row.Credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &core.ChannelConfig{
		ID:          utils.UUIDToString(row.ID),
		Channel:     row.Channel,
		Provider:    row.Provider,
		Credentials: creds,
	}, nil
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
func (r *notificationRepository) GetTemplateByID(ctx context.Context, workspaceID, templateID string) (*core.Template, error) {
	template, err := r.queries.GetTemplateByID(ctx, sqlc.GetTemplateByIDParams{
		ID:          utils.MustStringToUUID(templateID),
		WorkspaceID: utils.MustStringToUUID(workspaceID),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch template %q: %w", templateID, err)
	}

	return &core.Template{
		ID:          template.ID.String(),
		WorkspaceID: template.WorkspaceID.String(),
		EnvID:       template.EnvironmentID.String(),
		LayoutID: func() string {
			if template.LayoutID.Valid {
				return template.LayoutID.String()
			}
			return ""
		}(),
		EventType: template.EventType,
		Status:    template.Status,
		Name:      template.Name,
	}, nil
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

func mapToCoreLog(row sqlc.NotificationLog) *core.NotificationLog {
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
	}
}
