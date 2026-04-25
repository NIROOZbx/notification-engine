package handlers

import (

	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type PlanHandler struct {
	planSvc services.PlanService
}

func NewPlanHandler(planSvc services.PlanService) *PlanHandler {
	return &PlanHandler{planSvc: planSvc}
}

func (h *PlanHandler) GetAllPlans(c fiber.Ctx) {
	plans, err := h.planSvc.GetAllPlans(c.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.OK(c,"sucesfully fetched",plans)
}
