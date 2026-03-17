package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateAPIKeyParams struct {
	WorkspaceID   pgtype.UUID
	EnvironmentID pgtype.UUID
	UserID        pgtype.UUID
	Label         string
	ExpiresIn     int
}

type ListApiKeyParams struct {
	WorkspaceID   pgtype.UUID
	EnvironmentID pgtype.UUID
}

type RevokeKeyParams struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
}

type DeleteKeyParams struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
}

type ValidatedKey struct {
	ID          pgtype.UUID
	WorkspaceID pgtype.UUID
	EnvID       pgtype.UUID
}

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*dtos.CreateAPIKeyResponse, error)
	ListAPIKeys(ctx context.Context, params ListApiKeyParams) ([]dtos.APIKeyInfo, error)
	RevokeAPIKey(ctx context.Context, params RevokeKeyParams) (*dtos.RevokedKeyInfo, error)
	DeleteAPIKey(ctx context.Context, params DeleteKeyParams) error
	ValidateAPIKey(ctx context.Context, rawKey string) (*ValidatedKey, error)
}

type apiKeyService struct {
	repo *db.Queries
	pool *pgxpool.Pool
}

func NewAPIKeyService(repo *db.Queries) APIKeyService {
	return &apiKeyService{
		repo: repo,
	}
}

func (a *apiKeyService) CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*dtos.CreateAPIKeyResponse, error) {

	env, err := a.repo.GetEnvironmentByID(ctx, params.EnvironmentID)
	if err != nil {
		return nil, err
	}

	if utils.UUIDToString(env.WorkspaceID) != utils.UUIDToString(params.WorkspaceID) {
		return nil, apperrors.ErrForbidden
	}
	workspace, err := a.repo.GetWorkspaceWithPlan(ctx, params.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspace plan: %w", err)
	}

	count, err := a.repo.CountActiveAPIKeys(ctx, params.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to count api keys: %w", err)
	}

	if int(count) >= int(workspace.ApiKeysLimit) {
		return nil, apperrors.ErrForbidden
	}

	expiryTime, err := utils.CalculateExpiry(params.ExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry: %w", err)
	}

	rawKey, hashedKey, hint := utils.GenerateAPIKey(env.Name)

	apiKey, err := a.repo.CreateApiKey(ctx, a.mapToCreateParams(params, expiryTime, hashedKey, hint))
	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return mapToCreateAPIKeyResponse(rawKey, apiKey), nil

}

func (a *apiKeyService) ListAPIKeys(ctx context.Context, params ListApiKeyParams) ([]dtos.APIKeyInfo, error) {

	arg := db.ListAPIKeysByWorkspaceAndEnvParams{
		WorkspaceID:   params.WorkspaceID,
		EnvironmentID: params.EnvironmentID,
	}

	rows, err := a.repo.ListAPIKeysByWorkspaceAndEnv(ctx, arg)

	if err != nil {
		return nil, fmt.Errorf("api_key_service: list failure: %w", err)
	}

	result := make([]dtos.APIKeyInfo, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapToAPIKeyInfo(row))
	}
	return result, nil

}

func (a *apiKeyService) RevokeAPIKey(ctx context.Context, params RevokeKeyParams) (*dtos.RevokedKeyInfo, error) {

	args := db.RevokeAPIKeyParams{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
	}

	row, err := a.repo.RevokeAPIKey(ctx, args)
	if err != nil {
		return nil, apperrors.ErrNotFound
	}

	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		revokedAt = &row.RevokedAt.Time
	}

	return &dtos.RevokedKeyInfo{
		ID:        utils.UUIDToString(row.ID),
		IsRevoked: row.IsRevoked,
		RevokedAt: revokedAt,
	}, nil

}

func (a *apiKeyService) DeleteAPIKey(ctx context.Context, params DeleteKeyParams) error {

	rowsAffected, err := a.repo.DeleteAPIKey(ctx, db.DeleteAPIKeyParams{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (a *apiKeyService) ValidateAPIKey(ctx context.Context, rawKey string) (*ValidatedKey, error) {

	if !strings.HasPrefix(rawKey, "ne_test_") && !strings.HasPrefix(rawKey, "ne_live_") {
		return nil, apperrors.ErrForbidden
	}

	hashedKey := utils.HashAPIKey(rawKey)

	validatedKey, err := a.repo.ValidateAndTouchAPIKey(ctx, hashedKey)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	return &ValidatedKey{
		ID:          validatedKey.ID,
		WorkspaceID: validatedKey.WorkspaceID,
		EnvID:       validatedKey.EnvironmentID,
	}, nil
}

func mapToCreateAPIKeyResponse(rawKey string, row db.ApiKey) *dtos.CreateAPIKeyResponse {
	var expiresAt *time.Time
	if row.ExpiresAt.Valid {
		expiresAt = &row.ExpiresAt.Time
	}
	return &dtos.CreateAPIKeyResponse{
		RawKey:        rawKey,
		ID:            utils.UUIDToString(row.ID),
		Label:         row.Label,
		KeyHint:       row.KeyHint,
		EnvironmentID: utils.UUIDToString(row.EnvironmentID),
		ExpiresAt:     expiresAt,
		CreatedAt:     row.CreatedAt.Time,
	}
}

func (a *apiKeyService) mapToCreateParams(params CreateAPIKeyParams, expiry *time.Time, keyHash, hint string) db.CreateApiKeyParams {
	var expiresAt pgtype.Timestamp
	if expiry != nil {
		expiresAt = pgtype.Timestamp{Time: *expiry, Valid: true}
	}

	return db.CreateApiKeyParams{
		WorkspaceID:   params.WorkspaceID,
		EnvironmentID: params.EnvironmentID,
		Label:         params.Label,
		CreatedBy:     params.UserID,
		ExpiresAt:     expiresAt,
		KeyHash:       keyHash,
		KeyHint:       hint,
	}
}

func mapToAPIKeyInfo(row db.ListAPIKeysByWorkspaceAndEnvRow) dtos.APIKeyInfo {
	var expiresAt, revokedAt, lastUsedAt *time.Time

	if row.ExpiresAt.Valid {
		expiresAt = &row.ExpiresAt.Time
	}
	if row.RevokedAt.Valid {
		revokedAt = &row.RevokedAt.Time
	}
	if row.LastUsedAt.Valid {
		lastUsedAt = &row.LastUsedAt.Time
	}

	return dtos.APIKeyInfo{
		ID:            utils.UUIDToString(row.ID),
		Label:         row.Label,
		KeyHint:       row.KeyHint,
		EnvironmentID: utils.UUIDToString(row.EnvironmentID),
		IsRevoked:     row.IsRevoked,
		RevokedAt:     revokedAt,
		ExpiresAt:     expiresAt,
		LastUsedAt:    lastUsedAt,
		CreatedAt:     row.CreatedAt.Time,
	}

}
