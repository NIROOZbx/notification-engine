package services

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validWorkspaceID   = "00000000-0000-0000-0000-000000000001"
	validEnvironmentID = "00000000-0000-0000-0000-000000000002"
	validSubscriberID  = "00000000-0000-0000-0000-000000000003"
)

func TestSubscriberService_Identify(t *testing.T) {
	tests := []struct {
		name          string
		input         IdentifySubscriberInput
		mockSetup     func(m *mocks.SubscriberRepo)
		expectedError string
	}{
		{
			name: "Success: Valid Contacts",
			input: IdentifySubscriberInput{
				WorkspaceID:    validWorkspaceID,
				EnvironmentID:  validEnvironmentID,
				ExternalUserID: "user_123",
				Contacts: []ContactInput{
					{Channel: "email", ContactValue: "test@example.com"},
				},
				Metadata: map[string]any{"plan": "premium"},
			},
			mockSetup: func(m *mocks.SubscriberRepo) {
				m.On("UpsertSubscriber", mock.Anything, mock.Anything).Return(&domain.Subscriber{
					ID:             validSubscriberID,
					ExternalUserID: "user_123",
				}, nil)
			},
			expectedError: "",
		},
		{
			name: "Fail: Invalid Channel",
			input: IdentifySubscriberInput{
				WorkspaceID:    validWorkspaceID,
				EnvironmentID:  validEnvironmentID,
				ExternalUserID: "user_123",
				Contacts: []ContactInput{
					{Channel: "carrier_pigeon", ContactValue: "roof"},
				},
			},
			mockSetup:     nil,
			expectedError: "invalid channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.SubscriberRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewSubscriberService(mRepo)

			subs, err := svc.Identify(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, subs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, subs)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriberService_UpsertPreference(t *testing.T) {
	tests := []struct {
		name          string
		input         UpsertPreferenceInput
		mockSetup     func(m *mocks.SubscriberRepo)
		expectedError string
	}{
		{
			name: "Success: Valid Preference",
			input: UpsertPreferenceInput{
				WorkspaceID:    validWorkspaceID,
				EnvironmentID:  validEnvironmentID,
				ExternalUserID: "user_123",
				Channel:        "email",
				EventType:      "marketing",
				IsEnabled:      false,
			},
			mockSetup: func(m *mocks.SubscriberRepo) {
				// Mock finding the subscriber first
				m.On("GetSubscriberByExternalIDAndChannel", mock.Anything, validWorkspaceID, validEnvironmentID, "user_123", "email").
					Return(&domain.Subscriber{ID: validSubscriberID}, nil)

				// Mock upserting the preference
				m.On("UpsertPreference", mock.Anything, mock.Anything).
					Return(&domain.UserPreference{Channel: "email", IsEnabled: false}, nil)
			},
			expectedError: "",
		},
		{
			name: "Fail: Subscriber Not Found",
			input: UpsertPreferenceInput{
				WorkspaceID:    validWorkspaceID,
				EnvironmentID:  validEnvironmentID,
				ExternalUserID: "ghost_user",
				Channel:        "email",
			},
			mockSetup: func(m *mocks.SubscriberRepo) {
				m.On("GetSubscriberByExternalIDAndChannel", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, apperrors.ErrNotFound)
			},
			expectedError: "subscriber not found for channel email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.SubscriberRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewSubscriberService(mRepo)

			pref, err := svc.UpsertPreference(context.Background(), tt.input)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, pref)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pref)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriberService_List(t *testing.T) {
	tests := []struct {
		name          string
		workspaceID   string
		envID         string
		page          int32
		pageSize      int32
		mockSetup     func(m *mocks.SubscriberRepo)
		expectedError string
	}{
		{
			name:        "Success: Valid bounds",
			workspaceID: validWorkspaceID,
			envID:       validEnvironmentID,
			page:        2,
			pageSize:    10,
			mockSetup: func(m *mocks.SubscriberRepo) {
				m.On("ListSubscribers", mock.Anything, mock.Anything, mock.Anything, int32(10), int32(10)).
					Return([]*domain.Subscriber{}, nil)
				m.On("CountSubscribers", mock.Anything, mock.Anything, mock.Anything).
					Return(int64(25), nil)
			},
			expectedError: "",
		},
		{
			name:        "Success: Out of bounds correction",
			workspaceID: validWorkspaceID,
			envID:       validEnvironmentID,
			page:        -5,   // should become 1
			pageSize:    5000, // should become 20
			mockSetup: func(m *mocks.SubscriberRepo) {
				// Offset for page 1 is 0.
				m.On("ListSubscribers", mock.Anything, mock.Anything, mock.Anything, int32(20), int32(0)).
					Return([]*domain.Subscriber{}, nil)
				m.On("CountSubscribers", mock.Anything, mock.Anything, mock.Anything).
					Return(int64(5), nil)
			},
			expectedError: "",
		},
		{
			name:          "Fail: Invalid Workspace UUID",
			workspaceID:   "not-a-uuid",
			envID:         validEnvironmentID,
			page:          1,
			pageSize:      10,
			mockSetup:     nil,
			expectedError: "invalid input: workspace id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.SubscriberRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewSubscriberService(mRepo)

			list, err := svc.List(context.Background(), tt.workspaceID, tt.envID, tt.page, tt.pageSize)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, list)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, list)
			}
			mRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriberService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		workspaceID   string
		mockSetup     func(m *mocks.SubscriberRepo)
		expectedError string
	}{
		{
			name:        "Success: Delete",
			id:          validSubscriberID,
			workspaceID: validWorkspaceID,
			mockSetup: func(m *mocks.SubscriberRepo) {
				m.On("DeleteSubscriber", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: "",
		},
		{
			name:          "Fail: Invalid ID",
			id:            "invalid",
			workspaceID:   validWorkspaceID,
			mockSetup:     nil,
			expectedError: "invalid input: subscriber id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := new(mocks.SubscriberRepo)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo)
			}
			svc := NewSubscriberService(mRepo)

			err := svc.Delete(context.Background(), tt.id, tt.workspaceID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mRepo.AssertExpectations(t)
		})
	}
}
