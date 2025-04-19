package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRoleMiddleware(t *testing.T) {
	t.Run("allowed role", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("role", "employee")
			return c.Next()
		})
		app.Use(RoleMiddleware("employee", "moderator"))
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("missing role", func(t *testing.T) {
		app := fiber.New()
		app.Use(RoleMiddleware("any"))
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("forbidden role", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("role", "guest")
			return c.Next()
		})
		app.Use(RoleMiddleware("admin"))
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
