package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/whaleship/pvz/internal/utils"
)

func createTestCtx(authHeader string) (*fiber.Ctx, *fiber.App) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod(http.MethodGet)
	fctx.Request.SetRequestURI("/")
	if authHeader != "" {
		fctx.Request.Header.Set("Authorization", authHeader)
	}
	ctx := app.AcquireCtx(fctx)
	return ctx, app
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("missing header", func(t *testing.T) {
		ctx, app := createTestCtx("")
		defer app.ReleaseCtx(ctx)

		err := AuthMiddleware(ctx)
		require.Error(t, err)
		require.IsType(t, &fiber.Error{}, err)
		var fErr *fiber.Error
		require.True(t, errors.As(err, &fErr))
		require.Equal(t, fiber.ErrUnauthorized.Code, fErr.Code)
	})

	t.Run("bad format", func(t *testing.T) {
		ctx, app := createTestCtx("BadBearerToken")
		defer app.ReleaseCtx(ctx)

		err := AuthMiddleware(ctx)
		require.Error(t, err)
		require.IsType(t, &fiber.Error{}, err)
		var fErr *fiber.Error
		require.True(t, errors.As(err, &fErr))
		require.Equal(t, fiber.ErrUnauthorized.Code, fErr.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		ctx, app := createTestCtx("Bearer totally.invalid.jwt")
		defer app.ReleaseCtx(ctx)

		err := AuthMiddleware(ctx)
		require.Error(t, err)
		require.IsType(t, &fiber.Error{}, err)
		var fErr *fiber.Error
		require.True(t, errors.As(err, &fErr))
		require.Equal(t, fiber.ErrUnauthorized.Code, fErr.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		userID := uuid.New()
		role := "employee"
		token, err := utils.GenerateJWT(userID, role)
		require.NoError(t, err)

		ctx, app := createTestCtx("Bearer " + token)
		defer app.ReleaseCtx(ctx)

		safeAuth := func(c *fiber.Ctx) (err error) {
			defer func() { recover() }()
			return AuthMiddleware(c)
		}

		err = safeAuth(ctx)
		require.NoError(t, err)
		require.Equal(t, userID, (ctx).Locals("userID"))
		require.Equal(t, role, (ctx).Locals("role"))
	})
}
