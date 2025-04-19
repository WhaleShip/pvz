package http_handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/gen/oapi"
)

type mockProductService struct{ mock.Mock }

func (m *mockProductService) AddProduct(
	ctx context.Context,
	req oapi.PostProductsJSONRequestBody) (oapi.Product, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(oapi.Product), args.Error(1)
}
func (m *mockProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	return m.Called(ctx, pvzID).Error(0)
}

func TestPostProducts(t *testing.T) {
	mockSvc := new(mockProductService)
	h := NewProductHandler(mockSvc)
	app := fiber.New()
	app.Post("/products", h.PostProducts)

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`!!!`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		body := oapi.PostProductsJSONRequestBody{PvzId: uuid.New(), Type: oapi.PostProductsJSONBodyType("X")}
		mockSvc.On("AddProduct", mock.Anything, body).Return(oapi.Product{}, errors.New("boom"))
		req := httptest.NewRequest(http.MethodPost, "/products", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		body := oapi.PostProductsJSONRequestBody{PvzId: uuid.New(), Type: oapi.PostProductsJSONBodyType("X")}
		want := oapi.Product{Id: ptrUUID(uuid.New())}
		mockSvc.On("AddProduct", mock.Anything, body).Return(want, nil)
		req := httptest.NewRequest(http.MethodPost, "/products", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var got oapi.Product
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	mockSvc := new(mockProductService)
	h := NewProductHandler(mockSvc)
	app := fiber.New()
	app.Delete("/products/:pvzId", func(c *fiber.Ctx) error {
		return h.PostPvzPvzIdDeleteLastProduct(c, uuid.MustParse((c.Params("pvzId"))))
	})

	t.Run("service error", func(t *testing.T) {
		id := uuid.New()
		mockSvc.On("DeleteLastProduct", mock.Anything, id).Return(errors.New("boom"))
		req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		mockSvc.On("DeleteLastProduct", mock.Anything, id).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/products/"+id.String(), nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})
}
