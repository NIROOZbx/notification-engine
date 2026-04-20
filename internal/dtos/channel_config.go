package dtos


type CreateChannelConfigRequest struct {
    Channel     string         `json:"channel" validate:"required"`
    Provider    string         `json:"provider" validate:"required"`
    DisplayName string         `json:"display_name"`
    Credentials map[string]string `json:"credentials" validate:"required"`
    IsActive    bool           `json:"is_active"`
    IsDefault   bool           `json:"is_default"`
}

type UpdateChannelConfigRequest struct {
    DisplayName *string        `json:"display_name"`
    Credentials map[string]string `json:"credentials"`
    IsActive    *bool          `json:"is_active"`
}

type ChannelConfigResponse struct {
    ID          string         `json:"id"`
    WorkspaceID string         `json:"workspace_id"`
    Channel     string         `json:"channel"`
    Provider    string         `json:"provider"`
    DisplayName string         `json:"display_name"`
    IsActive    bool           `json:"is_active"`
    IsDefault   bool           `json:"is_default"`
    CreatedAt   string         `json:"created_at"`
    UpdatedAt   string         `json:"updated_at"`
}