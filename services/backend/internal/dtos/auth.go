package dtos


type UserDetails struct{
	Name string
	Email string
	AvatarURL string
	UserID string

}

type WorkSpaceDetails struct{
	WorkspaceID string
	WorkSpaceName string
	Slug string
	Role string
}

type AuthResponse struct{
	User       UserDetails `json:"user"`
	Workspace  *WorkSpaceDetails `json:"workspace,omitempty"`
}