package middleware

import (
	"strings"

	pvz_errors "github.com/whaleship/pvz/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/utils"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return fiber.NewError(fiber.StatusUnauthorized, pvz_errors.ErrMissingAuthHeader.Error())
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return fiber.NewError(fiber.StatusUnauthorized, pvz_errors.ErrInvalidAuthHeader.Error())
	}
	userID, role, err := utils.ParseJWTToken(parts[1])
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, pvz_errors.ErrInvalidToken.Error()+err.Error())
	}
	c.Locals("userID", userID)
	c.Locals("role", role)
	return c.Next()
}
