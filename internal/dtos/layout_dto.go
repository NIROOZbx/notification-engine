package dtos

type CreateLayoutRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=255"`
	Html      string `json:"html" validate:"required"`
	IsDefault *bool  `json:"is_default"`
}

type UpdateLayoutRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Html      *string `json:"html,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
}

type LayoutResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	IsDefault   bool   `json:"is_default"`
	Html        string `json:"html"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
