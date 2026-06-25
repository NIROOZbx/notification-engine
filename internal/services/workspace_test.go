package services

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveMember(t *testing.T) {
	callerID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	otherUserID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	wspID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}

	tests := []struct {
		name          string
		params        RemoveMemberParams
		mockRole      string
		mockCount     int64
		mockErr       error
		expectedError error
	}{{
		name: "Success: Owner removes a Member",
		params: RemoveMemberParams{
			WorkspaceID:  wspID,
			CallerID:     callerID,
			TargetUserID: otherUserID,
			CallerRole:   "owner",
		},
		mockRole:      "member",
		mockCount:     1,
		expectedError: nil,
	}, {
		name:          "Fail: Admin tries to remove an Owner",
		params:        RemoveMemberParams{CallerID: callerID, CallerRole: "admin", TargetUserID: otherUserID, WorkspaceID: wspID},
		mockRole:      "owner",
		expectedError: apperrors.ErrForbidden,
	},
		{
			name:          "Fail: Admin tries to remove another Admin",
			params:        RemoveMemberParams{WorkspaceID: wspID, CallerID: callerID, CallerRole: "admin", TargetUserID: otherUserID},
			mockRole:      "admin",
			expectedError: apperrors.ErrForbidden,
		},
		{
			name:          "Fail: Removing the Last Owner",
			params:        RemoveMemberParams{CallerID: callerID, CallerRole: "owner", TargetUserID: otherUserID, WorkspaceID: wspID},
			mockRole:      "owner",
			mockCount:     1,
			expectedError: apperrors.ErrForbidden,
		},
		{
			name: "Fail: User not in workspace (404)",
			params: RemoveMemberParams{
				CallerID:     callerID,
				CallerRole:   "owner",
				TargetUserID: otherUserID,
				WorkspaceID:  wspID,
			},
			mockErr:       pgx.ErrNoRows,
			expectedError: apperrors.ErrNotFound,
		},
		
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mRepo := new(mocks.WorkspaceRepository)

			svc := NewWorkSpaceService(mRepo,nil)

			mRepo.On("GetMemberRole", mock.Anything, sqlc.GetMemberRoleParams{
				WorkspaceID: wspID,
				UserID:      otherUserID,
			}).Return(test.mockRole, test.mockErr)


			if test.expectedError == nil {
				mRepo.On("DeleteWorkspaceMember", mock.Anything, sqlc.DeleteWorkspaceMemberParams{
					WorkspaceID: wspID,
					UserID:      otherUserID,
				}).Return(nil)
			}

			err := svc.RemoveMember(context.Background(), test.params)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceService_Delete(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	tests := []struct {
		name          string
		workspaceID   pgtype.UUID
		mockSetup     func(m *mocks.WorkspaceRepository)
		expectedError error
	}{
		{
			name:        "Success: Delete Workspace",
			workspaceID: wspID,
			mockSetup: func(m *mocks.WorkspaceRepository) {
				m.On("DeleteWorkspace", mock.Anything, wspID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "Fail: DB Error",
			workspaceID: wspID,
			mockSetup: func(m *mocks.WorkspaceRepository) {
				m.On("DeleteWorkspace", mock.Anything, wspID).Return(pgx.ErrTxClosed) // random DB error
			},
			expectedError: pgx.ErrTxClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.WorkspaceRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewWorkSpaceService(mRepo, nil) // no billing client needed for delete
			err := svc.Delete(context.Background(), tt.workspaceID)
			if tt.expectedError != nil {
				assert.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceService_UpdateMemberRole(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	callerID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	targetID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}

	tests := []struct {
		name          string
		params        UpdateMemberRoleParams
		mockSetup     func(m *mocks.WorkspaceRepository)
		expectedError error
	}{
		{
			name: "Fail: Caller is Member",
			params: UpdateMemberRoleParams{
				WorkspaceID:  wspID,
				CallerRole:   "member",
				CallerID:     callerID,
				TargetUserID: targetID,
				Role:         "admin",
			},
			expectedError: apperrors.ErrForbidden,
		},
		{
			name: "Fail: Assign Owner Role",
			params: UpdateMemberRoleParams{
				WorkspaceID:  wspID,
				CallerRole:   "owner",
				CallerID:     callerID,
				TargetUserID: targetID,
				Role:         "owner",
			},
			expectedError: apperrors.ErrForbidden,
		},
		{
			name: "Success: Owner makes Member an Admin",
			params: UpdateMemberRoleParams{
				WorkspaceID:  wspID,
				CallerRole:   "owner",
				CallerID:     callerID,
				TargetUserID: targetID,
				Role:         "admin",
			},
			mockSetup: func(m *mocks.WorkspaceRepository) {
				m.On("GetMemberRole", mock.Anything, sqlc.GetMemberRoleParams{
					WorkspaceID: wspID,
					UserID:      targetID,
				}).Return("member", nil)

				m.On("UpdateMemberRole", mock.Anything, sqlc.UpdateMemberRoleParams{
					WorkspaceID: wspID,
					UserID:      targetID,
					Role:        "admin",
				}).Return(sqlc.UpdateMemberRoleRow{
					WorkspaceID: wspID,
					UserID:      targetID,
					Role:        "admin",
				}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.WorkspaceRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewWorkSpaceService(mRepo, nil)
			_, err := svc.UpdateMemberRole(context.Background(), tt.params)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}
