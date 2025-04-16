package middleware

import (
	"github.com/gofiber/fiber/v2"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
)

func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, pvz_errors.ErrInvalidRole.Error())
		}
		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, pvz_errors.ErrInsufficientPermissions.Error())
	}
}
