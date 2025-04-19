package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/whaleship/pvz/internal/utils"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("missing header", func(t *testing.T) {
		app := fiber.New()
		app.Use(AuthMiddleware)
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("bad format", func(t *testing.T) {
		app := fiber.New()
		app.Use(AuthMiddleware)
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "BadBearerToken")
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		app := fiber.New()
		app.Use(AuthMiddleware)
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer totally.invalid.jwt")
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid token", func(t *testing.T) {
		userID := uuid.New()
		role := "employee"
		token, err := utils.GenerateJWT(userID, role)
		require.NoError(t, err)

		app := fiber.New()
		app.Use(AuthMiddleware)
		app.Get("/", func(c *fiber.Ctx) error {
			require.Equal(t, userID, c.Locals("userID"))
			require.Equal(t, role, c.Locals("role"))
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
