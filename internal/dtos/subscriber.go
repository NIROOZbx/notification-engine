package dtos

type IdentifyRequest struct {
	ExternalUserID string         `json:"external_user_id" validate:"required"`
	Channel        string         `json:"channel"          validate:"required"`
	ContactValue   string         `json:"contact_value"    validate:"required"`
	Metadata       map[string]any `json:"metadata"`
}

type UpsertPreferenceRequest struct {
	ExternalUserID string `json:"external_user_id" validate:"required"`
	Channel        string `json:"channel"          validate:"required"`
	EventType      string `json:"event_type"`
	IsEnabled      bool   `json:"is_enabled"`
}

type SubscriberResponse struct {
	ID             string         `json:"id"`
	WorkspaceID    string         `json:"workspace_id"`
	EnvironmentID  string         `json:"environment_id"`
	ExternalUserID string         `json:"external_user_id"`
	Channel        string         `json:"channel"`
	ContactValue   string         `json:"contact_value"`
	IsVerified     bool           `json:"is_verified"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type UserPreferenceResponse struct {
	ID             string `json:"id"`
	SubscriberID   string `json:"subscriber_id"`
	ExternalUserID string `json:"external_user_id,omitempty"`
	Channel        string `json:"channel"`
	EventType      string `json:"event_type,omitempty"`
	IsEnabled      bool   `json:"is_enabled"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type SubscriberListResponse struct {
	Subscribers []SubscriberResponse `json:"subscribers"`
	TotalCount  int64               `json:"total_count"`
	TotalPages  int32               `json:"total_pages"`
	CurrentPage int32               `json:"current_page"`
	PageSize    int32               `json:"page_size"`
}

