package billing

import (
	"context"
	"fmt"
	"time"

	billingv1 "github.com/NIROOZbx/notification-engine/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CheckLimitResponse struct {
	Allowed bool
	Reason  string
	Limit   int32
	Current int32
	ResetAt time.Time
}

type SubscriptionResponse struct {
	ID               string
	PlanName         string
	Status           string
	CurrentPeriodEnd time.Time
	PaymentProvider  string
}

type UsageResponse struct {
	WorkspaceID        string    `json:"workspace_id"`
	EnvironmentID      string    `json:"environment_id"`
	EmailCount         int32     `json:"email_count"`
	SMSCount           int32     `json:"sms_count"`
	PushCount          int32     `json:"push_count"`
	SlackCount         int32     `json:"slack_count"`
	WhatsAppCount      int32     `json:"whatsapp_count"`
	WebhookCount       int32     `json:"webhook_count"`
	InAppCount         int32     `json:"in_app_count"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	SubscriptionStatus string    `json:"subscription_status"`
}

type RecordUsageInput struct {
	WorkspaceID     string
	EnvironmentID   string
	ChannelConfigID string
	ChannelName     string
	Provider        string
	Success         bool
}

type Client interface {
	CheckLimit(ctx context.Context, workspaceID, environmentID string, channel string) (*CheckLimitResponse, error)
	RecordUsage(ctx context.Context, input RecordUsageInput) error
	CreateSubscription(ctx context.Context, workspaceID string, planID, paymentProvider string) (string, error)
	GetSubscription(ctx context.Context, workspaceID string) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, workspaceID, subscriptionID string) error
	GetUsage(ctx context.Context, workspaceID, environmentID string) (*UsageResponse, error)
	CreateCheckoutSession(ctx context.Context, workspaceID, planID,customerEmail string) (string, error)
	Close() error
}

type grpcClient struct {
	conn   *grpc.ClientConn
	client billingv1.BillingServiceClient
}

func NewGRPCClient(addr string) (Client, error) {
	fmt.Println("grpc port", addr)
	conn, err := grpc.NewClient("host.docker.internal:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err

	}

	billingClient := billingv1.NewBillingServiceClient(conn)
	return &grpcClient{
		conn:   conn,
		client: billingClient,
	}, nil

}

func (g *grpcClient) CheckLimit(ctx context.Context, workspaceID, environmentID string, channel string) (*CheckLimitResponse, error) {
	limit, err := g.client.CheckLimit(ctx, &billingv1.CheckLimitRequest{
		WorkspaceId:   workspaceID,
		EnvironmentId: environmentID,
		Channel:       channel,
	})

	if err != nil {
		return nil, err
	}
	parsedTime, err := time.Parse(time.RFC3339, limit.ResetAt)
	if err != nil {

		return nil, fmt.Errorf("Error parsing time: %v", err)
	}

	return &CheckLimitResponse{
		Allowed: limit.Allowed,
		Reason:  limit.Reason,
		Limit:   limit.Limit,
		Current: limit.Current,
		ResetAt: parsedTime,
	}, nil
}

func (g *grpcClient) RecordUsage(ctx context.Context, input RecordUsageInput) error {
	_, err := g.client.RecordUsage(ctx, &billingv1.RecordUsageRequest{
		WorkspaceId:     input.WorkspaceID,
		EnvironmentId:   input.EnvironmentID,
		ChannelConfigId: input.ChannelConfigID,
		Channel:         input.ChannelName,
		Provider:        input.Provider,
		Success:         input.Success,
	})
	return err
}

func (g *grpcClient) CreateSubscription(ctx context.Context, workspaceID, planID, paymentProvider string) (string, error) {
	resp, err := g.client.CreateSubscription(ctx, &billingv1.CreateSubscriptionRequest{
		WorkspaceId:     workspaceID,
		PlanId:          planID,
		PaymentProvider: paymentProvider,
	})
	if err != nil {
		return "", err
	}
	return resp.SubscriptionId, nil
}

func (g *grpcClient) GetSubscription(ctx context.Context, workspaceID string) (*SubscriptionResponse, error) {
	resp, err := g.client.GetSubscription(ctx, &billingv1.GetSubscriptionRequest{
		WorkspaceId: workspaceID,
	})
	if err != nil {
		return nil, err
	}

	periodEnd, err := time.Parse(time.RFC3339, resp.CurrentPeriodEnd)
	if err != nil {

		return nil, fmt.Errorf("Error parsing time: %v", err)
	}

	return &SubscriptionResponse{
		ID:               resp.SubscriptionId,
		PlanName:         resp.PlanName,
		Status:           resp.Status,
		CurrentPeriodEnd: periodEnd,
		PaymentProvider:  resp.PaymentProvider,
	}, nil
}

func (g *grpcClient) CancelSubscription(ctx context.Context, workspaceID, subscriptionID string) error {
	_, err := g.client.CancelSubscription(ctx, &billingv1.CancelSubscriptionRequest{
		WorkspaceId:    workspaceID,
		SubscriptionId: subscriptionID,
	})
	return err
}

func (g *grpcClient) GetUsage(ctx context.Context, workspaceID, environmentID string) (*UsageResponse, error) {
	resp, err := g.client.GetUsage(ctx, &billingv1.GetUsageRequest{
		WorkspaceId:   workspaceID,
		EnvironmentId: environmentID,
	})
	if err != nil {
		return nil, err
	}

	periodStart, startErr := time.Parse(time.RFC3339, resp.PeriodStart)
	periodEnd, err := time.Parse(time.RFC3339, resp.PeriodEnd)

	if err != nil || startErr != nil {
		return nil, fmt.Errorf("Error parsing time: %v", err)
	}

	return &UsageResponse{
		WorkspaceID:        resp.WorkspaceId,
		EnvironmentID:      resp.EnvironmentId,
		EmailCount:         resp.EmailCount,
		SMSCount:           resp.SmsCount,
		PushCount:          resp.PushCount,
		SlackCount:         resp.SlackCount,
		WhatsAppCount:      resp.WhatsappCount,
		WebhookCount:       resp.WebhookCount,
		InAppCount:         resp.InAppCount,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		SubscriptionStatus: resp.SubscriptionStatus,
	}, nil
}

func (g *grpcClient) CreateCheckoutSession(ctx context.Context, workspaceID, planID,customerEmail string) (string, error) {
	resp, err := g.client.CreateCheckoutSession(ctx, &billingv1.CreateCheckoutSessionRequest{
		WorkspaceId: workspaceID,
		PlanId:      planID,
		CustomerEmail: customerEmail,
	})
	if err != nil {
		return "", err
	}
	return resp.CheckoutUrl, nil
}

func (g *grpcClient) Close() error {
	return g.conn.Close()
}
