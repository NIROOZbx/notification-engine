package repositories

import (
	"context"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	GetUserWithWorkspace(ctx context.Context, id pgtype.UUID) (sqlc.GetUserWithWorkspaceRow, error)
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	UpsertOAuthUser(ctx context.Context, arg sqlc.UpsertOAuthUserParams) (sqlc.User, error)
	FindUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error)
	FindUserByProviderID(ctx context.Context, arg sqlc.FindUserByProviderIDParams) (sqlc.User, error)
	FindUserByEmail(ctx context.Context, email string) (sqlc.User, error)
}

type userRepo struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) UserRepository {
	return &userRepo{queries: queries}
}

func (r *userRepo) GetUserWithWorkspace(ctx context.Context, id pgtype.UUID) (sqlc.GetUserWithWorkspaceRow, error) {
	return r.queries.GetUserWithWorkspace(ctx, id)
}

func (r *userRepo) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *userRepo) UpsertOAuthUser(ctx context.Context, arg sqlc.UpsertOAuthUserParams) (sqlc.User, error) {
	return r.queries.UpsertOAuthUser(ctx, arg)
}

func (r *userRepo) FindUserByID(ctx context.Context, id pgtype.UUID) (sqlc.User, error) {
	return r.queries.FindUserByID(ctx, id)
}

func (r *userRepo) FindUserByProviderID(ctx context.Context, arg sqlc.FindUserByProviderIDParams) (sqlc.User, error) {
	return r.queries.FindUserByProviderID(ctx, arg)
}

func (r *userRepo) FindUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.FindUserByEmail(ctx, email)
}
