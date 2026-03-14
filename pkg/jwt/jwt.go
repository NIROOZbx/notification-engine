package jwt

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenPayload struct {
	Role        string
	UserID      string
	WorkspaceID string
	Version     int64
	IsOnboarded bool
	
}

type Pair struct {
	AccessToken  string
	RefreshToken string
	TokenID      string
}

type AccessClaims struct {
	UserID      string `json:"uid"`
	WorkspaceID string `json:"wid"`
	Role        string `json:"role"`
	Version     int64  `json:"ver"`
	IsOnboarded bool   `json:"onboarded"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID  string `json:"uid"`
	TokenID string `json:"tid"`
	jwt.RegisteredClaims
}

type Config struct {
    AccessTokenSecret  string
    RefreshTokenSecret string
    AccessExpiryMinutes int
    RefreshExpiryHours  int
}

func GenerateTokenPair(cfg Config, payload TokenPayload) (*Pair, error) {

	//access token with claims
	accessExpiry := time.Duration(cfg.AccessExpiryMinutes) * time.Minute
	accessClaims := &AccessClaims{
		UserID:      payload.UserID,
		WorkspaceID: payload.WorkspaceID,
		Role:        payload.Role,
		Version:     payload.Version,
		IsOnboarded: payload.IsOnboarded,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(cfg.AccessTokenSecret))
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	//refresh token with claims

	tokenID := uuid.NewString()

	refreshExpiry := time.Duration(cfg.RefreshExpiryHours) * time.Hour
	refreshClaims := &RefreshClaims{
		UserID:  payload.UserID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(cfg.RefreshTokenSecret))
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &Pair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenID:      tokenID,
	}, nil

}

func ParseAccessToken(tokenStr string, secretKey []byte) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil

}

func ParseRefreshToken(tokenStr string, secretKey []byte) (*RefreshClaims, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		 if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func SetTokenCookies(c fiber.Ctx, pair *Pair,accessExpiryMinutes int, refreshExpiryHours int,isProd bool) {
    c.Cookie(&fiber.Cookie{
        Name:     "access_token",
        Value:    pair.AccessToken,
        HTTPOnly: true,
        Secure:   isProd,
        SameSite: "Lax",
        MaxAge:   accessExpiryMinutes * 60,
        Path:     "/",
    })
    c.Cookie(&fiber.Cookie{
        Name:     "refresh_token",
        Value:    pair.RefreshToken,
        HTTPOnly: true,
        Secure:   isProd,
        SameSite: "Lax",
        MaxAge:   refreshExpiryHours * 3600,
        Path:     "/",
    })
}

func ClearTokenCookies(c fiber.Ctx) {
    c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", HTTPOnly: true, MaxAge: -1, Path: "/"})
    c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", HTTPOnly: true, MaxAge: -1, Path: "/"})
}

