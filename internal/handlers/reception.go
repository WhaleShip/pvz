package handlers

import (
	"github.com/gofiber/fiber/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/whaleship/pvz/internal/service"
)

type ReceptionHandler struct {
	receptionService service.ReceptionService
}

func NewReceptionHandler(receptSvc service.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{receptionService: receptSvc}
}

func (h *ReceptionHandler) PostReception(c *fiber.Ctx) error {
	return nil
}

func (h *ReceptionHandler) CloseReception(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	return nil
}
