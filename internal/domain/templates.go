package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Template struct {
	ID            pgtype.UUID
	WorkspaceID   pgtype.UUID
	EnvironmentID pgtype.UUID
	LayoutID      pgtype.UUID 
	CreatedBy     pgtype.UUID
	Name          string
	Description   string
	EventType     string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateTemplateParams struct {
	WorkspaceID   pgtype.UUID
	EnvironmentID pgtype.UUID
	LayoutID      pgtype.UUID
	CreatedBy     pgtype.UUID
	Name          string
	Description   string
	EventType     string
}

type UpdateTemplateParams struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	Name        *string
	Description *string
	LayoutID    *pgtype.UUID
	Status      *string
}

type TemplateChannel struct {
	ID              pgtype.UUID
	TemplateID      pgtype.UUID
	ChannelConfigID pgtype.UUID
	Channel         string
	Content         []byte 
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateTemplateChannelParams struct {
	TemplateID      pgtype.UUID
	WorkspaceID     pgtype.UUID 
	ChannelConfigID pgtype.UUID 
	Channel         string
	Content         map[string]any
}

type UpdateTemplateChannelParams struct {
	ID          pgtype.UUID
	TemplateID  pgtype.UUID
	WorkspaceID pgtype.UUID
	Content     map[string]any
	IsActive    *bool
}
