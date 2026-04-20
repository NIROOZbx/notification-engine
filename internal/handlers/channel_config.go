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

type ChannelConfigHandler struct {
	service services.ChannelConfigService
	log     zerolog.Logger
}

func NewChannelConfigHandler(service services.ChannelConfigService, log zerolog.Logger) *ChannelConfigHandler {
	return &ChannelConfigHandler{
		service: service,
		log:     log,
	}
}

func (h *ChannelConfigHandler) Create(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	var req dtos.CreateChannelConfigRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	params := domain.CreateChannelConfigParams{
		WorkspaceID: workspaceID,
		Channel:     req.Channel,
		Provider:    req.Provider,
		DisplayName: req.DisplayName,
		Credentials: req.Credentials,
		IsActive:    req.IsActive,
		IsDefault:   req.IsDefault,
	}

	config, err := h.service.Create(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("Channel config creation failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.Created(c, "channel config created", toChannelConfigResponse(config))
}

func (h *ChannelConfigHandler) List(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	configs, err := h.service.List(c.Context(), workspaceID)
	if err != nil {
		h.log.Error().Err(err).Msg("List channel configs failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.ChannelConfigResponse, len(configs))
	for i, cfg := range configs {
		resp[i] = toChannelConfigResponse(cfg)
	}

	return response.OK(c, "channel configs fetched", resp)
}

func (h *ChannelConfigHandler) GetByID(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel config id")
	}

	config, err := h.service.GetChannelConfigByID(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("GetChannelConfigByID failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "channel config fetched", toChannelConfigResponse(config))
}

func (h *ChannelConfigHandler) Update(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel config id")
	}

	var req dtos.UpdateChannelConfigRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	params := domain.UpdateChannelConfigParams{
		ID:          id,
		WorkspaceID: workspaceID,
		DisplayName: req.DisplayName,
		Credentials: req.Credentials,
		IsActive:    req.IsActive,
	}

	config, err := h.service.UpdateChannelConfig(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("Channel config update failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "channel config updated", toChannelConfigResponse(config))
}

func (h *ChannelConfigHandler) Delete(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel config id")
	}

	err = h.service.DeleteChannelConfig(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("Channel config deletion failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "channel config deleted", nil)
}

func (h *ChannelConfigHandler) SetDefault(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c,"id")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel config id")
	}

	err = h.service.SetChannelConfigDefault(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("SetDefault channel config failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "channel config set as default successfully", nil)
}

func toChannelConfigResponse(cfg *domain.ChannelConfig) dtos.ChannelConfigResponse {
	return dtos.ChannelConfigResponse{
		ID:          utils.UUIDToString(cfg.ID),
		WorkspaceID: utils.UUIDToString(cfg.WorkspaceID),
		Channel:     cfg.Channel,
		Provider:    cfg.Provider,
		DisplayName: cfg.DisplayName,
		IsActive:    cfg.IsActive,
		IsDefault:   cfg.IsDefault,
		CreatedAt:   cfg.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   cfg.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}