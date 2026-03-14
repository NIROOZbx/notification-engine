package middleware

import (

	"errors"
	"log"
	"time"

	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/session"
	"github.com/NIROOZbx/notification-engine/services/pkg/apperrors"
	"github.com/NIROOZbx/notification-engine/services/pkg/jwt"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	gojwt "github.com/golang-jwt/jwt/v5"
)

type authMiddleware struct {
	store session.Store
	cfg   *config.AuthConfig
}

type AuthMiddleware interface {
	Auth(c fiber.Ctx) error
	OnboardingAuth(c fiber.Ctx) error
}

func (a *authMiddleware) Auth(c fiber.Ctx) error {

	claims, err := a.validate(c)

	if err!=nil{
		return response.Unauthorized(c,apperrors.ErrForbidden.Error())
	}

	if !claims.IsOnboarded{
		return  response.Forbidden(c,nil, "workspace setup required")
	}

	
	c.Locals("uid", claims.UserID)
    c.Locals("wid", claims.WorkspaceID)
    c.Locals("role", claims.Role)
    return c.Next()

}

func (a *authMiddleware) OnboardingAuth(c fiber.Ctx) error {

	claims, err := a.validate(c)
if err!=nil{
		return response.Unauthorized(c, apperrors.ErrUnauthorized.Error())
	}

	if claims.IsOnboarded {
        return response.Forbidden(c,nil, "already onboarded")
    }
	c.Locals("uid", claims.UserID)

	return c.Next()

}

func (a *authMiddleware) validate(c fiber.Ctx) (*jwt.AccessClaims, error) {

	accessToken := c.Cookies("access_token")

	if accessToken == "" {
		return nil, apperrors.ErrUnauthorized
	}

	claims, err := jwt.ParseAccessToken(accessToken, []byte(a.cfg.AccessTokenSecret))

	if errors.Is(err, gojwt.ErrTokenExpired) {
		return a.silentRefresh(c, claims)
	}

	if err != nil {
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}
	version, verErr := a.store.GetTokenVersion(c.Context(), claims.UserID)
	if verErr != nil {
		return claims, nil
	}
	if claims.Version < version {
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	return claims, nil

}

func (a *authMiddleware) silentRefresh(c fiber.Ctx, claims *jwt.AccessClaims) (*jwt.AccessClaims, error) {

	token := c.Cookies("refresh_token")

	if token == "" {
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	refreshClaims, err := jwt.ParseRefreshToken(token, []byte(a.cfg.RefreshTokenSecret))

	if err != nil {
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	blacklisted, err := a.store.IsRefreshBlacklisted(c.Context(), refreshClaims.TokenID)

	if err != nil {
		log.Printf("redis down during blacklist check for user %s: %v", claims.UserID, err)
	}

	if blacklisted {
		a.store.UpgradeTokenVersion(c.Context(), claims.UserID)
		jwt.ClearTokenCookies(c)
		return nil, apperrors.ErrUnauthorized
	}

	if err := a.store.BlackListRefreshToken(c.Context(), refreshClaims.TokenID, refreshClaims.ExpiresAt.Time); err != nil {
		log.Printf("failed to blacklist token for user %s: %v", claims.UserID, err)
	}
	newVer, _ := a.store.GetTokenVersion(c.Context(), claims.UserID)

	payload := &jwt.TokenPayload{
		Role:        claims.Role,
		UserID:      claims.UserID,
		WorkspaceID: claims.WorkspaceID,
		Version:     newVer,
		IsOnboarded: claims.IsOnboarded,
	}

	jwtConfig := a.cfg.ToJWTConfig()

	pair, err := jwt.GenerateTokenPair(jwtConfig, *payload)
	if err != nil {
		return nil, err
	}
	expiry := time.Duration(a.cfg.RefreshExpiryHours) * time.Hour
	if err := a.store.StoreRefreshToken(c.Context(), pair.TokenID, claims.UserID, expiry); err != nil {
		log.Printf("failed to store new refresh token for user %s: %v", claims.UserID, err)

	}
	isProd := a.cfg.Environment == "production"

	jwt.SetTokenCookies(c, pair, a.cfg.AccessExpiryMinutes, a.cfg.RefreshExpiryHours, isProd)

	return jwt.ParseAccessToken(pair.AccessToken, []byte(a.cfg.AccessTokenSecret))

}

func NewMiddleware(store session.Store, cfg *config.AuthConfig) AuthMiddleware {
	return &authMiddleware{
		store: store,
		cfg:   cfg,
	}
}
