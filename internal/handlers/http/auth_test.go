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

type mockAuthService struct{ mock.Mock }

func (m *mockAuthService) RegisterUser(ctx context.Context, req oapi.PostRegisterJSONRequestBody) (oapi.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(oapi.User), args.Error(1)
}

func (m *mockAuthService) LoginUser(ctx context.Context, req oapi.PostLoginJSONRequestBody) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) DummyLogin(req oapi.PostDummyLoginJSONRequestBody) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func TestPostRegister(t *testing.T) {
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)
	app := fiber.New()
	app.Post("/register", h.PostRegister)

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`not json`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		body := oapi.PostRegisterJSONRequestBody{Email: "e@e.com", Password: "p", Role: oapi.Employee}
		mockSvc.On("RegisterUser", mock.Anything, body).Return(oapi.User{}, errors.New("boom"))
		req := httptest.NewRequest(http.MethodPost, "/register", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		body := oapi.PostRegisterJSONRequestBody{Email: "x@y.com", Password: "pwd", Role: oapi.Moderator}
		want := oapi.User{Id: ptrUUID(uuid.New()), Email: body.Email, Role: oapi.UserRole(body.Role)}
		mockSvc.On("RegisterUser", mock.Anything, body).Return(want, nil)
		req := httptest.NewRequest(http.MethodPost, "/register", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var got oapi.User
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}

func TestPostLogin(t *testing.T) {
	t.Run("bad body", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		h := NewAuthHandler(mockSvc)
		app := fiber.New()
		app.Post("/login", h.PostLogin)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`not-json`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		h := NewAuthHandler(mockSvc)
		app := fiber.New()
		app.Post("/login", h.PostLogin)

		body := oapi.PostLoginJSONRequestBody{Email: "a@b.c", Password: "pw"}
		mockSvc.
			On("LoginUser", mock.Anything, body).
			Return("", errors.New("fail"))

		req := httptest.NewRequest(http.MethodPost, "/login", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		h := NewAuthHandler(mockSvc)
		app := fiber.New()
		app.Post("/login", h.PostLogin)

		body := oapi.PostLoginJSONRequestBody{Email: "a@b.c", Password: "pw"}
		wantToken := "tok123"
		mockSvc.
			On("LoginUser", mock.Anything, body).
			Return(wantToken, nil)

		req := httptest.NewRequest(http.MethodPost, "/login", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var got string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, wantToken, got)

		mockSvc.AssertExpectations(t)
	})
}

func TestPostDummyLogin(t *testing.T) {
	mockSvc := new(mockAuthService)
	h := NewAuthHandler(mockSvc)
	app := fiber.New()
	app.Post("/dummy", h.PostDummyLogin)

	t.Run("bad body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dummy", bytes.NewBufferString(`%%%`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("service error", func(t *testing.T) {
		body := oapi.PostDummyLoginJSONRequestBody{Role: oapi.PostDummyLoginJSONBodyRoleEmployee}
		mockSvc.On("DummyLogin", body).Return("", errors.New("boom"))
		req := httptest.NewRequest(http.MethodPost, "/dummy", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		body := oapi.PostDummyLoginJSONRequestBody{Role: oapi.PostDummyLoginJSONBodyRoleModerator}
		want := "dummytoken"
		mockSvc.On("DummyLogin", body).Return(want, nil)
		req := httptest.NewRequest(http.MethodPost, "/dummy", marshaled(t, body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var got string
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		require.Equal(t, want, got)
		mockSvc.AssertExpectations(t)
	})
}
