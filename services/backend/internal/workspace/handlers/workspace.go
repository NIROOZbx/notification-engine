package handlers

import workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"

type WorkspaceHandler struct {
	workspaceSvc workspaceSvc.WorkspaceService
}

func NewWorkspaceHandler(svc workspaceSvc.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceSvc: svc}
}