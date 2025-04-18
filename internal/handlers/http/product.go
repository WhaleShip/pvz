package http_handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type productService interface {
	AddProduct(ctx context.Context, req oapi.PostProductsJSONRequestBody) (oapi.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

type ProductHandler struct {
	productService productService
}

func NewProductHandler(prodSvc productService) *ProductHandler {
	return &ProductHandler{productService: prodSvc}
}

func (h *ProductHandler) PostProducts(c *fiber.Ctx) error {
	var req oapi.PostProductsJSONRequestBody
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	product, err := h.productService.AddProduct(c.UserContext(), req)
	if err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

func (h *ProductHandler) PostPvzPvzIdDeleteLastProduct(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	if err := h.productService.DeleteLastProduct(c.UserContext(), pvzId); err != nil {
		status := pvz_errors.GetErrorStatusCode(err)
		return fiber.NewError(status, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}
