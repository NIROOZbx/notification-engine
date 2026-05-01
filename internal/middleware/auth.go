package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/session"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/pkg/jwt"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type authMiddleware struct {
	store session.Store
	repo  *sqlc.Queries
	cfg   *config.AuthConfig
	log   zerolog.Logger
}

type AuthMiddleware interface {
	Auth(c fiber.Ctx) error
	OnboardingAuth(c fiber.Ctx) error
	RequireRole(allowedRoles ...string) fiber.Handler
}

const gracePeriod = 5 * time.Second

func (a *authMiddleware) Auth(c fiber.Ctx) error {

	claims, err := a.validate(c)

	if err != nil {
		a.log.Warn().Err(err).Msg("authentication failed: invalid or missing token")
		return response.Unauthorized(c, apperrors.ErrForbidden.Error())
	}

	if !claims.IsOnboarded {
		a.log.Info().Str("userID", claims.UserID).Msg("authentication successful but onboarding required")
		return response.Forbidden(c, nil, "workspace setup required")
	}

	userID, err := utils.StringToUUID(claims.UserID)
	if err != nil {
		return response.InternalServerError(c)
	}

	workspaceID, err := utils.StringToUUID(claims.WorkspaceID)
	if err != nil {
		return response.InternalServerError(c)
	}
	envIDStr := c.Get("X-Environment-ID")
	if envIDStr != "" {
		var envID pgtype.UUID
		if err := envID.Scan(envIDStr); err != nil {
			return response.BadRequest(c, nil, "invalid environment id")
		}

		env, err := a.repo.GetEnvironmentByID(c.Context(), envID)
		if err != nil {
			return response.Forbidden(c, nil, "invalid environment")
		}
		if env.WorkspaceID != workspaceID {
			return response.Forbidden(c, nil, "environment does not belong to this workspace")
		}

		c.Locals(consts.ENVID, envID)
	}

	c.Locals(consts.UID, userID)
	c.Locals(consts.WID, workspaceID)
	c.Locals(consts.Role, claims.Role)

	a.log.Debug().
		Str("userID", claims.UserID).
		Str("workspaceID", claims.WorkspaceID).
		Str("role", claims.Role).
		Msg("user authenticated successfully")

	return c.Next()

}

func (a *authMiddleware) OnboardingAuth(c fiber.Ctx) error {

	claims, err := a.validate(c)
	if err != nil {
		fmt.Println("err", err)
		return response.Unauthorized(c, apperrors.ErrUnauthorized.Error())
	}

	if claims.IsOnboarded {
		a.log.Info().Str("userID", claims.UserID).Msg("onboarding auth attempt for already onboarded user")
		return response.Forbidden(c, nil, "already onboarded")
	}

	userID, err := utils.StringToUUID(claims.UserID)
	if err != nil {
		return response.InternalServerError(c)
	}

	c.Locals(consts.UID, userID)

	return c.Next()

}

func (a *authMiddleware) validate(c fiber.Ctx) (*jwt.AccessClaims, error) {

	accessToken := c.Cookies("access_token")

	if accessToken == "" {
		a.log.Debug().Msg("access token cookie missing, attempting silent refresh")
		return a.silentRefresh(c)
	}

	claims, err := jwt.ParseAccessToken(accessToken, []byte(a.cfg.AccessTokenSecret))

	if errors.Is(err, gojwt.ErrTokenExpired) {
		if claims == nil {
			a.log.Debug().Msg("access token expired no attempt")
			return nil, apperrors.ErrUnauthorized
		}
		a.log.Debug().Msg("access token expired, attempting silent refresh")
		return a.silentRefresh(c)
	}

	if err != nil {
		a.log.Debug().Err(err).Msg("access token invalid, attempting recovery via refresh")
		return a.silentRefresh(c)
	}
	version, verErr := a.store.GetTokenVersion(c.Context(), claims.UserID)
	if verErr != nil {
		return claims, nil
	}
	if claims.Version < version {
		a.log.Debug().Msg("token version mismatch")
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	return claims, nil

}

func (a *authMiddleware) silentRefresh(c fiber.Ctx) (*jwt.AccessClaims, error) {

	token := c.Cookies("refresh_token")

	if token == "" {
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	refreshClaims, err := jwt.ParseRefreshToken(token, []byte(a.cfg.RefreshTokenSecret))

	if err != nil {
		a.log.Warn().Err(err).Msg("failed to parse refresh token during silent refresh")
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	userID, err := utils.StringToUUID(refreshClaims.UserID)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	authCtx, err := a.repo.GetUserAuthContext(c.Context(), userID)
	if err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("failed to re-hydrate user context from DB")
		return nil, apperrors.ErrUnauthorized
	}

	blacklistedAt, err := a.store.IsRefreshBlacklisted(c.Context(), refreshClaims.TokenID)

	if err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("redis down during blacklist check")
		return nil, apperrors.ErrInternal
	}
	if !blacklistedAt.IsZero() && time.Since(blacklistedAt) <= gracePeriod{
			a.log.Warn().Str("userID", refreshClaims.UserID).Str("tokenID", refreshClaims.TokenID).Msg("refresh token is blacklisted")
		a.store.UpgradeTokenVersion(c.Context(), refreshClaims.UserID)
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	if err := a.store.BlackListRefreshToken(c.Context(), refreshClaims.TokenID, refreshClaims.IssuedAt.Time); err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("failed to blacklist token")
	}
	newVer, err := a.store.GetTokenVersion(c.Context(), refreshClaims.UserID)
	if err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("failed to fetch token version during silent refresh")
	}

	isOnboarded := authCtx.WorkspaceID.String() != consts.FallBackUUID

	payload := &jwt.TokenPayload{
		Role:        authCtx.Role,
		UserID:      refreshClaims.UserID,
		WorkspaceID: authCtx.WorkspaceID.String(),
		Version:     newVer,
		IsOnboarded: isOnboarded,
	}

	jwtConfig := a.cfg.ToJWTConfig()

	pair, err := jwt.GenerateTokenPair(jwtConfig, *payload)
	if err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("failed to generate token pair during silent refresh")
		return nil, err
	}
	expiry := time.Duration(a.cfg.RefreshExpiryHours) * time.Hour
	if err := a.store.StoreRefreshToken(c.Context(), pair.TokenID, refreshClaims.UserID, expiry); err != nil {
		a.log.Error().Err(err).Str("userID", refreshClaims.UserID).Msg("failed to store new refresh token")

	}
	isProd := a.cfg.Environment == "production"

	jwt.SetTokenCookies(c, pair, a.cfg.AccessExpiryMinutes, a.cfg.RefreshExpiryHours, isProd)

	a.log.Info().
		Str("userID", refreshClaims.UserID).
		Str("workspaceID", authCtx.WorkspaceID.String()).
		Msg("tokens issued successfully during silent refresh")

	return jwt.ParseAccessToken(pair.AccessToken, []byte(a.cfg.AccessTokenSecret))

}

func NewMiddleware(store session.Store, cfg *config.AuthConfig, log zerolog.Logger, repo *sqlc.Queries) AuthMiddleware {
	return &authMiddleware{
		store: store,
		cfg:   cfg,
		log:   log,
		repo:  repo,
	}
}
