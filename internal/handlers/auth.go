package handlers

import (
	"github.com/gofiber/fiber/v2"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
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
