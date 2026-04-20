package handlers

import (
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
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
	ScheduledAt   *time.Time      `json:"scheduled_at" `
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

	if err := channelValidator(&req); err != nil {
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
		ScheduledAt: req.ScheduledAt,
	}

	err = h.engine.Ingest(
		c.Context(),
		workspaceID.String(),
		envID.String(),
		payload,
	)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
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
