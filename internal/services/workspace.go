package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type WorkspaceWithRole struct {
	Workspace *sqlc.Workspace
	Role      string
}

type WorkspaceService interface {
	GetOrCreate(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Workspace, error)
	GetBySlug(ctx context.Context, slug string) (*sqlc.Workspace, error)
	UpdateName(ctx context.Context, id pgtype.UUID, name string) (*sqlc.Workspace, error)
	Delete(ctx context.Context, id pgtype.UUID) error

	GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.WorkspaceMember, error)
	GetMemberRole(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID) (string, error)
	UpdateMemberRole(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID, role string) (*sqlc.WorkspaceMember, error)
	RemoveMember(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID) error
	GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (*sqlc.WorkspaceMember, error)
}

type workspaceService struct {
	repo repositories.WorkspaceRepository
}

func NewService(repo repositories.WorkspaceRepository) WorkspaceService {
	return &workspaceService{
		repo: repo,
	}
}

func (w *workspaceService) GetOrCreate(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error) {

	member, err := w.repo.GetWorkspaceMemberByUserID(ctx, userID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("checking membership: %w", err)
	}

	if err == nil {
		return w.fetchExistingWorkspace(ctx, member.WorkspaceID, member.Role)
	}

	return w.setupNewWorkspace(ctx, userID, name)

}

func (w *workspaceService) fetchExistingWorkspace(ctx context.Context, workspaceID pgtype.UUID, role string) (*WorkspaceWithRole, error) {
	existingWorkspace, findErr := w.repo.FindWorkspaceByID(ctx, workspaceID)
	if findErr != nil {
		return nil, fmt.Errorf("fetching workspace: %w", findErr)
	}
	result := &WorkspaceWithRole{
		Workspace: &existingWorkspace,
		Role:      role,
	}
	return result, nil

}

func (w *workspaceService) setupNewWorkspace(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error) {

	var result *WorkspaceWithRole
	slug := utils.Slugify(name)

	args := sqlc.CreateWorkspaceParams{
		Name: name,
		Slug: slug,
	}

	err := w.repo.Atomic(ctx, func(repo repositories.WorkspaceRepository) error {
		workspace, err := repo.CreateWorkspace(ctx, args)
		if err != nil {
			return fmt.Errorf("creating workspace: %w", err)
		}
		args2 := sqlc.CreateWorkspaceMemberParams{
			WorkspaceID: workspace.ID,
			UserID:      userID,
			Role:        "owner",
		}

		workspaceMember, err := repo.CreateWorkspaceMember(ctx, args2)
		if err != nil {
			return fmt.Errorf("creating workspace member: %w", err)
		}
		result = &WorkspaceWithRole{
			Workspace: &workspace,
			Role:      workspaceMember.Role,
		}

		defaultEnvs := []string{"development", "production"}

		for _, val := range defaultEnvs {
			err := repo.CreateEnvironment(ctx,
				sqlc.CreateEnvironmentParams{
					WorkspaceID: workspace.ID,
					Name:        val,
				},
			)
			if err != nil {
				return fmt.Errorf("creating %s environment: %w", val, err)
			}
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (w *workspaceService) GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Workspace, error) {
	workspace, err := w.repo.FindWorkspaceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding workspace: %w", err)
	}
	return &workspace, nil
}

func (w *workspaceService) GetBySlug(ctx context.Context, slug string) (*sqlc.Workspace, error) {
	workspace, err := w.repo.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("finding workspace by slug: %w", err)
	}
	return &workspace, nil
}

func (w *workspaceService) UpdateName(ctx context.Context, id pgtype.UUID, name string) (*sqlc.Workspace, error) {
	workspace, err := w.repo.UpdateWorkspaceName(ctx, sqlc.UpdateWorkspaceNameParams{
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

func (w *workspaceService) GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]sqlc.WorkspaceMember, error) {
	members, err := w.repo.GetWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("getting members: %w", err)
	}
	return members, nil
}

func (w *workspaceService) GetMemberRole(ctx context.Context, workspaceID, userID pgtype.UUID) (string, error) {
	params := sqlc.GetMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}
	role, err := w.repo.GetMemberRole(ctx, params)
	if err != nil {
		return "", fmt.Errorf("getting member role: %w", err)
	}
	return role, nil
}

func (w *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID pgtype.UUID, role string) (*sqlc.WorkspaceMember, error) {

	params := sqlc.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        role,
	}
	member, err := w.repo.UpdateMemberRole(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("updating member role: %w", err)
	}
	return &member, nil

}

func (w *workspaceService) RemoveMember(ctx context.Context, workspaceID, userID pgtype.UUID) error {
	err := w.repo.DeleteWorkspaceMember(ctx, sqlc.DeleteWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

func (w *workspaceService) GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (*sqlc.WorkspaceMember, error) {
	workspaceMember, err := w.repo.GetWorkspaceMemberByUserID(ctx, userID)

	if err != nil {
		return nil, err
	}

	return &workspaceMember, nil
}
