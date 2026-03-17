package handlers

import (
	"errors"
	workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
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
	workspaceID, ok := c.Locals("wid").(pgtype.UUID)
	if !ok {
		h.log.Warn().Msg("Get CurrentWork space hit but no workspaceID in context")
		return response.BadRequest(c,nil,"no workspace associated with session")
	}


	workspace, err := h.workspaceSvc.GetByID(c.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.log.Warn().Interface("workspaceID", workspaceID).Msg("Requested workspace not found in database")
			return response.NotFound(c, "workspace not found")
		}
		h.log.Error().Err(err).Interface("workspaceID", workspaceID).Msg("Database failure while fetching workspace")
		
		return response.InternalServerError(c)
	}

	return response.OK(c,"workspace fetched successfully",workspace)
}