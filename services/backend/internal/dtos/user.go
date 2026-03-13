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

type UserDTO struct{
	UserDetails UserDetails
	WorkSpaceDetails WorkSpaceDetails
}