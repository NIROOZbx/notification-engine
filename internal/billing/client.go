package billing

import (
	"context"
	"fmt"
	"time"

	billingv1 "github.com/NIROOZbx/notification-engine/proto"
	"github.com/rs/zerolog"
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

type ChannelUsage struct {
	ChannelName  string `json:"channel_name"`
	CurrentUsage int64  `json:"current_usage"`
}

type UsageResponse struct {
	WorkspaceID        string          `json:"workspace_id"`
	EnvironmentID      string          `json:"environment_id"`
	Usage              []*ChannelUsage `json:"usage"`
	PeriodStart        time.Time       `json:"period_start"`
	PeriodEnd          time.Time       `json:"period_end"`
	SubscriptionStatus string          `json:"subscription_status"`
}

type RecordUsageInput struct {
	WorkspaceID     string
	EnvironmentID   string
	ChannelConfigID string
	ChannelName     string
	Provider        string
	Success         bool
}

type CheckoutSessionDetails struct {
	ID             string `json:"id"`
	CustomerEmail  string `json:"customer_email"`
	AmountTotal    int64  `json:"amount_total"`
	Currency       string `json:"currency"`
	PaymentStatus  string `json:"payment_status"`
	PlanName       string `json:"plan_name"`
	SubscriptionID string `json:"subscription_id"`
}

type Client interface {
	CheckLimit(ctx context.Context, workspaceID, environmentID string, channel string) (*CheckLimitResponse, error)
	RecordUsage(ctx context.Context, input RecordUsageInput) error
	CreateSubscription(ctx context.Context, workspaceID string, planID, paymentProvider string) (string, error)
	GetSubscription(ctx context.Context, workspaceID string) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, workspaceID, subscriptionID string) error
	GetUsage(ctx context.Context, workspaceID, environmentID string) (*UsageResponse, error)
	CreateCheckoutSession(ctx context.Context, workspaceID, planID, customerEmail string) (string, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSessionDetails, error)
	Close() error
}

type grpcClient struct {
	conn   *grpc.ClientConn
	client billingv1.BillingServiceClient
	log zerolog.Logger
}

func NewGRPCClient(addr string,log zerolog.Logger) (Client, error) {
	fmt.Println("grpc port", addr)
	conn, err := grpc.NewClient("host.docker.internal:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err

	}

	billingClient := billingv1.NewBillingServiceClient(conn)
	return &grpcClient{
		conn:   conn,
		client: billingClient,
		log: log,
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
	parsedTime, err := parseTime(limit.ResetAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing reset_at: %w", err)
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

	periodEnd, err := parseTime(resp.CurrentPeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("error parsing current_period_end: %w", err)
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
	g.log.Debug().
    Str("period_start", resp.PeriodStart).
    Str("period_end", resp.PeriodEnd).
    Str("status", resp.SubscriptionStatus).
    Msg("gRPC GetUsage response received")

	periodStart, startErr := parseTime(resp.PeriodStart)
	if startErr != nil {
		return nil, fmt.Errorf("error parsing period_start: %w", startErr)
	}

	periodEnd, endErr := parseTime(resp.PeriodEnd)
	if endErr != nil {
		return nil, fmt.Errorf("error parsing period_end: %w", endErr)
	}

	usageList := make([]*ChannelUsage, len(resp.Usage))
	for i, u := range resp.Usage {
		usageList[i] = &ChannelUsage{
			ChannelName:  u.ChannelName,
			CurrentUsage: u.CurrentUsage,
		}
	}

	return &UsageResponse{
		WorkspaceID:        resp.WorkspaceId,
		EnvironmentID:      resp.EnvironmentId,
		Usage:              usageList,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		SubscriptionStatus: resp.SubscriptionStatus,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
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
func (g *grpcClient) GetCheckoutSession(ctx context.Context, sessionID string) (*CheckoutSessionDetails, error) {
    resp, err := g.client.GetCheckoutSession(ctx, &billingv1.CreateGetSessionRequest{
        SessionId: sessionID,
    })
    if err != nil {
        return nil, err
    }

    return &CheckoutSessionDetails{
        ID:             resp.Id,
        CustomerEmail:  resp.CustomerEmail,
        AmountTotal:    resp.AmountTotal,
        Currency:       resp.Currency,
        PaymentStatus:  resp.PaymentStatus,
        PlanName:       resp.PlanName,
        SubscriptionID: resp.SubscriptionId,
    }, nil
}
