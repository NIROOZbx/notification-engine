package services

import (
	"context"
	"fmt"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
)

type PlanResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	MembersLimit      int32  `json:"members_limit"`
	ApiKeysLimit     int32  `json:"api_keys_limit"`
	LogRetentionDays  int32  `json:"log_retention_days"`
	PriceCents        int32  `json:"price_cents"`
	ExternalPriceID   string `json:"external_price_id"`
	EmailLimit        int32  `json:"email_limit"`
	SMSLimit          int32  `json:"sms_limit"`
	PushLimit         int32  `json:"push_limit"`
	SlackLimit        int32  `json:"slack_limit"`
	WhatsAppLimit     int32  `json:"whatsapp_limit"`
	WebhookLimit      int32  `json:"webhook_limit"`
	InAppLimit        int32  `json:"in_app_limit"`
	OriginalPriceCents int32 `json:"original_price"`
}

type PlanService interface {
	GetAllPlans(ctx context.Context) ([]PlanResponse, error)
}

type planService struct {
	repo repositories.PlanRepository
}

func NewPlanService(repo repositories.PlanRepository) *planService {
	return &planService{repo: repo}
}

func (s *planService) GetAllPlans(ctx context.Context) ([]PlanResponse, error) {
	plans, err := s.repo.GetAllPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plans: %w", err)
	}

	var resp []PlanResponse
	for _, p := range plans {
		resp = append(resp, PlanResponse{
			ID:               p.ID.String(),
			Name:             p.Name,
			MembersLimit:     p.MembersLimit,
			ApiKeysLimit:    p.ApiKeysLimit,
			LogRetentionDays: p.LogRetentionDays,
			PriceCents:       p.PriceCents,
			ExternalPriceID: p.ExternalPriceID.String,
			EmailLimit:      p.EmailLimitMonth,
			SMSLimit:        p.SmsLimitMonth,
			PushLimit:       p.PushLimitMonth,
			SlackLimit:      p.SlackLimitMonth,
			WhatsAppLimit:   p.WhatsappLimitMonth,
			WebhookLimit:    p.WebhookLimitMonth,
			InAppLimit:      p.InAppLimitMonth,
			OriginalPriceCents: p.OriginalPriceCents,
		})
	}
	return resp, nil
}
