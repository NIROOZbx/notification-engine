package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceWithRole struct {
	Workspace *db.Workspace
	Role      string
}

type WorkspaceService interface {
	GetOrCreate(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.Workspace, error)
	GetBySlug(ctx context.Context, slug string) (*db.Workspace, error)
	UpdateName(ctx context.Context, id pgtype.UUID, name string) (*db.Workspace, error)
	Delete(ctx context.Context, id pgtype.UUID) error

	GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error)
	GetMemberRole(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID) (string, error)
	UpdateMemberRole(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID, role string) (*db.WorkspaceMember, error)
	RemoveMember(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID) error
}

type workspaceService struct {
	repo *db.Queries
	pool *pgxpool.Pool
}

func (w *workspaceService) GetOrCreate(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error) {
	var result *WorkspaceWithRole

	err := pgx.BeginTxFunc(ctx, w.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		qtx := w.repo.WithTx(tx)

		member, err := qtx.GetWorkspaceMemberByUserID(ctx, userID)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("checking membership: %w", err)
		}

		if err == nil {
			existingWorkspace, findErr := qtx.FindWorkspaceByID(ctx, member.WorkspaceID)
			if findErr != nil {
				return fmt.Errorf("fetching workspace: %w", findErr)
			}
			result = &WorkspaceWithRole{
				Workspace: &existingWorkspace,
				Role:      member.Role,
			}
			return nil
		}
		slug := utils.Slugify(name)

		args := db.CreateWorkspaceParams{
			Name: name,
			Slug: slug,
		}

		workspace, err := qtx.CreateWorkspace(ctx, args)
		if err != nil {
			return fmt.Errorf("creating workspace: %w", err)
		}

		args2 := db.CreateWorkspaceMemberParams{
			WorkspaceID: workspace.ID,
			UserID:      userID,
			Role:        "owner",
		}

		workspaceMember, err := qtx.CreateWorkspaceMember(ctx, args2)
		if err != nil {
			return fmt.Errorf("creating workspace member: %w", err)
		}

		result = &WorkspaceWithRole{
			Workspace: &workspace,
			Role:      workspaceMember.Role,
		}
		params := db.CreateEnvironmentParams{
			WorkspaceID: workspace.ID,
			Name:        "development",
		}
		params2 := db.CreateEnvironmentParams{
			WorkspaceID: workspace.ID,
			Name:        "production",
		}

		err = qtx.CreateEnvironment(ctx, params)

		if err != nil {
			return fmt.Errorf("creating development environment: %w", err)
		}
		err = qtx.CreateEnvironment(ctx, params2)
		if err != nil {
			return fmt.Errorf("creating production environment: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (w *workspaceService) GetByID(ctx context.Context, id pgtype.UUID) (*db.Workspace, error) {
	workspace, err := w.repo.FindWorkspaceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding workspace: %w", err)
	}
	return &workspace, nil
}

func (w *workspaceService) GetBySlug(ctx context.Context, slug string) (*db.Workspace, error) {
	workspace, err := w.repo.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("finding workspace by slug: %w", err)
	}
	return &workspace, nil
}

func (w *workspaceService) UpdateName(ctx context.Context, id pgtype.UUID, name string) (*db.Workspace, error) {
	workspace, err := w.repo.UpdateWorkspaceName(ctx, db.UpdateWorkspaceNameParams{
		ID:   id,
		Name: name,
		Slug: utils.Slugify(name),
	})
	if err != nil {
		return nil, fmt.Errorf("updating workspace name: %w", err)
	}
	return &workspace, nil
}

func (w *workspaceService) Delete(ctx context.Context, id pgtype.UUID) error {
	if err := w.repo.DeleteWorkspace(ctx, id); err != nil {
		return fmt.Errorf("deleting workspace: %w", err)
	}
	return nil
}

func NewService(repo *db.Queries, pool *pgxpool.Pool) WorkspaceService {
	return &workspaceService{
		repo: repo,
		pool: pool,
	}
}
