package core

import (
	"context"
)



type Repository interface {
	GetTemplateByEventType(ctx context.Context, workspaceID, envID, eventType string) (*Template, error)
	GetContactByExternalUserAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*Contact, error)
	GetContactWithPreference(ctx context.Context, params GetContactWithPreferenceParams) (*Contact, *Preference, error)
	GetWorkspaceOwners(ctx context.Context, workspaceID string) ([]Contact, error)
	GetPreferencesBySubscriberAndChannel(ctx context.Context, subscriberID, channel, eventType string) ([]Preference, error)
	CreateNotificationLog(ctx context.Context, params CreateLogParams) (*NotificationLog, error)
	GetNotificationLogByID(ctx context.Context, id string) (*NotificationLog, error)
	GetNotificationLogByIdempotencyKey(ctx context.Context, key string) (*NotificationLog, error)
	UpdateNotificationLog(ctx context.Context, params UpdateNotificationLogParams) error
	UpdateNotificationStatus(ctx context.Context, id string, status string) error
	UpdateProviderMessageID(ctx context.Context, id string, providerMessageID string) error
	InsertNotificationAttempt(ctx context.Context, params CreateAttemptParams) error
	GetActiveChannelsByTemplateID(ctx context.Context, templateID string) ([]TemplateChannel, error)
	GetTemplateChannel(ctx context.Context, templateID, channel string) (*TemplateChannel, error)
	GetChannelConfigByID(ctx context.Context, channelConfigID, workspaceID string) (*ChannelConfig, error)
	GetLayoutByID(ctx context.Context, layoutID, workspaceID string) (*Layout, error)
	GetTemplateByID(ctx context.Context, workspaceID, templateID string) (*Template, error)
	GetTemplateWithChannel(ctx context.Context, templateID, workspaceID, channel string) (*Template, *TemplateChannel, error)
	GetDefaultChannelConfig(ctx context.Context, workspaceID, channel string) (*ChannelConfig, error)
	GetProductionEnvironmentID(ctx context.Context, workspaceID string) (string, error)
}

type Producer interface {
	Publish(ctx context.Context, topic string, event any) error
	Close() error
}

type Renderer interface {
	Render(template string, data map[string]any) (string, error)
}
