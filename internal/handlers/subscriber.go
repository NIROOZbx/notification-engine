package handlers

import (
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type SubscriberHandler struct {
	svc services.SubscriberService
	log zerolog.Logger
}

func NewSubscriberHandler(svc services.SubscriberService, log zerolog.Logger) *SubscriberHandler {
	return &SubscriberHandler{
		svc: svc,
		log: log,
	}
}

func (h *SubscriberHandler) Identify(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	var req dtos.IdentifyRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	subscriber, err := h.svc.Identify(c.Context(), services.IdentifySubscriberInput{
		WorkspaceID:    utils.UUIDToString(workspaceID),
		EnvironmentID:  utils.UUIDToString(envID),
		ExternalUserID: req.ExternalUserID,
		Channel:        req.Channel,
		ContactValue:   req.ContactValue,
		Metadata:       req.Metadata,
	})

	if err != nil {
		h.log.Error().Err(err).Str("external_user_id", req.ExternalUserID).Msg("failed to identify user")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "user identified successfully", toSubscriberResponse(subscriber))
}

func (h *SubscriberHandler) UpsertPreference(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	var req dtos.UpsertPreferenceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	pref, err := h.svc.UpsertPreference(c.Context(), services.UpsertPreferenceInput{
		WorkspaceID:    utils.UUIDToString(workspaceID),
		EnvironmentID:  utils.UUIDToString(envID),
		ExternalUserID: req.ExternalUserID,
		Channel:        req.Channel,
		EventType:      req.EventType,
		IsEnabled:      req.IsEnabled,
	})

	if err != nil {
		h.log.Error().Err(err).Str("external_user_id", req.ExternalUserID).Msg("failed to upsert preference")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "preference updated successfully", toUserPreferenceResponse(pref))
}

func (h *SubscriberHandler) Delete(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid subscriber id")
	}

	err = h.svc.Delete(c.Context(), utils.UUIDToString(id), utils.UUIDToString(workspaceID))
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("failed to delete subscriber")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "subscriber deleted successfully", nil)
}

func (h *SubscriberHandler) List(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	subscribers, err := h.svc.List(c.Context(), utils.UUIDToString(workspaceID), utils.UUIDToString(envID))
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list subscribers")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.SubscriberResponse, len(subscribers))
	for i, s := range subscribers {
		resp[i] = toSubscriberResponse(s)
	}

	return response.OK(c, "subscribers fetched successfully", resp)
}

func toSubscriberResponse(s *domain.Subscriber) dtos.SubscriberResponse {
	return dtos.SubscriberResponse{
		ID:             s.ID,
		WorkspaceID:    s.WorkspaceID,
		EnvironmentID:  s.EnvironmentID,
		ExternalUserID: s.ExternalUserID,
		Channel:        s.Channel,
		ContactValue:   s.ContactValue,
		IsVerified:     s.IsVerified,
		Metadata:       s.Metadata,
		CreatedAt:      s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toUserPreferenceResponse(p *domain.UserPreference) dtos.UserPreferenceResponse {
	return dtos.UserPreferenceResponse{
		ID:             p.ID,
		SubscriberID:   p.SubscriberID,
		ExternalUserID: p.ExternalUserID,
		Channel:        p.Channel,
		EventType:      p.EventType,
		IsEnabled:      p.IsEnabled,
		CreatedAt:      p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
