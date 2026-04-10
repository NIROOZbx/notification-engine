package dtos

import "time"

type CreateAPIKeyResponse struct {
	ID            string     `json:"id"`
	RawKey        string     `json:"raw_key"`
	Label         string     `json:"label"`
	KeyHint       string     `json:"key_hint"`
	EnvironmentID string     `json:"environment_id"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type APIKeyInfo struct {
    ID            string     `json:"id"`
    Label         string     `json:"label"`
    KeyHint       string     `json:"key_hint"`
    EnvironmentID string     `json:"environment_id"`
    IsRevoked     bool       `json:"is_revoked"`
    RevokedAt     *time.Time `json:"revoked_at,omitempty"`
    ExpiresAt     *time.Time `json:"expires_at,omitempty"`
    LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
    CreatedAt     time.Time  `json:"created_at"`
}

type RevokedKeyInfo struct{
	ID        string     `json:"id"`
    IsRevoked bool       `json:"is_revoked"`
    RevokedAt *time.Time `json:"revoked_at"`
}

type CreateAPIKeyRequest struct {
    Label         string `json:"label" validate:"required,min=3,max=50"`
    ExpiresIn     int    `json:"expires_in" validate:"required,min=1"`
}