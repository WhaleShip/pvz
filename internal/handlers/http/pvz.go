package http_handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/dto"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type pvzService interface {
	CreatePVZ(ctx context.Context, req oapi.PostPvzJSONRequestBody) (oapi.PVZ, error)
	GetPVZ(ctx context.Context, params oapi.GetPvzParams) ([]dto.PVZWithReceptions, error)
}

type PVZHandler struct {
	pvzService pvzService
}

func NewPVZHandler(pvzSvc pvzService) *PVZHandler {
	return &PVZHandler{pvzService: pvzSvc}
}

func (h *PVZHandler) PostPvz(c *fiber.Ctx) error {
	var req oapi.PostPvzJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	pvz, err := h.pvzService.CreatePVZ(c.UserContext(), req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(pvz)
}

func (h *PVZHandler) GetPvz(c *fiber.Ctx) error {
	var params oapi.GetPvzParams
	if err := c.QueryParser(&params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	response, err := h.pvzService.GetPVZ(c.UserContext(), params)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}

	return c.JSON(response)
}
