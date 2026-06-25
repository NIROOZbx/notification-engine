package services

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		params        CreateUser
		mockSetup     func(mRepo *mocks.UserRepository)
		expectedError error
	}{
		{
			name: "Success: Create User",
			params: CreateUser{
				Email:        "test@example.com",
				FullName:     "Test User",
				PasswordHash: "hashed_password",
			},
			mockSetup: func(mRepo *mocks.UserRepository) {
				mRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(p sqlc.CreateUserParams) bool {
					return p.Email == "test@example.com" && p.FullName == "Test User" && p.PasswordHash.String == "hashed_password"
				})).Return(sqlc.User{Email: "test@example.com"}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.UserRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewUserService(mRepo)
			res, err := svc.CreateUser(context.Background(), tt.params)

			assert.ErrorIs(t, err, tt.expectedError)
			if tt.expectedError == nil {
				assert.NotNil(t, res)
				assert.Equal(t, tt.params.Email, res.Email)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpsertOAuthUser(t *testing.T) {
	tests := []struct {
		name          string
		params        UpsertOAuthInput
		mockSetup     func(mRepo *mocks.UserRepository)
		expectedError string
	}{
		{
			name: "Success: Upsert User",
			params: UpsertOAuthInput{
				Email:      "oauth@example.com",
				FullName:   "OAuth User",
				AvatarURL:  "http://example.com/avatar.png",
				Provider:   "google",
				ProviderID: "123456789",
			},
			mockSetup: func(mRepo *mocks.UserRepository) {
				mRepo.On("UpsertOAuthUser", mock.Anything, mock.MatchedBy(func(p sqlc.UpsertOAuthUserParams) bool {
					return p.Email == "oauth@example.com" && p.AuthProvider == "google" && p.ProviderID.String == "123456789"
				})).Return(sqlc.User{Email: "oauth@example.com"}, nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.UserRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewUserService(mRepo)
			res, err := svc.UpsertOAuthUser(context.Background(), tt.params)

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

func TestUserService_MarkUserAsVerified(t *testing.T) {
	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

	tests := []struct {
		name          string
		mockSetup     func(mRepo *mocks.UserRepository)
		expectedError string
	}{
		{
			name: "Success: Mark Verified",
			mockSetup: func(mRepo *mocks.UserRepository) {
				mRepo.On("MarkUserAsVerified", mock.Anything, userID).Return(sqlc.User{IsVerified: true}, nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.UserRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewUserService(mRepo)
			res, err := svc.MarkUserAsVerified(context.Background(), userID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.True(t, res.IsVerified)
			}
			mRepo.AssertExpectations(t)
		})
	}
}
