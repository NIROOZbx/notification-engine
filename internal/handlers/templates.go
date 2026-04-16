package handlers

import (
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type TemplateHandler struct {
	service services.TemplateService
	log     zerolog.Logger
}

func NewTemplateHandler(service services.TemplateService, log zerolog.Logger) *TemplateHandler {
	return &TemplateHandler{service: service, log: log}
}

// ---- Template handlers ----

func (h *TemplateHandler) Create(c fiber.Ctx) error {
	authContext, err := utils.GetAuthContext(c)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	var req dtos.CreateTemplateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	layoutID := pgtype.UUID{}
	if req.LayoutID != nil && *req.LayoutID != "" {
		parsedID, err := utils.StringToUUID(*req.LayoutID)
		if err != nil {
			return response.BadRequest(c, nil, "Invalid format for layout_id")
		}
		layoutID = parsedID
	}

	params := domain.CreateTemplateParams{
		WorkspaceID:   authContext.WorkspaceID,
		EnvironmentID: authContext.EnvID,
		CreatedBy:     authContext.UserID,
		Name:          req.Name,
		Description:   req.Description,
		EventType:     utils.FormatEventType(req.EventType),
		LayoutID:      layoutID,
	}

	t, err := h.service.Create(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("Template creation failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.Created(c, "template created", toTemplateResponse(t))
}

func (h *TemplateHandler) GetByID(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	id, ok := utils.ParseIDParam(c, "templateID")

	if !ok {
		return response.BadRequest(c, nil, "invalid template id")
	}

	t, err := h.service.GetByID(c.Context(), id, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("templateID", id).Msg("GetByID failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.OK(c, "template fetched", toTemplateResponse(t))
}

func (h *TemplateHandler) List(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	templates, err := h.service.List(c.Context(), workspaceID, envID)
	if err != nil {
		h.log.Error().Err(err).Msg("List templates failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.TemplateResponse, len(templates))
	for i, t := range templates {
		resp[i] = toTemplateResponse(t)
	}
	return response.OK(c, "templates fetched", resp)
}

func (h *TemplateHandler) Update(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	id, ok := utils.ParseIDParam(c, "templateID")

	if !ok {
		return response.BadRequest(c, nil, "invalid template id")
	}

	var req dtos.UpdateTemplateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, apperrors.ErrReqBody.Error())
	}

	params := domain.UpdateTemplateParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	if req.LayoutID != nil {
		lid := utils.MustStringToUUID(*req.LayoutID)
		params.LayoutID = &lid
	}

	t, err := h.service.Update(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("Template update failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.OK(c, "updated", toTemplateResponse(t))
}

func (h *TemplateHandler) Delete(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	id, ok := utils.ParseIDParam(c, "templateID")

	if !ok {
		return response.BadRequest(c, nil, "invalid template id")
	}

	err = h.service.Delete(c.Context(), id, workspaceID)
	if err != nil {
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.OK(c, "template deleted", nil)
}

// ---- Template Channel handlers ----

func (h *TemplateHandler) CreateChannel(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	templateID,ok := utils.ParseIDParam(c,"templateID")

	if !ok{
		return response.BadRequest(c,nil,"provide valid template ID")
	}

	var req dtos.CreateTemplateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, apperrors.ErrReqBody.Error())
	}

	configID := pgtype.UUID{}
	if req.ChannelConfigID != nil {
		configID, _ = utils.StringToUUID(*req.ChannelConfigID)
	}

	params := domain.CreateTemplateChannelParams{
		TemplateID:      templateID,
		WorkspaceID:     workspaceID,
		ChannelConfigID: configID,
		Channel:         req.Channel,
		Content:         req.Content,
	}

	tc, err := h.service.CreateChannel(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("CreateChannel failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.Created(c, "channel created", toTemplateChannelResponse(tc))
}

func (h *TemplateHandler) ListChannels(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	templateID := c.Params("templateId")

	channels, err := h.service.ListChannels(c.Context(), utils.MustStringToUUID(templateID), workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("templateID", templateID).Msg("ListChannels failed")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.TemplateChannelResponse, len(channels))
	for i, tc := range channels {
		resp[i] = toTemplateChannelResponse(tc)
	}
	return response.OK(c, "channels fetched", resp)
}

func (h *TemplateHandler) UpdateChannel(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	templateID, ok := utils.ParseIDParam(c, "templateID")
	if !ok {
		return response.BadRequest(c, nil, "invalid template id")
	}
	channelID, ok := utils.ParseIDParam(c, "channelID")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel id")
	}

	var req dtos.UpdateTemplateChannelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.BadRequest(c, nil, apperrors.ErrReqBody.Error())
	}

	params := domain.UpdateTemplateChannelParams{
		ID:          channelID,
		TemplateID:  templateID,
		WorkspaceID: workspaceID,
		Content:     req.Content,
		IsActive:    req.IsActive,
	}

	tc, err := h.service.UpdateChannel(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Interface("params", params).Msg("UpdateChannel failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.OK(c, "channel updated", toTemplateChannelResponse(tc))
}

func (h *TemplateHandler) DeleteChannel(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	templateID, ok := utils.ParseIDParam(c, "templateID")
	if !ok {
		return response.BadRequest(c, nil, "invalid template id")
	}
	channelID, ok := utils.ParseIDParam(c, "channelID")
	if !ok {
		return response.BadRequest(c, nil, "invalid channel id")
	}

	err = h.service.DeleteChannel(c.Context(), channelID, templateID, workspaceID)
	if err != nil {
		h.log.Error().Err(err).Interface("channelID", channelID).Msg("DeleteChannel failed")
		return helpers.HandleServiceError(c, err, h.log)
	}
	return response.OK(c, "channel deleted", nil)
}

// ---- Mappers ----

func toTemplateResponse(t *domain.Template) dtos.TemplateResponse {
	layoutIDStr := utils.UUIDToString(t.LayoutID)
	var layoutID *string
	if layoutIDStr != "" {
		layoutID = &layoutIDStr
	}

	return dtos.TemplateResponse{
		ID:            utils.UUIDToString(t.ID),
		WorkspaceID:   utils.UUIDToString(t.WorkspaceID),
		EnvironmentID: utils.UUIDToString(t.EnvironmentID),
		LayoutID:      layoutID,
		Name:          t.Name,
		Description:   t.Description,
		EventType:     t.EventType,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toTemplateChannelResponse(tc *domain.TemplateChannel) dtos.TemplateChannelResponse {
	configIDStr := utils.UUIDToString(tc.ChannelConfigID)
	var configID *string
	if configIDStr != "" {
		configID = &configIDStr
	}

	var content map[string]any
	_ = sonic.Unmarshal(tc.Content, &content)

	return dtos.TemplateChannelResponse{
		ID:              utils.UUIDToString(tc.ID),
		TemplateID:      utils.UUIDToString(tc.TemplateID),
		ChannelConfigID: configID,
		Channel:         tc.Channel,
		Content:         content,
		IsActive:        tc.IsActive,
		CreatedAt:       tc.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       tc.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
