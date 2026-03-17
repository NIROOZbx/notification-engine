package middleware

import (
	services "github.com/NIROOZbx/notification-engine/services/backend/internal/api_keys"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type apiKeyMiddleware struct {
	svc services.APIKeyService
	log zerolog.Logger
}

type ApiKeyMiddleware interface {
	Authenticate(c fiber.Ctx) error
}

func (a *apiKeyMiddleware) Authenticate(c fiber.Ctx) error {

	rawKey := c.Get("X-API-Key")
	if rawKey == "" {
		a.log.Warn().Msg("api key authentication failed: missing X-API-Key header")
		return response.Unauthorized(c, "missing api key")
	}

	validatedKey, err := a.svc.ValidateAPIKey(c.Context(), rawKey)
	if err != nil {
		a.log.Warn().Err(err).Msg("api key authentication failed: invalid key")
		return response.Unauthorized(c, "invalid api key")
	}

	c.Locals("wid", validatedKey.WorkspaceID)
	c.Locals("envID", validatedKey.EnvID)
	c.Locals("keyID", validatedKey.ID)

	a.log.Debug().
		Str("workspaceID", validatedKey.WorkspaceID.String()).
		Str("keyID", validatedKey.ID.String()).
		Msg("api key authenticated successfully")

	return c.Next()

}

func NewApiKeyMiddleware(svc services.APIKeyService, log zerolog.Logger) ApiKeyMiddleware {
	return &apiKeyMiddleware{
		svc: svc,
		log: log,
	}
}