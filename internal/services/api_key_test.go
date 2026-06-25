package services

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAPIKey(t *testing.T) {
	// Dummy UUIDs for testing
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	envID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	keyID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}

	// 1. The Table Definition
	tests := []struct {
		name          string
		rawKey        string
		mockSetup     func(m *mocks.APIKeyRepository, hashedKey string) // How the DB mock should behave
		expectedError error
		expectSuccess bool
	}{
		{
			name:   "Success: Valid Live Key",
			rawKey: "ne_live_1234567890",
			mockSetup: func(m *mocks.APIKeyRepository, hashedKey string) {
				m.On("ValidateAndTouch", mock.Anything, hashedKey).Return(sqlc.ApiKey{
					ID:            keyID,
					WorkspaceID:   wspID,
					EnvironmentID: envID,
				}, nil)
			},
			expectSuccess: true,
		},
		{
			name:          "Fail: Invalid Prefix",
			rawKey:        "bad_prefix_123456",
			mockSetup:     func(m *mocks.APIKeyRepository, hashedKey string) {}, // No DB call expected
			expectedError: apperrors.ErrForbidden,
		},
		{
			name:   "Fail: Key Not Found or Revoked",
			rawKey: "ne_test_notfoundkey",
			mockSetup: func(m *mocks.APIKeyRepository, hashedKey string) {
				m.On("ValidateAndTouch", mock.Anything, hashedKey).Return(sqlc.ApiKey{}, pgx.ErrNoRows)
			},
			expectedError: apperrors.ErrUnauthorized,
		},
	}

	// 2. The Execution Loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.APIKeyRepository)
			noopLogger := zerolog.Nop() // Dummy logger for tests
			svc := NewAPIKeyService(mockRepo, noopLogger)

			// Setup expectations if a mockSetup func is provided
			if tt.mockSetup != nil {
				hashedKey := utils.HashAPIKey(tt.rawKey)
				tt.mockSetup(mockRepo, hashedKey)
			}

			// Act
			result, err := svc.ValidateAPIKey(context.Background(), tt.rawKey)

			// Assert
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, wspID, result.WorkspaceID)
				assert.Equal(t, envID, result.EnvID)
			}

			// Verify that all expected mock calls were actually made
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateAPIKey(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	envID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	otherWspID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}

	tests := []struct {
		name          string
		params        CreateAPIKeyParams
		mockSetup     func(m *mocks.APIKeyRepository)
		expectedError error
	}{
		{
			name: "Success: Create Key",
			params: CreateAPIKeyParams{
				WorkspaceID:   wspID,
				EnvironmentID: envID,
				UserID:        userID,
				Label:         "Test Key",
				ExpiresIn:     30,
			},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("GetEnvironment", mock.Anything, envID).Return(sqlc.Environment{WorkspaceID: wspID, Name: "production"}, nil)
				m.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{ApiKeysLimit: 5}, nil)
				m.On("CountActiveKeys", mock.Anything, wspID).Return(int64(2), nil)
				m.On("Create", mock.Anything, mock.Anything).Return(sqlc.ApiKey{ID: pgtype.UUID{Valid: true}}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Environment does not belong to Workspace",
			params: CreateAPIKeyParams{
				WorkspaceID:   wspID,
				EnvironmentID: envID,
			},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("GetEnvironment", mock.Anything, envID).Return(sqlc.Environment{WorkspaceID: otherWspID}, nil)
				m.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{ApiKeysLimit: 5}, nil)
				m.On("CountActiveKeys", mock.Anything, wspID).Return(int64(2), nil)
			},
			expectedError: apperrors.ErrForbidden,
		},
		{
			name: "Fail: Limit Reached",
			params: CreateAPIKeyParams{
				WorkspaceID:   wspID,
				EnvironmentID: envID,
			},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("GetEnvironment", mock.Anything, envID).Return(sqlc.Environment{WorkspaceID: wspID}, nil)
				m.On("GetWorkspaceWithPlan", mock.Anything, wspID).Return(sqlc.GetWorkspaceWithPlanRow{ApiKeysLimit: 5}, nil)
				m.On("CountActiveKeys", mock.Anything, wspID).Return(int64(5), nil)
			},
			expectedError: apperrors.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.APIKeyRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			noopLogger := zerolog.Nop()
			svc := NewAPIKeyService(mRepo, noopLogger)

			res, err := svc.CreateAPIKey(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestRevokeAPIKey(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	keyID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		params        RevokeKeyParams
		mockSetup     func(m *mocks.APIKeyRepository)
		expectedError error
	}{
		{
			name: "Success: Revoke Key",
			params: RevokeKeyParams{ID: keyID, WorkspaceID: wspID},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("Revoke", mock.Anything, sqlc.RevokeAPIKeyParams{
					ID:          keyID,
					WorkspaceID: wspID,
				}).Return(sqlc.RevokeAPIKeyRow{ID: keyID, IsRevoked: true}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Not Found",
			params: RevokeKeyParams{ID: keyID, WorkspaceID: wspID},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("Revoke", mock.Anything, mock.Anything).Return(sqlc.RevokeAPIKeyRow{}, pgx.ErrNoRows)
			},
			expectedError: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.APIKeyRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewAPIKeyService(mRepo, zerolog.Nop())

			res, err := svc.RevokeAPIKey(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	wspID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	keyID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	tests := []struct {
		name          string
		params        DeleteKeyParams
		mockSetup     func(m *mocks.APIKeyRepository)
		expectedError error
	}{
		{
			name: "Success: Delete Key",
			params: DeleteKeyParams{ID: keyID, WorkspaceID: wspID},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("Delete", mock.Anything, sqlc.DeleteAPIKeyParams{
					ID:          keyID,
					WorkspaceID: wspID,
				}).Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Not Found",
			params: DeleteKeyParams{ID: keyID, WorkspaceID: wspID},
			mockSetup: func(m *mocks.APIKeyRepository) {
				m.On("Delete", mock.Anything, mock.Anything).Return(int64(0), nil)
			},
			expectedError: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.APIKeyRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewAPIKeyService(mRepo, zerolog.Nop())

			err := svc.DeleteAPIKey(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}
