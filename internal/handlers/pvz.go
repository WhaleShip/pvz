package handlers

import (
	"github.com/gofiber/fiber/v2"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/service"
)

type PVZHandler struct {
	pvzService service.PVZService
}

func NewPVZHandler(pvzSvc service.PVZService) *PVZHandler {
	return &PVZHandler{pvzService: pvzSvc}
}

func (h *PVZHandler) PostPvz(c *fiber.Ctx) error {
	var req gen.PostPvzJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	pvz, err := h.pvzService.CreatePVZ(req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(pvz)
}

func (h *PVZHandler) GetPvz(c *fiber.Ctx) error {
	return nil
}
