package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/service"
)

type PVZHandler struct {
	pvzService service.PVZService
}

func NewPVZHandler(pvzSvc service.PVZService) *PVZHandler {
	return &PVZHandler{pvzService: pvzSvc}
}

func (h *PVZHandler) PostPvz(c *fiber.Ctx) error {
	return nil
}

func (h *PVZHandler) GetPvz(c *fiber.Ctx) error {
	return nil
}
