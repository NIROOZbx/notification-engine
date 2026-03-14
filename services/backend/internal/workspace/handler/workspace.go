package handlers

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type WorkspaceHandler struct {
	workspaceSvc workspaceSvc.WorkspaceService
	log zerolog.Logger
}

func NewWorkspaceHandler(svc workspaceSvc.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceSvc: svc}
}

func (h *WorkspaceHandler) GetCurrentWorkspace(c fiber.Ctx) error {
	workspaceIDStr, ok := c.Locals("workspaceID").(string)
	if !ok || workspaceIDStr == "" {
		h.log.Warn().Msg("GetCurrentWorkspace hit but no workspaceID in context")
		return response.BadRequest(c,nil,"no workspace associated with session")
	}

	wspID, err := utils.StringToUUID(workspaceIDStr)
	if err != nil {
		h.log.Warn().Err(err).Str("workspaceIDStr", workspaceIDStr).Msg("Invalid workspace UUID format")
		return response.BadRequest(c,nil,"invalid workspace ID")
	}

	workspace, err := h.workspaceSvc.GetByID(c.Context(), wspID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.log.Warn().Str("workspaceID", workspaceIDStr).Msg("Requested workspace not found in database")
			return response.NotFound(c, "workspace not found")
		}
		h.log.Error().Err(err).Str("workspaceID", workspaceIDStr).Msg("Database failure while fetching workspace")
		
		return response.InternalServerError(c)
	}

	return response.OK(c,"workspace fetched successfully",workspace)
}