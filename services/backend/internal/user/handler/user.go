package handler

import (
	"errors"

	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	userSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	userSvc userSvc.UserService
	log     zerolog.Logger
}

func NewUserHandler(svc userSvc.UserService) *UserHandler {
	return &UserHandler{userSvc: svc}
}

func (u *UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := c.Locals("uid").(pgtype.UUID)
	if !ok {
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
