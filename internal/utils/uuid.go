package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

func StringToUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid: %w", err)
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}
func MustStringToUUID(id string) pgtype.UUID {
    parsed, err := uuid.Parse(id)
    if err != nil {
        return pgtype.UUID{}
    }
    return pgtype.UUID{Bytes: parsed, Valid: true}
}