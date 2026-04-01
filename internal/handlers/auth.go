package handlers

import (
	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/pkg/jwt"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shareed2k/goth_fiber/v2"
)

type AuthHandler struct {
	service services.AuthService
	cfg     *config.AuthConfig
	log     zerolog.Logger
}

func NewAuthHandler(service services.AuthService, cfg *config.AuthConfig, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,
		log:     log,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dtos.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Warn().Err(err).Msg("failed to bind register request body")
		return response.BadRequest(c, nil, "invalid request body")
	}

	userResp, tokenPair, err := h.service.Register(c.Context(), req)
	if err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("failed to register user")
		return response.BadRequest(c, nil, "registration failed, please check your input or try again later")
	}

	jwt.SetTokenCookies(c, tokenPair, h.cfg.AccessExpiryMinutes, h.cfg.RefreshExpiryHours, h.isProd())
	h.log.Info().Str("userID", userResp.User.UserID).Msg("user registered successfully via local provider")
	return response.Created(c, "registered successfully", userResp)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Warn().Err(err).Msg("failed to bind login request body")
		return response.BadRequest(c, nil, "invalid request body")
	}
	userResp, tokenPair, err := h.service.Login(c.Context(), req)
	if err != nil {
		h.log.Warn().Err(err).Str("email", req.Email).Msg("failed application login attempt")
		return response.Unauthorized(c, "invalid credentials")
	}

	jwt.SetTokenCookies(c, tokenPair, h.cfg.AccessExpiryMinutes, h.cfg.RefreshExpiryHours, h.isProd())

	h.log.Info().Str("userID", userResp.User.UserID).Msg("user logged in successfully via local provider")
	return response.OK(c, "logged in successfully", userResp)
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
		Name:      gothUser.Name,
		Email:     gothUser.Email,
		AvatarURL: gothUser.AvatarURL,
		Provider:  gothUser.Provider,
		UserID:    gothUser.UserID,
	}

	user, tokenPair, err := h.service.HandleOAuthCallback(c.Context(), userDetails)
	if err != nil {
		h.log.Error().Err(err).Str("email", gothUser.Email).Msg("Failed to process OAuth callback in service")
		return response.InternalServerError(c)
	}

	jwt.SetTokenCookies(c, tokenPair, h.cfg.AccessExpiryMinutes, h.cfg.RefreshExpiryHours, h.isProd())
	h.log.Info().Str("userID", user.User.UserID).Str("provider", gothUser.Provider).Msg("User logged in successfully")

	return response.OK(c, "sucessful login", user)

}

func (h *AuthHandler) CompleteOnboarding(c fiber.Ctx) error {
	userID := c.Locals(consts.UID).(pgtype.UUID)
	var req dtos.OnboardingRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("Invalid request body payload during onboarding")
		return response.BadRequest(c, nil, "invalid request body")
	}

	dto, pair, err := h.service.CompleteOnboarding(c.Context(), userID, req.WorkspaceName)
	if err != nil {
		h.log.Error().Err(err).Interface("userID", userID).Msg("Failed to complete onboarding in service")
		return response.InternalServerError(c)
	}

	jwt.SetTokenCookies(c, pair, h.cfg.AccessExpiryMinutes, h.cfg.RefreshExpiryHours, h.isProd())
	h.log.Info().Interface("userID", userID).Str("workspaceID", dto.Workspace.WorkspaceID).Msg("User successfully completed onboarding")

	return response.OK(c, "onboarding complete", dto)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userID:=c.Locals(consts.UID).(pgtype.UUID)

	refreshToken := c.Cookies("refresh_token")

	err := h.service.Logout(c.Context(), userID, refreshToken)

	if err != nil {
		h.log.Error().Err(err).Interface("userID", userID).Msg("Service failed to fully process logout")
		return response.InternalServerError(c)
	}
	jwt.ClearTokenCookies(c)
	h.log.Info().Interface("userID", userID).Msg("User logged out successfully")

	return response.OK(c, "logged out successfully", nil)
}

func (h *AuthHandler) isProd() bool {
	return h.cfg.Environment == "production"
}
