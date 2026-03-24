package repositories

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type APIKeyRepository interface {
    GetEnvironment(ctx context.Context, id pgtype.UUID) (sqlc.Environment, error)
    GetWorkspaceWithPlan(ctx context.Context, id pgtype.UUID) (sqlc.GetWorkspaceWithPlanRow, error)
    CountActiveKeys(ctx context.Context, workspaceID pgtype.UUID) (int64, error)
    Create(ctx context.Context, arg sqlc.CreateApiKeyParams) (sqlc.ApiKey, error)
    List(ctx context.Context, arg sqlc.ListAPIKeysByWorkspaceAndEnvParams) ([]sqlc.ListAPIKeysByWorkspaceAndEnvRow, error)
    Revoke(ctx context.Context, arg sqlc.RevokeAPIKeyParams) (sqlc.RevokeAPIKeyRow, error)
    Delete(ctx context.Context, arg sqlc.DeleteAPIKeyParams) (int64, error)
    ValidateAndTouch(ctx context.Context, hash string) (sqlc.ValidateAndTouchAPIKeyRow, error)
}

type apiKeyRepo struct {
    queries *sqlc.Queries
}

func NewAPIKeyRepository(queries *sqlc.Queries) APIKeyRepository {
    return &apiKeyRepo{queries: queries}
}

func (r *apiKeyRepo) GetEnvironment(ctx context.Context, id pgtype.UUID) (sqlc.Environment, error) {
	return r.queries.GetEnvironmentByID(ctx, id)
}

func (r *apiKeyRepo) GetWorkspaceWithPlan(ctx context.Context, id pgtype.UUID) (sqlc.GetWorkspaceWithPlanRow, error) {
	return r.queries.GetWorkspaceWithPlan(ctx, id)
}

func (r *apiKeyRepo) CountActiveKeys(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
	return r.queries.CountActiveAPIKeys(ctx, workspaceID)
}

func (r *apiKeyRepo) Create(ctx context.Context, arg sqlc.CreateApiKeyParams) (sqlc.ApiKey, error) {
	return r.queries.CreateApiKey(ctx, arg)
}

func (r *apiKeyRepo) List(ctx context.Context, arg sqlc.ListAPIKeysByWorkspaceAndEnvParams) ([]sqlc.ListAPIKeysByWorkspaceAndEnvRow, error) {
	return r.queries.ListAPIKeysByWorkspaceAndEnv(ctx, arg)
}

func (r *apiKeyRepo) Revoke(ctx context.Context, arg sqlc.RevokeAPIKeyParams) (sqlc.RevokeAPIKeyRow, error) {
	return r.queries.RevokeAPIKey(ctx, arg)
}

func (r *apiKeyRepo) Delete(ctx context.Context, arg sqlc.DeleteAPIKeyParams) (int64, error) {
	return r.queries.DeleteAPIKey(ctx, arg)
}

func (r *apiKeyRepo) ValidateAndTouch(ctx context.Context, hash string) (sqlc.ValidateAndTouchAPIKeyRow, error) {
	return r.queries.ValidateAndTouchAPIKey(ctx, hash)
}