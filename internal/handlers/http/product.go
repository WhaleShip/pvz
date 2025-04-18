package http_handlers

import (
	"github.com/gofiber/fiber/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/service"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(prodSvc service.ProductService) *ProductHandler {
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
