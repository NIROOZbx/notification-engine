package dtos

type UserDetails struct {
	Name      string
	Email     string
	AvatarURL string
	UserID    string
}
type WorkSpaceDetails struct {
	WorkspaceID   string
	WorkSpaceName string
	Slug          string
	Role          string
}

type AuthResponse struct {
	User      UserDetails       `json:"user"`
	Workspace *WorkSpaceDetails `json:"workspace,omitempty"`
}

type OnboardingRequest struct {
	WorkspaceName string `json:"workspace_name" validate:"required"`
}
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
