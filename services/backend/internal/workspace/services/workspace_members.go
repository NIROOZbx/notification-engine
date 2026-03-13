package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func (w *workspaceService) GetMembers(ctx context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error) {
	members, err := w.repo.GetWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("getting members: %w", err)
	}
	return members, nil
}

func (w *workspaceService) GetMemberRole(ctx context.Context, workspaceID, userID pgtype.UUID) (string, error) {
	params := db.GetMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}
	role, err := w.repo.GetMemberRole(ctx, params)
	if err != nil {
		return "", fmt.Errorf("getting member role: %w", err)
	}
	return role, nil
}

func (w *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID pgtype.UUID, role string) (*db.WorkspaceMember, error) {

	params := db.UpdateMemberRoleParams{
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
	err := w.repo.DeleteWorkspaceMember(ctx, db.DeleteWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}
