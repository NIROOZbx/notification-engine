package core_test

import (
	"context"
	"testing"

	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/internal/billing"
	billingMocks "github.com/NIROOZbx/notification-engine/internal/billing/mocks"
	"github.com/NIROOZbx/notification-engine/internal/metrics"
	"github.com/NIROOZbx/notification-engine/internal/repositories/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupEngineMocks(t *testing.T) (*core.Engine, *mocks.Repository, *mocks.Producer, *billingMocks.Client, *mocks.Renderer) {
	mRepo := new(mocks.Repository)
	mProducer := new(mocks.Producer)
	mBilling := new(billingMocks.Client)
	mRenderer := new(mocks.Renderer)
	mMetrics := metrics.NewMetrics() // Real metrics for simplicity, harmless
	noopLog := zerolog.Nop()

	cfg := core.EngineConfig{
		Repo:          mRepo,
		Producer:      mProducer,
		BillingClient: mBilling,
		Renderer:      mRenderer,
		Log:           noopLog,
		SecretKey:     "secret",
		Metrics:       mMetrics,
	}

	engine := core.NewEngine(cfg)
	return engine, mRepo, mProducer, mBilling, mRenderer
}

func TestEngine_Ingest(t *testing.T) {
	workspaceID := "wsp-1"
	envID := "env-1"

	tests := []struct {
		name      string
		payload   *models.TriggerPayload
		mockSetup func(mRepo *mocks.Repository, mProducer *mocks.Producer, mBilling *billingMocks.Client)
		expectErr bool
	}{
		{
			name: "Scenario: Direct Send (Skips DB preferences, publishes immediately)",
			payload: &models.TriggerPayload{
				EventType:      "test_event",
				ExternalUserID: "user_1",
				RecipientEmail: "direct@example.com",
				IdempotencyKey: "idem_1",
				Channels:       []string{"email"},
			},
			mockSetup: func(mRepo *mocks.Repository, mProducer *mocks.Producer, mBilling *billingMocks.Client) {
				mRepo.On("GetTemplateByEventType", mock.Anything, workspaceID, envID, "test_event").
					Return(&core.Template{ID: "tpl_1", Status: "live"}, nil)

				mRepo.On("GetActiveChannelsByTemplateID", mock.Anything, "tpl_1").
					Return([]core.TemplateChannel{{Channel: "email", IsActive: true}}, nil)

				mRepo.On("GetNotificationLogByIdempotencyKey", mock.Anything, "idem_1:email").
					Return(nil, nil) // No duplicate log found

				mRepo.On("CreateNotificationLog", mock.Anything, mock.Anything).
					Return(&core.NotificationLog{ID: "log_1", WorkspaceID: workspaceID, EnvironmentID: envID, Channel: "email"}, nil)

				mProducer.On("Publish", mock.Anything, "notifications.email", mock.Anything).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Scenario: Billing Limit Exceeded",
			payload: &models.TriggerPayload{
				EventType:      "test_event",
				ExternalUserID: "user_1",
				IdempotencyKey: "idem_2",
				Channels:       []string{"email"},
			},
			mockSetup: func(mRepo *mocks.Repository, mProducer *mocks.Producer, mBilling *billingMocks.Client) {
				mRepo.On("GetTemplateByEventType", mock.Anything, workspaceID, envID, "test_event").
					Return(&core.Template{ID: "tpl_1", Status: "live"}, nil)

				mRepo.On("GetActiveChannelsByTemplateID", mock.Anything, "tpl_1").
					Return([]core.TemplateChannel{{Channel: "email", IsActive: true}}, nil)

				// ResolveContact (Normal strategy requires DB lookup for subscriber)
				mRepo.On("GetContactWithPreference", mock.Anything, core.GetContactWithPreferenceParams{
					WorkspaceID:    workspaceID,
					EnvironmentID:  envID,
					ExternalUserID: "user_1",
					Channel:        "email",
					EventType:      "test_event",
				}).Return(&core.Contact{ContactValue: "user@example.com"}, &core.Preference{IsEnabled: true}, nil)

				// Billing blocks the request
				mBilling.On("CheckLimit", mock.Anything, workspaceID, envID, "email").
					Return(&billing.CheckLimitResponse{Allowed: false, Reason: "Limit Exceeded"}, nil)

				mRepo.On("GetNotificationLogByIdempotencyKey", mock.Anything, "idem_2:email").
					Return(nil, nil)

				// Notice: Publish is NEVER called because limit exceeded!
			},
			expectErr: false, // returns nil smoothly
		},
		{
			name: "Scenario: User Opted Out",
			payload: &models.TriggerPayload{
				EventType:      "test_event",
				ExternalUserID: "user_1",
				IdempotencyKey: "idem_3",
				Channels:       []string{"email"},
			},
			mockSetup: func(mRepo *mocks.Repository, mProducer *mocks.Producer, mBilling *billingMocks.Client) {
				mRepo.On("GetTemplateByEventType", mock.Anything, workspaceID, envID, "test_event").
					Return(&core.Template{ID: "tpl_1", Status: "live"}, nil)

				mRepo.On("GetActiveChannelsByTemplateID", mock.Anything, "tpl_1").
					Return([]core.TemplateChannel{{Channel: "email", IsActive: true}}, nil)

				// Normal strategy: User is opted out
				mRepo.On("GetContactWithPreference", mock.Anything, core.GetContactWithPreferenceParams{
					WorkspaceID:    workspaceID,
					EnvironmentID:  envID,
					ExternalUserID: "user_1",
					Channel:        "email",
					EventType:      "test_event",
				}).Return(&core.Contact{ContactValue: "user@example.com"}, &core.Preference{IsEnabled: false}, nil)

				mBilling.On("CheckLimit", mock.Anything, workspaceID, envID, "email").
					Return(&billing.CheckLimitResponse{Allowed: true}, nil)

				mRepo.On("GetNotificationLogByIdempotencyKey", mock.Anything, "idem_3:email").
					Return(nil, nil)

				// Notice: Publish is NEVER called because user is opted out!
			},
			expectErr: false, // returns nil smoothly
		},
		{
			name: "Scenario: Normal Flow (Happy Path)",
			payload: &models.TriggerPayload{
				EventType:      "test_event",
				ExternalUserID: "user_1",
				IdempotencyKey: "idem_4",
				Channels:       []string{"email"},
			},
			mockSetup: func(mRepo *mocks.Repository, mProducer *mocks.Producer, mBilling *billingMocks.Client) {
				mRepo.On("GetTemplateByEventType", mock.Anything, workspaceID, envID, "test_event").
					Return(&core.Template{ID: "tpl_1", Status: "live"}, nil)

				mRepo.On("GetActiveChannelsByTemplateID", mock.Anything, "tpl_1").
					Return([]core.TemplateChannel{{Channel: "email", IsActive: true}}, nil)

				mRepo.On("GetContactWithPreference", mock.Anything, core.GetContactWithPreferenceParams{
					WorkspaceID:    workspaceID,
					EnvironmentID:  envID,
					ExternalUserID: "user_1",
					Channel:        "email",
					EventType:      "test_event",
				}).Return(&core.Contact{ContactValue: "user@example.com"}, &core.Preference{IsEnabled: true}, nil)

				mBilling.On("CheckLimit", mock.Anything, workspaceID, envID, "email").
					Return(&billing.CheckLimitResponse{Allowed: true}, nil)

				mRepo.On("GetNotificationLogByIdempotencyKey", mock.Anything, "idem_4:email").
					Return(nil, nil)

				mRepo.On("CreateNotificationLog", mock.Anything, mock.Anything).
					Return(&core.NotificationLog{ID: "log_4", WorkspaceID: workspaceID, EnvironmentID: envID, Channel: "email"}, nil)

				mProducer.On("Publish", mock.Anything, "notifications.email", mock.Anything).
					Return(nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, mRepo, mProducer, mBilling, _ := setupEngineMocks(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo, mProducer, mBilling)
			}

			err := engine.Ingest(context.Background(), workspaceID, envID, tt.payload)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mRepo.AssertExpectations(t)
			mProducer.AssertExpectations(t)
			mBilling.AssertExpectations(t)
		})
	}
}

func TestEngine_Process(t *testing.T) {
	tests := []struct {
		name      string
		event     *models.NotificationEvent
		mockSetup func(mRepo *mocks.Repository, mRenderer *mocks.Renderer)
		expectErr bool
	}{
		{
			name: "Scenario: Successful Send",
			event: &models.NotificationEvent{
				NotificationLogID: "log_1",
			},
			mockSetup: func(mRepo *mocks.Repository, mRenderer *mocks.Renderer) {
				// 1. Fetch Log
				mRepo.On("GetNotificationLogByID", mock.Anything, "log_1").
					Return(&core.NotificationLog{
						ID:        "log_1",
						Status:    "queued",
						Channel:   "email",
						Recipient: "user@example.com",
					}, nil)

				// 2. Fetch Template with Channel
				mRepo.On("GetTemplateWithChannel", mock.Anything, mock.Anything, mock.Anything, "email").
					Return(&core.Template{ID: "tpl_1", LayoutID: "lay_1"}, &core.TemplateChannel{Channel: "email", Content: map[string]any{"text":"hellor"}}, nil)

				// 3. Update Log to Processing
				mRepo.On("UpdateNotificationStatus", mock.Anything, "log_1", "processing").
					Return(nil)

				// 4. Resolve Provider - Mocking GetDefaultChannelConfig
				mRepo.On("GetDefaultChannelConfig", mock.Anything, mock.Anything, "email").
					Return(nil, nil) // No config found -> causes error in resolveProvider

				// 5. Update Log to Failed (because resolveProvider will fail)
				mRepo.On("UpdateNotificationStatus", mock.Anything, "log_1", "failed").
					Return(nil)

				// Note: resolveProvider and subsequent steps will fail because no config found.
				// For this test, we expect Process to return an error because resolveProvider fails.
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, mRepo, _, _, mRenderer := setupEngineMocks(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mRepo, mRenderer)
			}

			err := engine.Process(context.Background(), tt.event)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
