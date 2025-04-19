package http_handlers

import (
	"bytes"
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

type mockReceptionService struct{ mock.Mock }

func (m *mockReceptionService) CreateReception(req oapi.PostReceptionsJSONRequestBody) (oapi.Reception, error) {
	args := m.Called(req)
	return args.Get(0).(oapi.Reception), args.Error(1)
}
func (m *mockReceptionService) CloseLastReception(pvzID uuid.UUID) (oapi.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(oapi.Reception), args.Error(1)
}

func TestPostReception(t *testing.T) {
	mockSvc := new(mockReceptionService)
	h := NewReceptionHandler(mockSvc)
	app := fiber.New()
	app.Post("/receptions", h.PostReception)

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBufferString(`???`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		body := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}
		mockSvc.On("CreateReception", body).Return(oapi.Reception{}, errors.New("boom"))
		req := httptest.NewRequest(http.MethodPost, "/receptions", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		body := oapi.PostReceptionsJSONRequestBody{PvzId: uuid.New()}
		want := oapi.Reception{Id: ptrUUID(uuid.New())}
		mockSvc.On("CreateReception", body).Return(want, nil)
		req := httptest.NewRequest(http.MethodPost, "/receptions", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
		var got oapi.Reception
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}

func TestCloseReception(t *testing.T) {
	mockSvc := new(mockReceptionService)
	h := NewReceptionHandler(mockSvc)
	app := fiber.New()
	app.Post("/receptions/:pvzId/close", func(c *fiber.Ctx) error {
		return h.CloseReception(c, uuid.MustParse(c.Params("pvzId")))
	})

	t.Run("service error", func(t *testing.T) {
		id := uuid.New()
		mockSvc.On("CloseLastReception", id).Return(oapi.Reception{}, errors.New("boom"))
		req := httptest.NewRequest(http.MethodPost, "/receptions/"+id.String()+"/close", nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		want := oapi.Reception{Id: ptrUUID(uuid.New())}
		mockSvc.On("CloseLastReception", id).Return(want, nil)
		req := httptest.NewRequest(http.MethodPost, "/receptions/"+id.String()+"/close", nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		var got oapi.Reception
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}
