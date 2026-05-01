package handlers

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/engine/delivery"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type WebhookRepository interface {
	UpdateDeliveryStatusByProviderID(ctx context.Context, input domain.UpdateDeliveryStatusInput) error
}

type WebhookHandler struct {
	repo WebhookRepository
	log  zerolog.Logger
}

func NewWebhookHandler(repo WebhookRepository, log zerolog.Logger) *WebhookHandler {
	return &WebhookHandler{
		repo: repo,
		log:  log,
	}
}

func (h *WebhookHandler) processEvent(c fiber.Ctx, event *delivery.DeliveryEvent) error {
	if event == nil {
		return nil
	}

	h.log.Info().
		Str("provider", event.Provider).
		Str("provider_id", event.ProviderMessageID).
		Str("status", string(event.Status)).
		Msg("processing webhook delivery event")


	input := domain.UpdateDeliveryStatusInput{
		ProviderMessageID: event.ProviderMessageID,
		DeliveryStatus:    string(event.Status),
		Timestamp:         event.Timestamp,
		ErrorMessage:      event.ErrorMessage,
	}

	err := h.repo.UpdateDeliveryStatusByProviderID(c.Context(), input)
	if err != nil {
		h.log.Error().Err(err).Str("provider_id", event.ProviderMessageID).Msg("failed to update delivery status")
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	return nil
}

func (h *WebhookHandler) HandleTwilioWebhook(c fiber.Ctx) error {
	event, err := delivery.ParseTwilioEvent(c.Body())
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse twilio webhook")
		return response.BadRequest(c,nil, "invalid twilio payload")
	}

	if err := h.processEvent(c, event); err != nil {
		return response.InternalServerError(c)
	}

	return response.OK(c, "webhook processed successfully", nil)
}

func (h *WebhookHandler) HandleSendGridWebhook(c fiber.Ctx) error {
	events, err := delivery.ParseSendGridEvent(c.Body())
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse sendgrid webhook")
		return response.BadRequest(c,nil, "invalid sendgrid payload")
	}

	for _, event := range events {
		if err := h.processEvent(c, event); err != nil {
			h.log.Error().Err(err).Str("provider_id", event.ProviderMessageID).Msg("Failed to process sendgrid event")
		}
	}

	return response.OK(c, "webhook processed successfully", nil)
}

func (h *WebhookHandler) HandleSESWebhook(c fiber.Ctx) error {
	event, err := delivery.ParseSesEvent(c.Body())
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse ses webhook")
		return response.BadRequest(c,nil, "invalid ses payload")
	}

	if err := h.processEvent(c, event); err != nil {
		return response.InternalServerError(c)
	}

	return response.OK(c, "webhook processed successfully", nil)
}
