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

type LayoutHandler struct {
	service services.LayoutService
	log     zerolog.Logger
}

func NewLayoutHandler(service services.LayoutService, log zerolog.Logger) *LayoutHandler {
	return &LayoutHandler{service: service, log: log}
}

func (h *LayoutHandler) Create(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	var req dtos.CreateLayoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	params := domain.CreateLayoutParams{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Html:        req.Html,
		IsDefault:   req.IsDefault != nil && *req.IsDefault,
	}

	layout, err := h.service.Create(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("Layout creation failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.Created(c, "layout created", toLayoutResponse(layout))
}

func (h *LayoutHandler) GetByID(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c,"id")
	if !ok {
		return response.BadRequest(c, nil, "invalid layout id")
	}
	

	layout, err := h.service.GetByID(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("GetLayoutByID failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "layout fetched", toLayoutResponse(layout))
}

func (h *LayoutHandler) List(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	layouts, err := h.service.List(c.Context(), workspaceID)
	if err != nil {
		h.log.Error().Err(err).Msg("List layouts failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.LayoutResponse, len(layouts))
	for i, l := range layouts {
		resp[i] = toLayoutResponse(l)
	}

	return response.OK(c, "layouts fetched", resp)
}

func (h *LayoutHandler) Update(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid layout id")
	}

	var req dtos.UpdateLayoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	params := domain.UpdateLayoutParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        "",
		Html:        "",
	}
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Html != nil {
		params.Html = *req.Html
	}

	layout, err := h.service.Update(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("Layout update failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "layout updated", toLayoutResponse(layout))
}

func (h *LayoutHandler) Delete(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid layout id")
	}

	err = h.service.Delete(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("Layout deletion failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "layout deleted", nil)
}

func (h *LayoutHandler) SetDefault(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return response.BadRequest(c, nil, "invalid layout id")
	}

	err = h.service.SetDefault(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("id", id).Msg("SetDefault Layout failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	return response.OK(c, "layout set as default successfully", nil)
}


func toLayoutResponse(l *domain.Layout) dtos.LayoutResponse {
	return dtos.LayoutResponse{
		ID:          utils.UUIDToString(l.ID),
		WorkspaceID: utils.UUIDToString(l.WorkspaceID),
		Name:        l.Name,
		IsDefault:   l.IsDefault,
		Html:        l.Html,
		CreatedAt:   l.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   l.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
