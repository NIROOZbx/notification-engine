package repositories

import (
	"context"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
)

type PlanRepository interface {
	GetAllPlans(ctx context.Context) ([]sqlc.Plan, error)
}

type planRepo struct {
	queries *sqlc.Queries
}

func NewPlanRepository(queries *sqlc.Queries) PlanRepository {
	return &planRepo{queries: queries}
}

func (r *planRepo) GetAllPlans(ctx context.Context) ([]sqlc.Plan, error) {
	return r.queries.GetAllPlans(ctx)
}
