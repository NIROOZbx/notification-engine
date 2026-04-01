package handlers

import (
	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type APIKeyHandler struct {
	svc services.APIKeyService
	log zerolog.Logger
}

func NewAPIKeyHandler(svc services.APIKeyService,log zerolog.Logger) *APIKeyHandler {
	return &APIKeyHandler{svc: svc,log: log}
}

func (h *APIKeyHandler) CreateAPIKey(c fiber.Ctx) error {

	userID := c.Locals(consts.UID).(pgtype.UUID)
	workspaceID := c.Locals(consts.WID).(pgtype.UUID)

	log := h.log.With().
		Interface("user_id", userID).
		Interface("workspace_id", workspaceID).
		Logger()

	var req dtos.CreateAPIKeyRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("failed to bind create api key body")
		return response.BadRequest(c, nil, "invalid json")
	}
	envID, err := utils.StringToUUID(req.EnvironmentID)
	if err != nil {
		log.Warn().Str("env_id_raw", req.EnvironmentID).Msg("invalid environment_id format")
		return response.BadRequest(c, nil, "invalid environment_id format")
	}

	params := services.CreateAPIKeyParams{
		WorkspaceID:   workspaceID,
		EnvironmentID: envID,
		UserID:        userID,
		Label:         req.Label,
		ExpiresIn:     req.ExpiresIn,
	}

	resp, err := h.svc.CreateAPIKey(c.Context(), params)

	if err != nil {
		log.Error().Err(err).
			Interface("env_id", envID).
			Str("label", req.Label).
			Msg("service failed to create api key")
		switch err {
		case apperrors.ErrForbidden:
			return response.Forbidden(c, nil, "api key limit reached for your plan")
		case apperrors.ErrNotFound:
			return response.NotFound(c, "environment not found")
		default:
			return response.InternalServerError(c)
		}
	}
	log.Info().Str("key_id", resp.ID).Msg("api key created successfully")

	return response.Created(c, "api key created successfully", resp)

}

func (h *APIKeyHandler) RevokeAPIKey(c fiber.Ctx) error {
	workspaceID := c.Locals(consts.WID).(pgtype.UUID)

	keyID, ok := utils.ParseIDParam(c,"keyID")
	if !ok {
		return response.BadRequest(c, nil, "invalid key_id")
	}

	log := h.log.With().
		Interface("workspace_id", workspaceID).
		Interface("key_id_param", keyID).
		Logger()
	resp, err := h.svc.RevokeAPIKey(c.Context(), services.RevokeKeyParams{
		ID:          keyID,
		WorkspaceID: workspaceID,
	})

	if err != nil {
		log.Error().Err(err).Msg("service failed to revoke api key")
		switch err {
		case apperrors.ErrNotFound:
			return response.NotFound(c, "api key not found")
		default:
			return response.InternalServerError(c)
		}
	}

	log.Info().Msg("api key revoked successfully")
	return response.OK(c, "api key revoked successfully", resp)
}

func (h *APIKeyHandler) DeleteAPIKey(c fiber.Ctx) error {
	workspaceID := c.Locals(consts.WID).(pgtype.UUID)
	

	keyID, err := utils.StringToUUID(c.Params("keyID"))
	if err != nil {
		return response.BadRequest(c, nil, "invalid key_id")
	}

	log := h.log.With().
		Interface("workspace_id", workspaceID).
		Interface("key_id_param", keyID).
		Logger()

	err = h.svc.DeleteAPIKey(c.Context(), services.DeleteKeyParams{
		ID:          keyID,
		WorkspaceID: workspaceID,
	})

	if err != nil {
		log.Error().Err(err).Msg("service failed to delete api key")
		switch err {
		case apperrors.ErrNotFound:
			return response.NotFound(c, "api key not found")
		default:
			return response.InternalServerError(c)
		}
	}

	log.Info().Msg("api key hard deleted")
	return response.OK(c, "api key deleted successfully", nil)
}

func (h *APIKeyHandler) ListAPIKeys(c fiber.Ctx) error {
	workspaceID := c.Locals(consts.WID).(pgtype.UUID)

	envIDStr := c.Query("env_id")
	if envIDStr == "" {
		return response.BadRequest(c, nil, "environment_id is required as a query parameter")
	}
	log := h.log.With().
		Interface("workspace_id", workspaceID).
		Str("env_id_query", envIDStr).
		Logger()

	envID, err := utils.StringToUUID(envIDStr)
	if err != nil {
		log.Warn().Msg("invalid environment_id format in list request")
		return response.BadRequest(c, nil, "invalid environment_id format")
	}

	keys, err := h.svc.ListAPIKeys(c.Context(), services.ListApiKeyParams{
		WorkspaceID:   workspaceID,
		EnvironmentID: envID,
	})

	if err != nil {
		log.Error().Err(err).Msg("service failed to list api keys")
		return response.InternalServerError(c)
	}
	log.Debug().Int("count", len(keys)).Msg("api keys retrieved")
	return response.OK(c, "API keys retrieved successfully", keys)
}
