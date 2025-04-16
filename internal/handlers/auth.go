package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authSvc}
}

func (h *AuthHandler) PostDummyLogin(c *fiber.Ctx) error {
	var req gen.PostDummyLoginJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.authService.DummyLogin(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(token)
}

func (h *AuthHandler) PostRegister(c *fiber.Ctx) error {
	var req gen.PostRegisterJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := h.authService.RegisterUser(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	c.Status(fiber.StatusCreated)
	return c.JSON(user)
}

func (h *AuthHandler) PostLogin(c *fiber.Ctx) error {
	var req gen.PostLoginJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	token, err := h.authService.LoginUser(c.UserContext(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(token)
}
