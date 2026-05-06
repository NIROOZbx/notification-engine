package repositories

import (
	"context"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceRepository interface {
	GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.WorkspaceMember, error)
	GetWorkspaceMemberWithDetailsByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.GetWorkspaceMemberWithDetailsByUserIDRow, error)
	FindWorkspaceByID(ctx context.Context, id pgtype.UUID) (sqlc.Workspace, error)
	CreateWorkspace(ctx context.Context,arg sqlc.CreateWorkspaceParams) (sqlc.Workspace, error)
	CreateWorkspaceMember(ctx context.Context, arg sqlc.CreateWorkspaceMemberParams) (sqlc.WorkspaceMember, error)
	CreateEnvironment(ctx context.Context, arg sqlc.CreateEnvironmentParams) error
	GetWorkspaceBySlug(ctx context.Context, slug string) (sqlc.Workspace, error)
	UpdateWorkspaceName(ctx context.Context, arg sqlc.UpdateWorkspaceNameParams) (sqlc.Workspace, error)
	DeleteWorkspace(ctx context.Context, id pgtype.UUID) error

	GetWorkspaceMembers(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.GetWorkspaceMembersWithDetailsRow, error)
	GetMemberRole(ctx context.Context, arg sqlc.GetMemberRoleParams) (string, error)
	UpdateMemberRole(ctx context.Context, arg sqlc.UpdateMemberRoleParams) (sqlc.UpdateMemberRoleRow, error)
	DeleteWorkspaceMember(ctx context.Context, arg sqlc.DeleteWorkspaceMemberParams) error
	GetOwnerCount(ctx context.Context,workspaceID pgtype.UUID)(int64,error)
    GetWorkspaceWithPlan(ctx context.Context, id pgtype.UUID) (sqlc.GetWorkspaceWithPlanRow, error)
	GetEnvironmentsByWorkspace(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.Environment, error)
	GetWorkspaceWithEnvironmentsBySlug(ctx context.Context, slug string) ([]sqlc.GetWorkspaceWithEnvironmentsBySlugRow, error)

	Atomic(ctx context.Context, fn func(repo WorkspaceRepository) error) error
}

type workspaceRepo struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewWorkspaceRepository(queries *sqlc.Queries, pool *pgxpool.Pool) WorkspaceRepository {
	return &workspaceRepo{queries: queries, pool: pool}
}

func (r *workspaceRepo) GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.WorkspaceMember, error) {
	return r.queries.GetWorkspaceMemberByUserID(ctx, userID)
}

func (r *workspaceRepo) GetWorkspaceMemberWithDetailsByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.GetWorkspaceMemberWithDetailsByUserIDRow, error) {
	return r.queries.GetWorkspaceMemberWithDetailsByUserID(ctx, userID)
}
func (r *workspaceRepo) GetWorkspaceWithPlan(ctx context.Context, id pgtype.UUID) (sqlc.GetWorkspaceWithPlanRow, error) {
	return r.queries.GetWorkspaceWithPlan(ctx, id)
}


func (r *workspaceRepo) FindWorkspaceByID(ctx context.Context, id pgtype.UUID) (sqlc.Workspace, error) {
	return r.queries.FindWorkspaceByID(ctx, id)
}

func (r *workspaceRepo) Atomic(ctx context.Context, fn func(repo WorkspaceRepository) error) error {
	return pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		qtx := r.queries.WithTx(tx)

		workspaceRepo := workspaceRepo{
			queries: qtx,
			pool:    r.pool,
		}

		return fn(&workspaceRepo)
	})
}

func (r *workspaceRepo) CreateWorkspace(ctx context.Context,arg sqlc.CreateWorkspaceParams) (sqlc.Workspace, error) {
	return r.queries.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		Name: arg.Name,
		Slug: arg.Slug,
	})

}

func (r *workspaceRepo) CreateWorkspaceMember(ctx context.Context, arg sqlc.CreateWorkspaceMemberParams) (sqlc.WorkspaceMember, error) {
	return r.queries.CreateWorkspaceMember(ctx, arg)
}

func (r *workspaceRepo) CreateEnvironment(ctx context.Context, arg sqlc.CreateEnvironmentParams) error {
	return r.queries.CreateEnvironment(ctx, arg)
}

func (r *workspaceRepo) GetWorkspaceBySlug(ctx context.Context, slug string) (sqlc.Workspace, error) {
	return r.queries.GetWorkspaceBySlug(ctx, slug)
}

func (r *workspaceRepo) UpdateWorkspaceName(ctx context.Context, arg sqlc.UpdateWorkspaceNameParams) (sqlc.Workspace, error) {
	return r.queries.UpdateWorkspaceName(ctx, arg)
}

func (r *workspaceRepo) DeleteWorkspace(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteWorkspace(ctx, id)
}

func (r *workspaceRepo) GetWorkspaceMembers(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.GetWorkspaceMembersWithDetailsRow, error) {
	return r.queries.GetWorkspaceMembersWithDetails(ctx, workspaceID)
}

func (r *workspaceRepo) GetMemberRole(ctx context.Context, arg sqlc.GetMemberRoleParams) (string, error) {
	return r.queries.GetMemberRole(ctx, arg)
}

func (r *workspaceRepo) UpdateMemberRole(ctx context.Context, arg sqlc.UpdateMemberRoleParams) (sqlc.UpdateMemberRoleRow, error) {
	return r.queries.UpdateMemberRole(ctx, arg)
}

func (r *workspaceRepo) DeleteWorkspaceMember(ctx context.Context, arg sqlc.DeleteWorkspaceMemberParams) error {
	return r.queries.DeleteWorkspaceMember(ctx, arg)
}

func (r *workspaceRepo)GetOwnerCount(ctx context.Context,workspaceID pgtype.UUID)(int64,error){

	return r.queries.CountOwners(ctx,workspaceID)
}

func (r *workspaceRepo) GetEnvironmentsByWorkspace(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.Environment, error) {
	return r.queries.GetEnvironmentsByWorkspace(ctx, workspaceID)
}

func (r *workspaceRepo) GetWorkspaceWithEnvironmentsBySlug(ctx context.Context, slug string) ([]sqlc.GetWorkspaceWithEnvironmentsBySlugRow, error) {
	return r.queries.GetWorkspaceWithEnvironmentsBySlug(ctx, slug)
}