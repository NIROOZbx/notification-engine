package models

import "time"

type UserPreference struct {
	ID             string    `json:"id"`
	SubscriberID   string    `json:"subscriber_id"`
	ExternalUserID string    `json:"external_user_id"`
	Channel        string    `json:"channel"`
	EventType      string    `json:"event_type,omitempty"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Subscriber struct {
	ID             string         `json:"id"`
	WorkspaceID    string         `json:"workspace_id"`
	EnvironmentID  string         `json:"environment_id"`
	ExternalUserID string         `json:"external_user_id"`
	Channel        string         `json:"channel"`
	ContactValue   string         `json:"contact_value"`
	IsVerified     bool           `json:"is_verified"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}