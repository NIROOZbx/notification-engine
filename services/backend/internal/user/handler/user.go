package handler

import (
	"errors"

	userSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	userSvc userSvc.UserService
	log zerolog.Logger
}

func NewUserHandler(svc userSvc.UserService) *UserHandler {
    return &UserHandler{userSvc: svc}
}

func (u *UserHandler) GetMe(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		u.log.Warn().Msg("GetMe hit without userID in locals")
		return response.Unauthorized(c,"invalid session")
	}

	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		u.log.Warn().Err(err).Str("userIDStr", userIDStr).Msg("Invalid UUID format for user")
		return response.BadRequest(c, nil, "invalid user ID format")
	}

	user, err := u.userSvc.FindUserByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			u.log.Warn().Str("userID", userIDStr).Msg("User profile not found in database")
			return response.NotFound(c, "user profile not found")
		}
		u.log.Error().Err(err).Str("userID", userIDStr).Msg("Database failure while fetching user profile")
		return response.InternalServerError(c)
	}

	return response.OK(c, "Profile fetched successfully", user)
}