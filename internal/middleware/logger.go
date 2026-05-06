package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func NewTimeLoggerMiddleware(log zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("duration", duration.String()).
			Int64("duration_ms", duration.Milliseconds()).
			Int("status", c.Response().StatusCode()).
			Msg("HTTP request processed")

		return err
	}
}
