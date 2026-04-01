package handlers

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type WorkspaceHandler struct {
	workspaceSvc services.WorkspaceService
	log          zerolog.Logger
}

func NewWorkspaceHandler(svc services.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceSvc: svc}
}

func (h *WorkspaceHandler) GetCurrentWorkspace(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		h.log.Warn().Err(err).Msg("failed to extract WID")
		return response.BadRequest(c, nil, err.Error())
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

	return response.OK(c, "workspace fetched successfully", workspace)
}

func (h *WorkspaceHandler) GetWorkspaceMembers(c fiber.Ctx) error {
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		h.log.Warn().Err(err).Msg("failed to extract WID")
		return response.BadRequest(c, nil, err.Error())
	}
	members, err := h.workspaceSvc.GetMembers(c.Context(), workspaceID)

	if err != nil {
		h.log.Error().Err(err).Interface("workspaceID", workspaceID).Msg("failed to fetch members")
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NotFound(c, "workspace members not found")
		}
		return response.InternalServerError(c)
	}

	return response.OK(c, "members", members)

}

func (h *WorkspaceHandler) UpdateName(c fiber.Ctx) error {

	var req struct {
		Name string `json:"name" validate:"required"`
	}

	if err := c.Bind().Body(&req); err != nil {
		h.log.Warn().Err(err).Msg("failed to bind register request body")
		return response.BadRequest(c, nil, "invalid request body")
	}
	workspaceID, err := utils.GetWID(c)
	if err != nil {
		h.log.Warn().Err(err).Msg("failed to extract WID")
		return response.BadRequest(c, nil, err.Error())
	}

	updatedWorkspace, err := h.workspaceSvc.UpdateName(c.Context(), workspaceID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			return response.NotFound(c, "workspace not found")
		default:
			h.log.Error().Err(err).
				Interface("workspaceID", workspaceID).
				Str("newName", req.Name).
				Msg("failed to update workspace name")
			return response.InternalServerError(c)
		}
	}

	return response.OK(c, "workspace updated", updatedWorkspace)
}

func (h *WorkspaceHandler) UpdateMemberRole(c fiber.Ctx) error {
	var req dtos.UpdateMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, nil, "invalid request body")
	}

	targetUserID, ok := utils.ParseIDParam(c, "userID")
	if !ok {
		return response.BadRequest(c, nil, "invalid userID format")
	}

	caller, err := utils.GetCallerContext(c)

	params := services.UpdateMemberRoleParams{
		WorkspaceID:  caller.WorkspaceID,
		CallerRole:   caller.CallerRole,
		CallerID:     caller.CallerID,
		TargetUserID: targetUserID,
		Role:         req.Role,
	}

	updatedMember, err := h.workspaceSvc.UpdateMemberRole(c.Context(), params)

	if err != nil {
		h.log.Error().Err(err).Msg("failed to update member role")
		if errors.Is(err, apperrors.ErrForbidden) {
			return response.Forbidden(c, nil, err.Error())
		}
		return response.InternalServerError(c)
	}

	return response.OK(c, "member role updated", updatedMember)
}

func (h *WorkspaceHandler) RemoveMember(c fiber.Ctx) error {
	targetUserID, ok := utils.ParseIDParam(c, "userID")
	if !ok {
		return response.BadRequest(c, nil, "invalid userID format")
	}

	caller, err := utils.GetCallerContext(c)
	if err != nil {
		return response.BadRequest(c, nil, err.Error())
	}

	err = h.workspaceSvc.RemoveMember(c.Context(), services.RemoveMemberParams{
		WorkspaceID:  caller.WorkspaceID,
		CallerID:     caller.CallerID,
		CallerRole:   caller.CallerRole,
		TargetUserID: targetUserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrForbidden):
			return response.Forbidden(c, nil, "insufficient permissions")
		case errors.Is(err, apperrors.ErrNotFound):
			return response.NotFound(c, "member not found")
		default:
			h.log.Error().Err(err).Msg("failed to remove member")
			return response.InternalServerError(c)
		}
	}

	return response.OK(c, "member removed", nil)
}

