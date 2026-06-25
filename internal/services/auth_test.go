package services_test

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/services/mocks"
	sessionMocks "github.com/NIROOZbx/notification-engine/internal/session/mocks"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		req           dtos.RegisterRequest
		mockSetup     func(mu *mocks.UserService)
		expectedError error
	}{
		{
			name: "Success: Register new user",
			req: dtos.RegisterRequest{
				Email:    "test@example.com",
				Name:     "John Doe",
				Password: "password123",
			},
			mockSetup: func(mu *mocks.UserService) {
				// We use mock.MatchedBy to verify the correct data is passed to services.CreateUser
				// without caring about the exact password hash string (which changes every time).
				mu.On("CreateUser", mock.Anything, mock.MatchedBy(func(p services.CreateUser) bool {
					return p.Email == "test@example.com" && p.FullName == "John Doe" && p.PasswordHash != ""
				})).Return(&sqlc.User{ID: pgtype.UUID{Valid: true}}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Duplicate Email",
			req: dtos.RegisterRequest{
				Email:    "duplicate@example.com",
				Password: "password123",
			},
			mockSetup: func(mu *mocks.UserService) {
				// We mock a generic error that satisfies the IsUniqueViolation check (or return the exact apperror)
				mu.On("CreateUser", mock.Anything, mock.Anything).
					Return(nil, apperrors.NewAlreadyExistsError("email"))
			},
			expectedError: apperrors.NewAlreadyExistsError("email"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(mocks.UserService)
			mw := new(mocks.WorkspaceService)
			ms := new(sessionMocks.Store)

			if tt.mockSetup != nil {
				tt.mockSetup(mu)
			}

			// Dummy config
			cfg := &config.AuthConfig{}
			svc := services.NewAuthService(cfg, mu, mw, ms)

			user, err := svc.Register(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.ErrorContains(t, err, tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
			mu.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		req           dtos.LoginRequest
		mockSetup     func(mu *mocks.UserService, ms *sessionMocks.Store)
		expectedError error
	}{
		{
			name: "Success: Valid Login",
			req: dtos.LoginRequest{
				Email:    "login@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(mu *mocks.UserService, ms *sessionMocks.Store) {
				hashed, _ := helpers.HashPassword("correctpassword")
				
				// Mock finding the user
				mu.On("GetAuthContextByEmail", mock.Anything, "login@example.com").Return(&sqlc.GetAuthContextByEmailRow{
					User: sqlc.User{
						ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
						IsVerified:   true,
						PasswordHash: pgtype.Text{String: hashed, Valid: true},
					},
				}, nil)

				// Mock getting the token version from the store
				ms.On("GetTokenVersion", mock.Anything, mock.Anything).Return(int64(1), nil)

				// Mock saving the refresh token in the redis session store
				ms.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Fail: Unverified Email",
			req: dtos.LoginRequest{
				Email:    "unverified@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(mu *mocks.UserService, ms *sessionMocks.Store) {
				hashed, _ := helpers.HashPassword("correctpassword")
				mu.On("GetAuthContextByEmail", mock.Anything, "unverified@example.com").Return(&sqlc.GetAuthContextByEmailRow{
					User: sqlc.User{
						IsVerified: false, // This triggers the ErrNotVerified check
						PasswordHash: pgtype.Text{String: hashed, Valid: true},
					},
				}, nil)
			},
			expectedError: apperrors.ErrNotVerified,
		},
		{
			name: "Fail: Wrong Password",
			req: dtos.LoginRequest{
				Email:    "login@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mu *mocks.UserService, ms *sessionMocks.Store) {
				hashed, _ := helpers.HashPassword("correctpassword")
				mu.On("GetAuthContextByEmail", mock.Anything, "login@example.com").Return(&sqlc.GetAuthContextByEmailRow{
					User: sqlc.User{
						IsVerified: true,
						PasswordHash: pgtype.Text{String: hashed, Valid: true},
					},
				}, nil)
			},
			expectedError: apperrors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := new(mocks.UserService)
			mw := new(mocks.WorkspaceService)
			ms := new(sessionMocks.Store)

			if tt.mockSetup != nil {
				tt.mockSetup(mu, ms)
			}

			// Provide dummy secrets to prevent nil pointer panics when building JWTs
			cfg := &config.AuthConfig{
				AccessTokenSecret:  "secret",
				RefreshTokenSecret: "secret",
			}
			svc := services.NewAuthService(cfg, mu, mw, ms)

			res, tokens, err := svc.Login(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, res)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotNil(t, tokens)
			}
			mu.AssertExpectations(t)
			ms.AssertExpectations(t)
		})
	}
}
