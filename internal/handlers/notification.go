package handlers

import (
	"fmt"

	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type NotificationHandler struct {
	engine *core.Engine
	repo   core.Repository
	log    zerolog.Logger
}

func NewNotificationHandler(engine *core.Engine, repo core.Repository, log zerolog.Logger) *NotificationHandler {
	return &NotificationHandler{
		engine: engine,
		repo:   repo,
		log:    log,
	}
}

type TriggerRequest struct {
	ExternalUserID string         `json:"external_user_id" validate:"required"`
	EventType      string         `json:"event_type"       validate:"required"`
	Data           map[string]any `json:"data"`
	Channels       []string       `json:"channels"`
	IdempotencyKey string         `json:"idempotency_key"`
}

func (h *NotificationHandler) Trigger(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}
	isTest := utils.GetIsTest(c)

	var req TriggerRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}
	
	if err:=channelValidator(&req);err!=nil{
		return response.BadRequest(c, nil, fmt.Sprintf("invalid channel: %s", err))
	}

	log := h.log.With().
		Interface("workspace_id", workspaceID).
		Str("event_type", req.EventType).
		Logger()

	payload := &models.TriggerPayload{
		ExternalUserID: req.ExternalUserID,
		EventType:      req.EventType,
		Data:           req.Data,
		Channels:       req.Channels,
		IdempotencyKey: req.IdempotencyKey,
		IsTest:         isTest,
	}

	err = h.engine.Ingest(
		c.Context(),
		workspaceID.String(),
		envID.String(),
		payload,
	)
	if err != nil {
		switch err {
		case apperrors.ErrTemplateNotFound:
			log.Warn().Err(err).Msg("ingest failed: template not found")
			return response.NotFound(c, "template not found")

		case apperrors.ErrTemplateNotLive:
			log.Warn().Err(err).Msg("ingest failed: template not live")
			return response.BadRequest(c, nil, "template is not live")

		case apperrors.ErrNoActiveChannels:
			log.Warn().Err(err).Msg("ingest failed: no active channels")
			return response.BadRequest(c, nil, "no active channels for template")

		default:
			log.Error().Err(err).Msg("failed to ingest notification")
			return response.InternalServerError(c)
		}
	}

	log.Info().Str("event_type", req.EventType).Msg("notification ingested")
	return response.Accepted(c, "notification queued", nil)
}

func channelValidator(req *TriggerRequest) error {
	
	if len(req.Channels) == 0 {
		return nil
	}
	validChannels := map[string]bool{
		"email": true, "sms": true, "push": true,
		"slack": true, "whatsapp": true, "webhook": true, "in_app": true,
	}

	for _, ch := range req.Channels {
		if !validChannels[ch] {
			return fmt.Errorf("invalid channel: %s", ch)
		}
	}
	return nil
}
