package core

import "time"


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
	ScheduledAt    *time.Time
	TriggerData    map[string]any
	Status         string
}

type UpdateNotificationLogParams struct {
	ID              string
	Status          string
	RenderedContent map[string]any
	AttemptCount    int
	SentAt          *time.Time
	NextRetryAt     *time.Time
	ErrorMessage    *string
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
type GetContactWithPreferenceParams struct {
	WorkspaceID    string
	EnvironmentID  string
	ExternalUserID string
	Channel        string
	EventType      string
}

type ContactPreferencePair struct {
    Contact    *Contact
    Preference *Preference
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
	ID                 string
	TemplateID         string
	Channel            string
	OverrideProviderID string
	Content            map[string]any
	IsActive           bool
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
	ScheduledAt     *time.Time
	NextRetryAt     *time.Time
	ErrorMessage    string
	TriggerData     map[string]any
}
type ChannelConfig struct {
	ID          string
	Channel     string
	Provider    string
	Credentials map[string]string
	Encrypted   string
	IsTest      bool
}

type Layout struct {
	ID   string
	HTML string
}

type resolveProviderParams struct {
	WorkspaceID        string
	Channel            string
	IsTest             bool
	OverrideProviderID string
}

