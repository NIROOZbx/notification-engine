package dtos

type CreateTemplateRequest struct {
	Name        string  `json:"name"  validate:"required"`
	Description string  `json:"description" `
	EventType   string  `json:"event_type"  validate:"required"`
	LayoutID    *string `json:"layout_id" `
}

type UpdateTemplateRequest struct {
    Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
    Description *string `json:"description" validate:"omitempty,max=1000"`
    LayoutID    *string `json:"layout_id"   validate:"omitempty,uuid"`
    Status      *string `json:"status"      validate:"omitempty,oneof=draft live dropped"`
}

type CreateTemplateChannelRequest struct {
	Channel         string         `json:"channel"  validate:"required"`
	ChannelConfigID *string        `json:"channel_config_id" `
	Content         map[string]any `json:"content"  validate:"required"`
}

type UpdateTemplateChannelRequest struct {
	Content  map[string]any `json:"content"`
	IsActive *bool          `json:"is_active"`
}

type TemplateResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	EnvironmentID string  `json:"environment_id"`
	LayoutID      *string `json:"layout_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	EventType     string  `json:"event_type"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type TemplateChannelResponse struct {
	ID              string         `json:"id"`
	TemplateID      string         `json:"template_id"`
	ChannelConfigID *string        `json:"channel_config_id"`
	Channel         string         `json:"channel"`
	Content         map[string]any `json:"content"`
	IsActive        bool           `json:"is_active"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}
