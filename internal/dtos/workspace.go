package dtos

import "time"

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type WorkspaceResponse struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Slug         string                `json:"slug"`
	PlanName     string                `json:"plan_name"`
	Environments []EnvironmentResponse `json:"environments"`
	CreatedAt    time.Time             `json:"created_at"`
}

type EnvironmentResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WorkspaceMemberResponse struct {
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}
