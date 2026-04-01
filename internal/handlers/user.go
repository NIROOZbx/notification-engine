package handlers

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	userSvc services.UserService
	log     zerolog.Logger
}

func NewUserHandler(svc services.UserService,log zerolog.Logger) *UserHandler {
	return &UserHandler{userSvc: svc,log: log}
}

func (u *UserHandler) GetMe(c fiber.Ctx) error {
	userID, err:= utils.GetUID(c)
	if err!=nil {
		u.log.Warn().Msg("GetMe hit without userID in locals")
		return response.Unauthorized(c, "invalid session")
	}

	row, err := u.userSvc.GetFullUserDetails(c.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			u.log.Warn().Interface("userID", userID).Msg("User profile not found in database")
			return response.NotFound(c, "user profile not found")
		}
		u.log.Error().Err(err).Interface("userID", userID).Msg("Database failure while fetching user profile")
		return response.InternalServerError(c)
	}
	res := dtos.AuthResponse{
        User: dtos.UserDetails{
            UserID:    row.UserID.String(),
            Name:      row.FullName,
            Email:     row.Email,
            AvatarURL: row.AvatarUrl.String,
        },
        Workspace: &dtos.WorkSpaceDetails{
            WorkspaceID:   row.WorkspaceID.String(),
            WorkSpaceName: row.WorkspaceName,
            Slug:          row.Slug,
            Role:          row.Role,
        },
    }

    return response.OK(c, "Profile fetched successfully", res)
}
