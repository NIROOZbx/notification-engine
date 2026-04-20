package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChannelConfig struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	Channel     string
	Provider    string
	DisplayName string
	Credentials map[string]string
	Encrypted   string
	IsActive    bool
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateChannelConfigParams struct {
	WorkspaceID pgtype.UUID
	Channel     string
	Provider    string
	DisplayName string
	Credentials map[string]string
	IsActive    bool
	IsDefault   bool
}

type UpdateChannelConfigParams struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	DisplayName *string
	Credentials map[string]string
	IsActive    *bool
}
