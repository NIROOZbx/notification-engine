package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/parallel"
	"github.com/jackc/pgx/v5/pgtype"
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
	IsTest      bool
}

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*dtos.CreateAPIKeyResponse, error)
	ListAPIKeys(ctx context.Context, params ListApiKeyParams) ([]dtos.APIKeyInfo, error)
	RevokeAPIKey(ctx context.Context, params RevokeKeyParams) (*dtos.RevokedKeyInfo, error)
	DeleteAPIKey(ctx context.Context, params DeleteKeyParams) error
	ValidateAPIKey(ctx context.Context, rawKey string) (*ValidatedKey, error)
}

type apiKeyService struct {
	repo repositories.APIKeyRepository
}

func NewAPIKeyService(repo repositories.APIKeyRepository) APIKeyService {
	return &apiKeyService{
		repo: repo,
	}
}

func (a *apiKeyService) CreateAPIKey(ctx context.Context, params CreateAPIKeyParams) (*dtos.CreateAPIKeyResponse, error) {

	env, workspace, count, err := parallel.Query3(ctx,
		func(c context.Context) (sqlc.Environment, error) {
			return a.repo.GetEnvironment(c, params.EnvironmentID)
		},
		func(c context.Context) (sqlc.GetWorkspaceWithPlanRow, error) {
			return a.repo.GetWorkspaceWithPlan(c, params.WorkspaceID)
		},
		func(c context.Context) (int64, error) {
			return a.repo.CountActiveKeys(c, params.WorkspaceID)
		},
	)

	if err != nil {
		return nil, err
	}

	if utils.UUIDToString(env.WorkspaceID) != utils.UUIDToString(params.WorkspaceID) {
		return nil, apperrors.ErrForbidden
	}

	if int(count) >= int(workspace.ApiKeysLimit) {
		return nil, apperrors.ErrForbidden
	}

	expiryTime, err := utils.CalculateExpiry(params.ExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry: %w", err)
	}

	rawKey, hashedKey, hint := utils.GenerateAPIKey(env.Name)

	apiKey, err := a.repo.Create(ctx, a.mapToCreateParams(params, expiryTime, hashedKey, hint))
	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return mapToCreateAPIKeyResponse(rawKey, apiKey), nil

}

func (a *apiKeyService) ListAPIKeys(ctx context.Context, params ListApiKeyParams) ([]dtos.APIKeyInfo, error) {

	arg := sqlc.ListAPIKeysByWorkspaceAndEnvParams{
		WorkspaceID:   params.WorkspaceID,
		EnvironmentID: params.EnvironmentID,
	}

	rows, err := a.repo.List(ctx, arg)

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

	args := sqlc.RevokeAPIKeyParams{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
	}

	row, err := a.repo.Revoke(ctx, args)
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

	rowsAffected, err := a.repo.Delete(ctx, sqlc.DeleteAPIKeyParams{
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

	validatedKey, err := a.repo.ValidateAndTouch(ctx, hashedKey)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	isTest := strings.HasPrefix(rawKey, "ne_test_")
	return &ValidatedKey{
		ID:          validatedKey.ID,
		WorkspaceID: validatedKey.WorkspaceID,
		EnvID:       validatedKey.EnvironmentID,
		IsTest:      isTest,
	}, nil
}

func mapToCreateAPIKeyResponse(rawKey string, row sqlc.ApiKey) *dtos.CreateAPIKeyResponse {
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

func (a *apiKeyService) mapToCreateParams(params CreateAPIKeyParams, expiry *time.Time, keyHash, hint string) sqlc.CreateApiKeyParams {
	var expiresAt pgtype.Timestamptz
	if expiry != nil {
		expiresAt = pgtype.Timestamptz{Time: *expiry, Valid: true}
	}

	return sqlc.CreateApiKeyParams{
		WorkspaceID:   params.WorkspaceID,
		EnvironmentID: params.EnvironmentID,
		Label:         params.Label,
		CreatedBy:     params.UserID,
		ExpiresAt:     expiresAt,
		KeyHash:       keyHash,
		KeyHint:       hint,
	}
}

func mapToAPIKeyInfo(row sqlc.ListAPIKeysByWorkspaceAndEnvRow) dtos.APIKeyInfo {
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
