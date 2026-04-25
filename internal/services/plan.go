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
		})
	}
	return resp, nil
}
