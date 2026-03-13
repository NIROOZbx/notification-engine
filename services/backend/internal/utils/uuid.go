package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDToString(id pgtype.UUID) (string, error) {
	if !id.Valid {
		return "", fmt.Errorf("uuid is null")
	}

	return uuid.UUID(id.Bytes).String(), nil
}

func StringToUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid: %w", err)
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}
