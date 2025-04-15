package handlers

import (
	"github.com/gofiber/fiber/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/whaleship/pvz/internal/service"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(prodSvc service.ProductService) *ProductHandler {
	return &ProductHandler{productService: prodSvc}
}

func (h *ProductHandler) PostProducts(c *fiber.Ctx) error {
	return nil
}

func (h *ProductHandler) PostPvzPvzIdDeleteLastProduct(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	return nil
}
