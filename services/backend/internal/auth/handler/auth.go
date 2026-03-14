package handler

import (
	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/auth/services"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/jwt"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"github.com/shareed2k/goth_fiber/v2"
)



type AuthHandler struct {
	service services.AuthService
	cfg *config.AuthConfig
	log zerolog.Logger
}

func NewAuthHandler(service services.AuthService,cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,

	}
}
func (h *AuthHandler) OAuthLogin(c fiber.Ctx) error {
	return goth_fiber.BeginAuthHandler(c)
}

func (h *AuthHandler) OAuthCallback(c fiber.Ctx) error {

	gothUser, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		h.log.Warn().Err(err).Msg("OAuth authentication failed at provider level")
		return response.Unauthorized(c, "Authentication failed")
	}
	userDetails := &dtos.OAuthUserDetails{
		Name:  gothUser.Name,
		Email: gothUser.Email,
		AvatarURL: gothUser.AvatarURL,
		Provider: gothUser.Provider,
		UserID:   gothUser.UserID,
	
	}

	user,tokenPair,err:=h.service.HandleOAuthCallback(c.Context(), userDetails)
	if err!=nil{
		h.log.Error().Err(err).Str("email", gothUser.Email).Msg("Failed to process OAuth callback in service")
		return response.InternalServerError(c)
	}


	jwt.SetTokenCookies(c,tokenPair,h.cfg.AccessExpiryMinutes,h.cfg.RefreshExpiryHours,h.isProd())
	h.log.Info().Str("userID", user.User.UserID).Str("provider", gothUser.Provider).Msg("User logged in successfully")

	return response.OK(c,"sucessful login",user)

	
}


func (h *AuthHandler) CompleteOnboarding(c fiber.Ctx) error {
    userIDStr := c.Locals("uid").(string)

	if userIDStr == "" {
		h.log.Warn().Msg("CompleteOnboarding hit without uid in locals")
		return response.Unauthorized(c, "invalid session")
	}

    userID, err := utils.StringToUUID(userIDStr)
    if err != nil {
		h.log.Warn().Err(err).Str("userIDStr", userIDStr).Msg("Invalid UUID format in onboarding")
        return response.BadRequest(c, nil, "invalid user id")
    }

    var body struct {
        WorkspaceName string `json:"workspace_name"`
    }
    if err := c.Bind().JSON(&body); err != nil {
		h.log.Warn().Err(err).Msg("Invalid request body payload during onboarding")
        return response.BadRequest(c, nil, "invalid request body")
    }
    if body.WorkspaceName == "" {
		h.log.Warn().Msg("Workspace name omitted during onboarding")
        return response.BadRequest(c, nil, "workspace name is required")
    }

    dto, pair, err := h.service.CompleteOnboarding(c.Context(), userID, body.WorkspaceName)
    if err != nil {
		h.log.Error().Err(err).Str("userID", userIDStr).Msg("Failed to complete onboarding in service")
        return response.InternalServerError(c)
    }

  
	jwt.SetTokenCookies(c,pair,h.cfg.AccessExpiryMinutes,h.cfg.RefreshExpiryHours,h.isProd())
	h.log.Info().Str("userID", userIDStr).Str("workspaceID", dto.Workspace.WorkspaceID).Msg("User successfully completed onboarding")

    return response.OK(c, "onboarding complete", dto)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("uid").(string)
	if !ok {
		h.log.Warn().Msg("Logout attempted without valid uid in context")
		return response.Unauthorized(c,apperrors.ErrUnauthorized.Error())
	}

	refreshToken := c.Cookies("refresh_token")

	err := h.service.Logout(c.Context(), userIDStr, refreshToken)
	
	if err != nil {
		h.log.Error().Err(err).Str("userID", userIDStr).Msg("Service failed to fully process logout")
		return response.InternalServerError(c)
	}
	jwt.ClearTokenCookies(c)
	h.log.Info().Str("userID", userIDStr).Msg("User logged out successfully")

	return response.OK(c, "logged out successfully", nil)
}

func (h *AuthHandler) isProd() bool {
    return h.cfg.Environment == "production"
}