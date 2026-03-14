package middleware

import (
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
)

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return response.Forbidden(c, nil, "role not found")
		}
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next() 
			}
		}

		return response.Forbidden(c, nil, "insufficient permissions")
	}
}