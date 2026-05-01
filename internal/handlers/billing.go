package handlers

import (
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type BillingHandler struct {
	billingSvc services.BillingService
	userSvc    services.UserService
	log        zerolog.Logger
}

func NewBillingHandler(billingSvc services.BillingService, userSvc services.UserService, log zerolog.Logger) *BillingHandler {
	return &BillingHandler{
		billingSvc: billingSvc,
		userSvc:    userSvc,
		log:        log,
	}
}

func (h *BillingHandler) GetUsage(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	usage, err := h.billingSvc.GetUsage(c.Context(), utils.UUIDToString(workspaceID), utils.UUIDToString(envID))
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	h.log.Info().
		Str("workspace_id", utils.UUIDToString(workspaceID)).
		Str("env_id", utils.UUIDToString(envID)).
		Msg("Workspace usage retrieved")

	return response.OK(c, "Fetched usage successfully", usage)
}

func (h *BillingHandler) GetSubscription(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	sub, err := h.billingSvc.GetSubscription(c.Context(), utils.UUIDToString(workspaceID))
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "Fetched subscription successfully", sub)
}

func (h *BillingHandler) CancelSubscription(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	subscriptionID := c.Query("id")

	if subscriptionID == "" {
		return response.BadRequest(c, nil, "Subscription ID is required")
	}

	h.log.Info().
		Str("workspace_id", utils.UUIDToString(workspaceID)).
		Str("subscription_id", subscriptionID).
		Msg("Attempting to cancel subscription")

	err = h.billingSvc.CancelSubscription(c.Context(), utils.UUIDToString(workspaceID), subscriptionID)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	h.log.Info().
		Str("workspace_id", utils.UUIDToString(workspaceID)).
		Str("subscription_id", subscriptionID).
		Msg("Subscription cancelled successfully")

	return response.OK(c, "Subscription cancelled successfully", nil)
}

func (h *BillingHandler) CreateCheckout(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	userID, err := utils.GetUID(c)
	if err != nil {
		return response.Unauthorized(c, "missing user id")
	}
	userDetails, err := h.userSvc.GetFullUserDetails(c.Context(), userID)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	planID := c.Query("plan_id")
	if planID == "" {
		return response.BadRequest(c, nil, "Plan ID is required")
	}

	url, err := h.billingSvc.CreateCheckoutSession(c.Context(), utils.UUIDToString(workspaceID), planID, userDetails.Email)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	h.log.Info().
		Str("workspace_id", utils.UUIDToString(workspaceID)).
		Str("plan_id", planID).
		Str("email", userDetails.Email).
		Msg("Checkout session created")

	return response.OK(c, "Checkout link generated", url)
}

func (h *BillingHandler) GetCheckoutSession(c fiber.Ctx) error {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		return response.BadRequest(c, nil, "session_id is required")
	}

	details, err := h.billingSvc.GetCheckoutSession(c.Context(), sessionID)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "Checkout session details retrieved", details)
}