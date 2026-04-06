package handlers

import (
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
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

type IdentifyRequest struct {
	ExternalUserID string         `json:"external_user_id" validate:"required"`
	Channel        string         `json:"channel"          validate:"required"`
	ContactValue   string         `json:"contact_value"    validate:"required"`
	Metadata       map[string]any `json:"metadata"`
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

	log := h.log.With().
		Interface("workspace_id", workspaceID).
		Interface("env_id", envID).
		Str("event_type", "identify").
		Logger()

	var req IdentifyRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	_, err = h.svc.Identify(c.Context(), services.IdentifySubscriberInput{
		WorkspaceID:    workspaceID.String(),
		EnvironmentID:  envID.String(),
		ExternalUserID: req.ExternalUserID,
		Channel:        req.Channel,
		ContactValue:   req.ContactValue,
		Metadata:       req.Metadata,
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to identify user")
		return response.InternalServerError(c)
	}

	log.Info().Str("external_user_id", req.ExternalUserID).Msg("user identified smoothly")
	return response.OK(c, "user identified successfully", nil)
}

type UpsertPreferenceRequest struct {
	SubscriberID string `json:"subscriber_id" validate:"required"`
	Channel      string `json:"channel" validate:"required"`
	EventType    string `json:"event_type"` // optional
	IsEnabled    bool   `json:"is_enabled"`
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

	var req UpsertPreferenceRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	_, err = h.svc.UpsertPreference(c.Context(), services.UpsertPreferenceInput{
		WorkspaceID:   workspaceID.String(),
		EnvironmentID: envID.String(),
		SubscriberID:  req.SubscriberID,
		Channel:       req.Channel,
		EventType:     req.EventType,
		IsEnabled:     req.IsEnabled,
	})

	if err != nil {
		h.log.Error().Err(err).Msg("failed to upsert preference")
		return response.InternalServerError(c)
	}

	return response.OK(c, "preference updated successfully", nil)
}
