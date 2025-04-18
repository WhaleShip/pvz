package http_handlers

import (
	"github.com/gofiber/fiber/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/service"
)

type ReceptionHandler struct {
	receptionService service.ReceptionService
}

func NewReceptionHandler(receptSvc service.ReceptionService) *ReceptionHandler {
	return &ReceptionHandler{receptionService: receptSvc}
}

func (h *ReceptionHandler) PostReception(c *fiber.Ctx) error {
	var req oapi.PostReceptionsJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result, err := h.receptionService.CreateReception(req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *ReceptionHandler) CloseReception(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	result, err := h.receptionService.CloseLastReception(pvzId)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	return c.JSON(result)
}
