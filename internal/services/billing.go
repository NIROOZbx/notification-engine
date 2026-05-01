package services

import (
	"context"

	"github.com/NIROOZbx/notification-engine/internal/billing"
)

type BillingService interface {
	GetUsage(ctx context.Context, workspaceID, environmentID string) (*billing.UsageResponse, error)
	GetSubscription(ctx context.Context, workspaceID string) (*billing.SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, workspaceID, subscriptionID string) error
	CreateCheckoutSession(ctx context.Context, workspaceID, planID, customerEmail string) (string, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*billing.CheckoutSessionDetails, error)
}

type billingService struct {
	billingClient billing.Client
}

func NewBillingService(billingClient billing.Client) *billingService {
	return &billingService{
		billingClient: billingClient,
	}
}

func (s *billingService) GetUsage(ctx context.Context, workspaceID, envID string) (*billing.UsageResponse, error) {
	return s.billingClient.GetUsage(ctx, workspaceID, envID)
}

func (s *billingService) GetSubscription(ctx context.Context, workspaceID string) (*billing.SubscriptionResponse, error) {
	return s.billingClient.GetSubscription(ctx, workspaceID)
}

func (s *billingService) CancelSubscription(ctx context.Context, workspaceID, subID string) error {
	return s.billingClient.CancelSubscription(ctx, workspaceID, subID)
}

func (s *billingService) CreateCheckoutSession(ctx context.Context, workspaceID, planID,customerEmail string) (string, error) {
	return s.billingClient.CreateCheckoutSession(ctx, workspaceID, planID,customerEmail)
}

func (s *billingService) GetCheckoutSession(ctx context.Context, sessionID string) (*billing.CheckoutSessionDetails, error) {
	return s.billingClient.GetCheckoutSession(ctx, sessionID)
}