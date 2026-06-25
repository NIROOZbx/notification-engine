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

func TestTemplateService_Create(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	layoutID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		params        domain.CreateTemplateParams
		mockSetup     func(mRepo *mocks.TemplateRepository, mWs *mocks.WorkspaceRepository, mLayout *mocks.LayoutRepo)
		expectedError error
	}{
		{
			name: "Success: Under limit with layout",
			params: domain.CreateTemplateParams{
				WorkspaceID: wspID,
				LayoutID:    layoutID,
				Name:        "Promo Template",
			},
			mockSetup: func(mRepo *mocks.TemplateRepository, mWs *mocks.WorkspaceRepository, mLayout *mocks.LayoutRepo) {
				mWs.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{MaxTemplates: 10}, nil)
				mRepo.On("CountTemplates", mock.Anything, wspID).Return(int64(5), nil)
				mRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.Template{ID: pgtype.UUID{Valid: true}}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Success: Uses Default Layout when not provided",
			params: domain.CreateTemplateParams{
				WorkspaceID: wspID,
				Name:        "No Layout Template",
			},
			mockSetup: func(mRepo *mocks.TemplateRepository, mWs *mocks.WorkspaceRepository, mLayout *mocks.LayoutRepo) {
				mWs.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{MaxTemplates: 10}, nil)
				mRepo.On("CountTemplates", mock.Anything, wspID).Return(int64(5), nil)
				mLayout.On("GetDefaultLayout", mock.Anything, wspID).Return(&domain.Layout{ID: layoutID}, nil)
				// Create should be called with the LayoutID populated
				mRepo.On("Create", mock.Anything, mock.MatchedBy(func(p domain.CreateTemplateParams) bool {
					return p.LayoutID == layoutID
				})).Return(&domain.Template{ID: pgtype.UUID{Valid: true}}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Plan Limit Reached",
			params: domain.CreateTemplateParams{
				WorkspaceID: wspID,
			},
			mockSetup: func(mRepo *mocks.TemplateRepository, mWs *mocks.WorkspaceRepository, mLayout *mocks.LayoutRepo) {
				mWs.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{MaxTemplates: 5}, nil)
				mRepo.On("CountTemplates", mock.Anything, wspID).Return(int64(5), nil)
			},
			expectedError: apperrors.ErrLimitReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.TemplateRepository)
			mWs := new(mocks.WorkspaceRepository)
			mLayout := new(mocks.LayoutRepo)
			mConfig := new(mocks.ChannelConfigRepo)

			if tt.mockSetup != nil {
				tt.mockSetup(mRepo, mWs, mLayout)
			}
			svc := NewTemplateService(mRepo, mLayout, mWs, mConfig)

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
			mLayout.AssertExpectations(t)
		})
	}
}

func TestTemplateService_Update(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	templateID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	magicStr := "magic"
	liveStr := "live"
	draftStr := "draft"
	newName := "New Name"

	tests := []struct {
		name          string
		params        domain.UpdateTemplateParams
		mockSetup     func(mRepo *mocks.TemplateRepository)
		expectedError string
	}{
		{
			name: "Fail: Invalid Status String",
			params: domain.UpdateTemplateParams{
				ID:          templateID,
				WorkspaceID: wspID,
				Status:      &magicStr,
			},
			mockSetup:     nil,
			expectedError: "invalid status magic",
		},
		{
			name: "Fail: Go live with no active channels",
			params: domain.UpdateTemplateParams{
				ID:          templateID,
				WorkspaceID: wspID,
				Status:      &liveStr,
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "draft"}, nil)
				mRepo.On("HasActiveChannels", mock.Anything, templateID).Return(false, nil)
			},
			expectedError: apperrors.ErrNoActiveChannels.Error(),
		},
		{
			name: "Fail: Update dropped template without recovering",
			params: domain.UpdateTemplateParams{
				ID:          templateID,
				WorkspaceID: wspID,
				Name:        &newName, // trying to just change name while dropped
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "dropped"}, nil)
				mRepo.On("HasActiveChannels", mock.Anything, templateID).Return(true, nil)
			},
			expectedError: "cannot update dropped template unless restoring status",
		},
		{
			name: "Success: Recover dropped template",
			params: domain.UpdateTemplateParams{
				ID:          templateID,
				WorkspaceID: wspID,
				Status:      &draftStr,
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "dropped"}, nil)
				mRepo.On("HasActiveChannels", mock.Anything, templateID).Return(true, nil)
				mRepo.On("Update", mock.Anything, mock.Anything).Return(&domain.Template{Status: "draft"}, nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.TemplateRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewTemplateService(mRepo, nil, nil, nil)

			res, err := svc.Update(context.Background(), tt.params)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestTemplateService_Delete(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	templateID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		mockSetup     func(mRepo *mocks.TemplateRepository)
		expectedError string
	}{
		{
			name: "Fail: Cannot delete live template",
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "live"}, nil)
			},
			expectedError: "cannot delete a live template",
		},
		{
			name: "Success: Delete draft template",
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "draft"}, nil)
				mRepo.On("Delete", mock.Anything, templateID, wspID).Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.TemplateRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewTemplateService(mRepo, nil, nil, nil)
			err := svc.Delete(context.Background(), templateID, wspID)
			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestTemplateService_UpdateChannel(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	templateID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	channelID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	otherChannelID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}

	falseBool := false

	tests := []struct {
		name          string
		params        domain.UpdateTemplateChannelParams
		mockSetup     func(mRepo *mocks.TemplateRepository)
		expectedError string
	}{
		{
			name: "Fail: Modifying dropped template",
			params: domain.UpdateTemplateChannelParams{
				ID:          channelID,
				TemplateID:  templateID,
				WorkspaceID: wspID,
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetChannelByID", mock.Anything, channelID).Return(&domain.TemplateChannel{TemplateID: templateID}, nil)
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "dropped"}, nil)
			},
			expectedError: "cannot modify channels of a dropped template",
		},
		{
			name: "Fail: Disabling last active channel on live template",
			params: domain.UpdateTemplateChannelParams{
				ID:          channelID,
				TemplateID:  templateID,
				WorkspaceID: wspID,
				IsActive:    &falseBool,
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetChannelByID", mock.Anything, channelID).Return(&domain.TemplateChannel{TemplateID: templateID}, nil)
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "live"}, nil)
				
				// ListChannels returns only this channel (and maybe inactive ones)
				mRepo.On("ListChannels", mock.Anything, templateID).Return([]*domain.TemplateChannel{
					{ID: channelID, IsActive: true},
					{ID: otherChannelID, IsActive: false},
				}, nil)
			},
			expectedError: "cannot disable the last active channel of a live template",
		},
		{
			name: "Success: Update channel on draft template",
			params: domain.UpdateTemplateChannelParams{
				ID:          channelID,
				TemplateID:  templateID,
				WorkspaceID: wspID,
				Content:     map[string]any{"text": "hello"},
			},
			mockSetup: func(mRepo *mocks.TemplateRepository) {
				mRepo.On("GetChannelByID", mock.Anything, channelID).Return(&domain.TemplateChannel{TemplateID: templateID}, nil)
				mRepo.On("GetByID", mock.Anything, templateID, wspID).Return(&domain.Template{Status: "draft"}, nil)
				mRepo.On("UpdateChannel", mock.Anything, mock.Anything).Return(&domain.TemplateChannel{ID: channelID}, nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.TemplateRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewTemplateService(mRepo, nil, nil, nil)

			res, err := svc.UpdateChannel(context.Background(), tt.params)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
			mRepo.AssertExpectations(t)
		})
	}
}
