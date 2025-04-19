package http_handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/whaleship/pvz/internal/dto"
	"github.com/whaleship/pvz/internal/gen/oapi"
)

type mockPVZService struct{ mock.Mock }

func (m *mockPVZService) CreatePVZ(ctx context.Context, req oapi.PostPvzJSONRequestBody) (oapi.PVZ, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(oapi.PVZ), args.Error(1)
}

func (m *mockPVZService) GetPVZ(ctx context.Context, params oapi.GetPvzParams) ([]dto.PVZWithReceptions, error) {
	args := m.Called(ctx, params)
	var res []dto.PVZWithReceptions
	if ans := args.Get(0); ans != nil {
		res = ans.([]dto.PVZWithReceptions)
	} else {
		res = []dto.PVZWithReceptions{}
	}
	return res, args.Error(1)
}

func TestPostPvz(t *testing.T) {
	t.Run("bad body", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Post("/pvz", h.PostPvz)

		req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBufferString(`not-json`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Post("/pvz", h.PostPvz)

		body := oapi.PostPvzJSONRequestBody{City: oapi.Москва}
		mockSvc.
			On("CreatePVZ", mock.Anything, body).
			Return(oapi.PVZ{}, errors.New("insert failed"))

		req := httptest.NewRequest(http.MethodPost, "/pvz", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Post("/pvz", h.PostPvz)

		body := oapi.PostPvzJSONRequestBody{City: oapi.Казань}
		fakeID := uuid.New()
		fakeTime := time.Now()
		mockSvc.
			On("CreatePVZ", mock.Anything, body).
			Return(oapi.PVZ{Id: ptrUUID(fakeID), City: body.City, RegistrationDate: ptrTime(fakeTime)}, nil)

		req := httptest.NewRequest(http.MethodPost, "/pvz", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var got oapi.PVZ
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, body.City, got.City)
		require.Equal(t, fakeID, *got.Id)
		require.WithinDuration(t, fakeTime, *got.RegistrationDate, time.Millisecond)
		mockSvc.AssertExpectations(t)
	})
}

func TestGetPvz(t *testing.T) {
	t.Run("bad query", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Get("/pvz", h.GetPvz)

		req := httptest.NewRequest(http.MethodGet, "/pvz?limit=abc", nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Get("/pvz", h.GetPvz)

		params := oapi.GetPvzParams{Page: ptrInt(1), Limit: ptrInt(5)}
		mockSvc.
			On("GetPVZ", mock.Anything, params).
			Return(nil, errors.New("select failed"))

		req := httptest.NewRequest(http.MethodGet, "/pvz?page=1&limit=5", nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockPVZService)
		h := NewPVZHandler(mockSvc)
		app := fiber.New()
		app.Get("/pvz", h.GetPvz)

		params := oapi.GetPvzParams{Page: ptrInt(2), Limit: ptrInt(3)}
		want := []dto.PVZWithReceptions{
			{Pvz: oapi.PVZ{Id: ptrUUID(uuid.New())}},
		}
		mockSvc.
			On("GetPVZ", mock.Anything, params).
			Return(want, nil)

		req := httptest.NewRequest(http.MethodGet, "/pvz?page=2&limit=3", nil)
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var got []dto.PVZWithReceptions
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}
