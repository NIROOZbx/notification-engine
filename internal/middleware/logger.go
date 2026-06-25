package middleware

import (
	"strconv"
	"time"

	"github.com/NIROOZbx/notification-engine/internal/metrics"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func NewTimeLoggerMiddleware(log zerolog.Logger,metrics *metrics.Metrics) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		path := "unknown"
		if route := c.Route(); route != nil {
			path = route.Path
		}
		method := c.Method()
		status := c.Response().StatusCode()
		statusStr := strconv.Itoa(status)


		metrics.HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()

		return err
	}
}
