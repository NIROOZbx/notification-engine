package middleware

import (
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
)

func (a *authMiddleware)RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			a.log.Warn().Msg("role middleware: role not found in context")
			return response.Forbidden(c, nil, "role not found")
		}
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next() 
			}
		}

		a.log.Warn().
			Str("userRole", userRole).
			Interface("allowedRoles", allowedRoles).
			Msg("role middleware: insufficient permissions")
		return response.Forbidden(c, nil, "insufficient permissions")
	}
}