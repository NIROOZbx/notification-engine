package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Layout struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	Name        string
	IsDefault   bool
	Html        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateLayoutParams struct {
	WorkspaceID pgtype.UUID
	Name        string
	Html        string
	IsDefault   bool
}

type UpdateLayoutParams struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	Name        string
	Html        string
	IsDefault   *bool
}
