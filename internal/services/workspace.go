package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type WorkspaceWithRole struct {
	Workspace *dtos.WorkspaceResponse
	Role      string
}

type UpdateMemberRoleParams struct {
	WorkspaceID  pgtype.UUID
	CallerRole   string
	CallerID     pgtype.UUID
	TargetUserID pgtype.UUID
	Role         string
}

type RemoveMemberParams struct {
	WorkspaceID  pgtype.UUID
	CallerID     pgtype.UUID
	CallerRole   string
	TargetUserID pgtype.UUID
}

type WorkspaceService interface {
	GetOrCreate(ctx context.Context, userID pgtype.UUID, name string) (*WorkspaceWithRole, error)
	GetByID(ctx context.Context, workspaceID pgtype.UUID) (*dtos.WorkspaceResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dtos.WorkspaceResponse, error)
	UpdateName(ctx context.Context, workspaceID pgtype.UUID, name string) (*dtos.WorkspaceResponse, error)
	Delete(ctx context.Context, workspaceID pgtype.UUID) error

	GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]dtos.WorkspaceMemberResponse, error)
	GetMemberRole(ctx context.Context, workspaceID pgtype.UUID, userID pgtype.UUID) (string, error)
	UpdateMemberRole(ctx context.Context, params UpdateMemberRoleParams) (*dtos.WorkspaceMemberResponse, error)
	RemoveMember(ctx context.Context, params RemoveMemberParams) error
	GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (*dtos.WorkspaceMemberResponse, error)
}

type workspaceService struct {
	repo repositories.WorkspaceRepository
}

func NewWorkSpaceService(repo repositories.WorkspaceRepository) WorkspaceService {
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
	existingWorkspace, findErr := w.repo.GetWorkspaceWithPlan(ctx, workspaceID)
	if findErr != nil {
		return nil, fmt.Errorf("fetching workspace: %w", findErr)
	}
	result := &WorkspaceWithRole{
		Workspace: mapToWorkspaceResponseFromRow(existingWorkspace),
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
			Workspace: mapToWorkspaceResponse(workspace, "Free"),
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

func (w *workspaceService) GetByID(ctx context.Context, workspaceID pgtype.UUID) (*dtos.WorkspaceResponse, error) {
	workspace, err := w.repo.GetWorkspaceWithPlan(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("finding workspace: %w", err)
	}
	return mapToWorkspaceResponseFromRow(workspace), nil
}

func (w *workspaceService) GetBySlug(ctx context.Context, slug string) (*dtos.WorkspaceResponse, error) {
	workspace, err := w.repo.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("finding workspace by slug: %w", err)
	}
	return mapToWorkspaceResponse(workspace, ""), nil
}

func (w *workspaceService) UpdateName(ctx context.Context, workspaceID pgtype.UUID, name string) (*dtos.WorkspaceResponse, error) {
	workspace, err := w.repo.UpdateWorkspaceName(ctx, sqlc.UpdateWorkspaceNameParams{
		ID:   workspaceID,
		Name: name,
		Slug: utils.Slugify(name),
	})
	if err != nil {
		return nil, fmt.Errorf("updating workspace name: %w", err)
	}
	return mapToWorkspaceResponse(workspace, ""), nil
}

func (w *workspaceService) Delete(ctx context.Context, workspaceID pgtype.UUID) error {
	if err := w.repo.DeleteWorkspace(ctx, workspaceID); err != nil {
		return fmt.Errorf("deleting workspace: %w", err)
	}
	return nil
}

func (w *workspaceService) GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]dtos.WorkspaceMemberResponse, error) {
	rows, err := w.repo.GetWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("getting members: %w", err)
	}

	return mapToMemberResponses(rows), nil
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

func (w *workspaceService) UpdateMemberRole(ctx context.Context, params UpdateMemberRoleParams) (*dtos.WorkspaceMemberResponse, error) {
	roleToAssign := strings.ToLower(params.Role)
	if roleToAssign != "admin" && roleToAssign != "member" {
		return nil, fmt.Errorf("invalid role: %w", apperrors.ErrBadRequest)
	}

	if params.CallerRole != "owner" && params.CallerRole != "admin" {
		return nil, apperrors.ErrForbidden
	}
	if utils.UUIDToString(params.CallerID) == utils.UUIDToString(params.TargetUserID) {
		return nil, fmt.Errorf("self-update: %w", apperrors.ErrForbidden)
	}
	if params.Role == "owner" {
		return nil, fmt.Errorf("cannot assign owner role: %w", apperrors.ErrForbidden)
	}

	currentTargetRole, err := w.repo.GetMemberRole(ctx, sqlc.GetMemberRoleParams{
		WorkspaceID: params.WorkspaceID,
		UserID:      params.TargetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("target user not found in workspace: %w", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("fetching target role: %w", err)
	}

	if currentTargetRole == "owner" && params.CallerRole != "owner" {
		return nil, apperrors.ErrForbidden
	}
	if currentTargetRole == "admin" && params.CallerRole != "owner" {
		return nil, fmt.Errorf("only owners can modify an admin's role: %w", apperrors.ErrForbidden)
	}

	params2 := sqlc.UpdateMemberRoleParams{
		WorkspaceID: params.WorkspaceID,
		UserID:      params.TargetUserID,
		Role:        params.Role,
	}
	_, err = w.repo.UpdateMemberRole(ctx, params2)

	if err != nil {
		return nil, fmt.Errorf("updating member role: %w", err)
	}

	// Fetch detailed record to map to DTO
	rows, err := w.repo.GetWorkspaceMembers(ctx, params.WorkspaceID)
	if err != nil {
		return nil, err
	}

	for _, m := range rows {
		if m.UserID == params.TargetUserID {
			resp := mapToMemberResponse(m)
			return &resp, nil
		}
	}

	return nil, fmt.Errorf("updated member not found")
}

func (w *workspaceService) GetWorkspaceMemberByUserID(ctx context.Context, userID pgtype.UUID) (*dtos.WorkspaceMemberResponse, error) {
	member, err := w.repo.GetWorkspaceMemberByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	rows, err := w.repo.GetWorkspaceMembers(ctx, member.WorkspaceID)
	if err != nil {
		return nil, err
	}

	for _, m := range rows {
		if m.UserID == userID {
			resp := mapToMemberResponse(m)
			return &resp, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (w *workspaceService) RemoveMember(ctx context.Context, params RemoveMemberParams) error {
	isSelf := utils.UUIDToString(params.CallerID) == utils.UUIDToString(params.TargetUserID)

	if isSelf {
		return fmt.Errorf("self-deletion not allowed in this endpoint: %w", apperrors.ErrForbidden)
	}

	if params.CallerRole != "owner" && params.CallerRole != "admin" {
		return apperrors.ErrForbidden
	}

	targetUserRole, err := w.repo.GetMemberRole(ctx, sqlc.GetMemberRoleParams{
		WorkspaceID: params.WorkspaceID,
		UserID:      params.TargetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not in workspace: %w", apperrors.ErrNotFound)
		}
		return err
	}

	if params.CallerRole == "owner" && targetUserRole == "owner" {
		return apperrors.ErrForbidden
	}

	if params.CallerRole == "admin" && (targetUserRole == "owner" || targetUserRole == "admin") {
		return fmt.Errorf("admins cannot remove peers or owners: %w", apperrors.ErrForbidden)
	}

	if targetUserRole == "owner" {
		ownerCount, countErr := w.repo.GetOwnerCount(ctx, params.WorkspaceID)
		if countErr != nil {
			return fmt.Errorf("checking owner count: %w", countErr)
		}

		if ownerCount <= 1 {
			return fmt.Errorf("cannot remove the last owner: %w", apperrors.ErrForbidden)
		}
	}

	err = w.repo.DeleteWorkspaceMember(ctx, sqlc.DeleteWorkspaceMemberParams{
		WorkspaceID: params.WorkspaceID,
		UserID:      params.TargetUserID,
	})
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}

	return nil
}

func mapToWorkspaceResponse(w sqlc.Workspace, planName string) *dtos.WorkspaceResponse {
	return &dtos.WorkspaceResponse{
		ID:        utils.UUIDToString(w.ID),
		Name:      w.Name,
		Slug:      w.Slug,
		PlanName:  planName,
		CreatedAt: w.CreatedAt.Time,
	}
}

func mapToWorkspaceResponseFromRow(w sqlc.GetWorkspaceWithPlanNameRow) *dtos.WorkspaceResponse {
	return &dtos.WorkspaceResponse{
		ID:        utils.UUIDToString(w.ID),
		Name:      w.Name,
		Slug:      w.Slug,
		PlanName:  w.PlanName,
		CreatedAt: w.CreatedAt.Time,
	}
}

func mapToMemberResponse(m sqlc.GetWorkspaceMembersWithDetailsRow) dtos.WorkspaceMemberResponse {
	return dtos.WorkspaceMemberResponse{
		WorkspaceID: utils.UUIDToString(m.WorkspaceID),
		UserID:      utils.UUIDToString(m.UserID),
		Name:        m.Name,
		Email:       m.Email,
		AvatarURL:   m.AvatarUrl.String,
		Role:        m.Role,
		JoinedAt:    m.JoinedAt.Time,
	}
}

func mapToMemberResponses(members []sqlc.GetWorkspaceMembersWithDetailsRow) []dtos.WorkspaceMemberResponse {
	result := make([]dtos.WorkspaceMemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, mapToMemberResponse(m))
	}
	return result
}
