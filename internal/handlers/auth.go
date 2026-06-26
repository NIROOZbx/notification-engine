package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/session"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
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
	store   session.Store
	engine  *core.Engine
}

func NewAuthHandler(service services.AuthService, cfg *config.AuthConfig, log zerolog.Logger, store session.Store, engine *core.Engine) *AuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,
		log:     log,
		store:   store,
		engine:  engine,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dtos.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Warn().Err(err).Msg("failed to bind register request body")
		return response.BadRequest(c, nil, "invalid request body")
	}

	user, err := h.service.Register(c.Context(), req)
	if err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("failed to register user")
		return response.BadRequest(c, nil, "registration failed, please check your input or try again later")
	}

	if err := h.sendVerificationEmail(c.Context(), utils.UUIDToString(user.ID), req.Email); err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("failed to send initial verification email")
		return response.InternalServerError(c)
	}

	h.log.Info().Str("email", req.Email).Msg("user registered successfully, awaiting email verification")
	return response.Created(c, "Registration successful. Please check your email to verify your account.", nil)
}

func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		h.log.Warn().Msg("verification token is missing from request")
		return response.BadRequest(c, nil, "verification token is required")
	}

	h.log.Debug().Msg("processing email verification token")
	userResp, tokenPair, err := h.service.VerifyEmail(c.Context(), token)
	h.log.Warn().Str("the token user send",token).Msg("in verify email handler")
	if err != nil {
		h.log.Warn().Err(err).Msg("email verification service call failed")
		return response.BadRequest(c, nil, "invalid or expired verification token")
	}

	jwt.SetTokenCookies(c, tokenPair, h.cfg.ToJWTConfig())

	h.log.Info().
		Str("userID", userResp.User.UserID).
		Str("email", userResp.User.Email).
		Bool("hasWorkspace", userResp.User.HasWorkspace).
		Msg("user email verified and logged in successfully")

	return response.OK(c, "email verified successfully", userResp)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Warn().Err(err).Msg("failed to bind login request body")
		return response.BadRequest(c, nil, "invalid request body")
	}
	userResp, tokenPair, err := h.service.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotVerified) {
			user, findErr := h.service.FindUserByEmail(c.Context(), req.Email)
			if findErr != nil {
				h.log.Error().Err(findErr).Str("email", req.Email).Msg("failed to find user for verification resend")
				return response.Forbidden(c, nil, "Email not verified. Please request a new verification link.")
			}
			if sendErr := h.sendVerificationEmail(c.Context(), utils.UUIDToString(user.ID), req.Email); sendErr != nil {
				h.log.Error().Err(sendErr).Str("email", req.Email).Msg("failed to resend verification email during login")
				return response.Forbidden(c, nil, "Email not verified. We tried to send a new link but failed. Please try again later.")
			}
			return response.Forbidden(c, nil, "Email not verified. A new verification link has been sent to your email.")
		}
		h.log.Warn().Err(err).Str("email", req.Email).Msg("failed application login attempt")
		return response.Unauthorized(c, "invalid credentials")
	}

	jwt.SetTokenCookies(c, tokenPair, h.cfg.ToJWTConfig())

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

	jwt.SetTokenCookies(c, tokenPair, h.cfg.ToJWTConfig())
	h.log.Info().Str("userID", user.User.UserID).Bool("has workspace", user.User.HasWorkspace).
		Str("provider", gothUser.Provider).Msg("User logged in successfully")

	redirectURL := h.cfg.FrontendURL
	if !user.User.HasWorkspace {
		redirectURL += consts.OnboardingRoute
	} else {
		redirectURL += consts.DashboardRoute
	}

	return c.Redirect().To(redirectURL)

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

	jwt.SetTokenCookies(c, pair, h.cfg.ToJWTConfig())
	h.log.Info().Interface("userID", userID).Str("workspaceID", dto.Workspace.WorkspaceID).Msg("User successfully completed onboarding")

	return response.OK(c, "onboarding complete", dto)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userID := c.Locals(consts.UID).(pgtype.UUID)

	refreshToken := c.Cookies("refresh_token")

	err := h.service.Logout(c.Context(), userID, refreshToken)

	if err != nil {
		h.log.Error().Err(err).Interface("userID", userID).Msg("Service failed to fully process logout")
		return response.InternalServerError(c)
	}
	jwt.ClearTokenCookies(c, h.cfg.ToJWTConfig())
	h.log.Info().Interface("userID", userID).Msg("User logged out successfully")

	return response.OK(c, "logged out successfully", nil)
}


func (h *AuthHandler) sendVerificationEmail(ctx context.Context, userID, email string) error {
	tkn, err := helpers.GenerateSecureToken()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("email-verify:%s", tkn)
	h.log.Debug().Str("key", key).Str("userID", userID).Msg("storing verification token in redis")

	err = h.store.Set(ctx, key, userID, consts.EmailVerificationTTL)
	if err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", h.cfg.FrontendURL, tkn)

	return h.engine.Ingest(ctx, h.cfg.SystemWorkspaceID, h.cfg.SystemEnvID, &models.TriggerPayload{
		EventType:      "user.verification",
		RecipientEmail: email,
		Data: map[string]interface{}{
			"verification_url": verifyURL,
		},
	})
}

func (h *AuthHandler) ResendEmail(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind().Body(&req); err != nil || req.Email == "" {
		h.log.Warn().Err(err).Msg("failed to bind resend email request")
		return response.BadRequest(c, nil, "email is required")
	}

	h.log.Info().Str("email", req.Email).Msg("request to resend verification email")
	user, err := h.service.FindUserByEmail(c.Context(), req.Email)
	if err != nil {
		h.log.Warn().Err(err).Str("email", req.Email).Msg("resend requested for non-existent user email")
		return response.OK(c, "If your email is registered, a verification link has been sent.", nil)
	}

	if user.IsVerified {
		h.log.Info().Str("email", req.Email).Msg("resend requested for already verified user")
		return response.BadRequest(c, nil, "email is already verified")
	}

	if err := h.sendVerificationEmail(c.Context(), utils.UUIDToString(user.ID), req.Email); err != nil {
		h.log.Error().Err(err).Str("email", req.Email).Msg("failed to resend verification email")
		return response.InternalServerError(c)
	}

	h.log.Info().Str("email", req.Email).Msg("verification email resent successfully")
	return response.OK(c, "Verification email resent successfully", nil)
}
