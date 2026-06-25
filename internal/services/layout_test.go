package services

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLayoutService_Create(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

	tests := []struct {
		name          string
		params        domain.CreateLayoutParams
		mockSetup     func(mRepo *mocks.LayoutRepo, mWs *mocks.WorkspaceRepository)
		expectedError error
	}{
		{
			name: "Success: Under Limit",
			params: domain.CreateLayoutParams{
				WorkspaceID: wspID,
				Name:        "Base Layout",
			},
			mockSetup: func(mRepo *mocks.LayoutRepo, mWs *mocks.WorkspaceRepository) {
				mWs.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{MaxLayouts: 5}, nil)
				mRepo.On("CountLayouts", mock.Anything, wspID).Return(int64(2), nil)
				mRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.Layout{ID: pgtype.UUID{Valid: true}}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Limit Reached",
			params: domain.CreateLayoutParams{
				WorkspaceID: wspID,
			},
			mockSetup: func(mRepo *mocks.LayoutRepo, mWs *mocks.WorkspaceRepository) {
				mWs.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{MaxLayouts: 5}, nil)
				mRepo.On("CountLayouts", mock.Anything, wspID).Return(int64(5), nil)
			},
			expectedError: apperrors.ErrLimitReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.LayoutRepo)
			mWs := new(mocks.WorkspaceRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo, mWs)
			}
			svc := NewLayoutService(mRepo, mWs)
			res, err := svc.Create(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
			mRepo.AssertExpectations(t)
			mWs.AssertExpectations(t)
		})
	}
}

func TestLayoutService_Delete(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	layoutID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		mockSetup     func(mRepo *mocks.LayoutRepo)
		expectedError string
	}{
		{
			name: "Success: Delete Non-Default Layout",
			mockSetup: func(mRepo *mocks.LayoutRepo) {
				mRepo.On("GetLayoutByID", mock.Anything, layoutID, wspID).Return(&domain.Layout{IsDefault: false}, nil)
				mRepo.On("DeleteLayout", mock.Anything, layoutID, wspID).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Fail: Delete Default Layout",
			mockSetup: func(mRepo *mocks.LayoutRepo) {
				mRepo.On("GetLayoutByID", mock.Anything, layoutID, wspID).Return(&domain.Layout{IsDefault: true}, nil)
			},
			expectedError: "cannot delete a default layout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.LayoutRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewLayoutService(mRepo, nil)
			err := svc.Delete(context.Background(), layoutID, wspID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestLayoutService_SetDefault(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	layoutID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		mockSetup     func(mRepo *mocks.LayoutRepo)
		expectedError error
	}{
		{
			name: "Success: Set Default",
			mockSetup: func(mRepo *mocks.LayoutRepo) {
				mRepo.On("SetLayoutDefault", mock.Anything, layoutID, wspID).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.LayoutRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewLayoutService(mRepo, nil)
			err := svc.SetDefault(context.Background(), layoutID, wspID)

			assert.ErrorIs(t, err, tt.expectedError)
			mRepo.AssertExpectations(t)
		})
	}
}
