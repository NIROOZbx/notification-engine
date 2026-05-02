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

	var contacts []services.ContactInput
	for _, c := range req.Contacts {
		contacts = append(contacts, services.ContactInput{
			Channel:      c.Channel,
			ContactValue: c.ContactValue,
		})
	}

	subscribers, err := h.svc.Identify(c.Context(), services.IdentifySubscriberInput{
		WorkspaceID:    utils.UUIDToString(workspaceID),
		EnvironmentID:  utils.UUIDToString(envID),
		ExternalUserID: req.ExternalUserID,
		Contacts:       contacts,
		Metadata:       req.Metadata,
	})

	if err != nil {
		h.log.Error().Err(err).Str("external_user_id", req.ExternalUserID).Msg("failed to identify user")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.SubscriberResponse, len(subscribers))
	for i, sub := range subscribers {
		resp[i] = toSubscriberResponse(sub)
	}

	return response.OK(c, "user identified successfully", resp)
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

	page := utils.QueryInt32(c, "page", 1)
	pageSize := utils.QueryInt32(c, "page_size", 20)

	result, err := h.svc.List(c.Context(), utils.UUIDToString(workspaceID), utils.UUIDToString(envID), page, pageSize)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list subscribers")
		return helpers.HandleServiceError(c, err, h.log)
	}

	subscriberResponses := make([]dtos.SubscriberResponse, len(result.Subscribers))
	for i, s := range result.Subscribers {
		subscriberResponses[i] = toSubscriberResponse(s)
	}

	resp := dtos.SubscriberListResponse{
		Subscribers: subscriberResponses,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		CurrentPage: result.CurrentPage,
		PageSize:    result.PageSize,
	}

	return response.OK(c, "subscribers fetched successfully", resp)
}

func (h *SubscriberHandler) GetPreferences(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "missing workspace id")
	}
	envID, err := utils.GetEnvID(c)
	if err != nil {
		return response.Unauthorized(c, "missing environment id")
	}

	externalUserID := c.Params("externalUserId")
	if externalUserID == "" {
		return response.BadRequest(c, nil, "external user id is required")
	}

	prefs, err := h.svc.GetPreferencesByExternalID(c.Context(), utils.UUIDToString(workspaceID), utils.UUIDToString(envID), externalUserID)
	if err != nil {
		h.log.Error().Err(err).Str("external_user_id", externalUserID).Msg("failed to get preferences")
		return helpers.HandleServiceError(c, err, h.log)
	}

	resp := make([]dtos.UserPreferenceResponse, len(prefs))
	for i, p := range prefs {
		resp[i] = toUserPreferenceResponse(p)
	}

	return response.OK(c, "preferences fetched successfully", resp)
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
