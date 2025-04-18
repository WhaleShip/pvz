package http_handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type authService interface {
	RegisterUser(ctx context.Context, req oapi.PostRegisterJSONRequestBody) (oapi.User, error)
	LoginUser(ctx context.Context, req oapi.PostLoginJSONRequestBody) (string, error)
	DummyLogin(req oapi.PostDummyLoginJSONRequestBody) (string, error)
}

type AuthHandler struct {
	authService authService
}

func NewAuthHandler(authSvc authService) *AuthHandler {
	return &AuthHandler{authService: authSvc}
}

func (h *AuthHandler) PostRegister(c *fiber.Ctx) error {
	var req oapi.PostRegisterJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := h.authService.RegisterUser(c.UserContext(), req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	c.Status(fiber.StatusCreated)
	return c.JSON(user)
}

func (h *AuthHandler) PostLogin(c *fiber.Ctx) error {
	var req oapi.PostLoginJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.authService.LoginUser(c.UserContext(), req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	return c.JSON(token)
}

func (h *AuthHandler) PostDummyLogin(c *fiber.Ctx) error {
	var req oapi.PostDummyLoginJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.authService.DummyLogin(req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	return c.JSON(token)
}
