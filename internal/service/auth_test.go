package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/utils"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) InsertUser(ctx context.Context, id uuid.UUID, email, password, role string) error {
	args := m.Called(ctx, id, email, password, role)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, string, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(uuid.UUID), args.String(1), args.String(2), args.Error(3)
}

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockUserRepo)
	svc := NewAuthService(mockRepo)

	t.Run("invalid role", func(t *testing.T) {
		req := oapi.PostRegisterJSONRequestBody{Email: "vozmiteVAvito@pj.com", Password: "pwd", Role: "invalid"}
		_, err := svc.RegisterUser(ctx, req)
		require.ErrorIs(t, err, pvz_errors.ErrInvalidRole)
	})

	t.Run("insert error", func(t *testing.T) {
		req := oapi.PostRegisterJSONRequestBody{Email: "vozmiteVAvito@pj.com", Password: "pwd", Role: oapi.Employee}
		hashed, _ := utils.HashPassword(req.Password)
		mockRepo.
			On("InsertUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), string(req.Email), hashed, string(req.Role)).
			Return(errors.New("db fail"))

		_, err := svc.RegisterUser(ctx, req)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		req := oapi.PostRegisterJSONRequestBody{Email: "vozmiteVAvito@pj.com", Password: "pwd", Role: oapi.Moderator}
		hashed, _ := utils.HashPassword(req.Password)
		var capturedID uuid.UUID
		mockRepo.
			On("InsertUser", mock.Anything, mock.AnythingOfType("uuid.UUID"), string(req.Email), hashed, string(req.Role)).
			Run(func(args mock.Arguments) {
				capturedID = args.Get(1).(uuid.UUID)
			}).
			Return(nil)

		user, err := svc.RegisterUser(ctx, req)
		require.NoError(t, err)
		require.Equal(t, req.Email, user.Email)
		require.Equal(t, oapi.UserRole(req.Role), user.Role)
		require.Equal(t, &capturedID, user.Id)
		mockRepo.AssertExpectations(t)
	})
}

func TestLoginUser(t *testing.T) {
	ctx := context.Background()

	t.Run("repo error", func(t *testing.T) {
		mockRepo := new(mockUserRepo)
		svc := NewAuthService(mockRepo)

		req := oapi.PostLoginJSONRequestBody{Email: "x@y.com", Password: "pw"}
		mockRepo.
			On("GetUserByEmail", mock.Anything, string(req.Email)).
			Return(uuid.Nil, "", "", errors.New("not found")).
			Once()

		_, err := svc.LoginUser(ctx, req)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo := new(mockUserRepo)
		svc := NewAuthService(mockRepo)

		req := oapi.PostLoginJSONRequestBody{Email: "x@y.com", Password: "pw"}
		otherHash, _ := utils.HashPassword("other")
		mockRepo.
			On("GetUserByEmail", mock.Anything, string(req.Email)).
			Return(uuid.New(), otherHash, string(oapi.Employee), nil).
			Once()

		_, err := svc.LoginUser(ctx, req)
		require.ErrorIs(t, err, pvz_errors.ErrInvalidPassword)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo := new(mockUserRepo)
		svc := NewAuthService(mockRepo)

		req := oapi.PostLoginJSONRequestBody{Email: "x@y.com", Password: "pw"}
		hashed, _ := utils.HashPassword(req.Password)
		userID := uuid.New()
		role := string(oapi.Moderator)

		mockRepo.
			On("GetUserByEmail", mock.Anything, string(req.Email)).
			Return(userID, hashed, role, nil).
			Once()

		token, err := svc.LoginUser(ctx, req)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})
}

func TestDummyLogin(t *testing.T) {
	svc := NewAuthService(nil)

	t.Run("invalid role", func(t *testing.T) {
		_, err := svc.DummyLogin(oapi.PostDummyLoginJSONRequestBody{Role: "bad"})
		require.ErrorIs(t, err, pvz_errors.ErrInvalidRole)
	})

	t.Run("employee success", func(t *testing.T) {
		token, err := svc.DummyLogin(oapi.PostDummyLoginJSONRequestBody{Role: oapi.PostDummyLoginJSONBodyRoleEmployee})
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("moderator success", func(t *testing.T) {
		token, err := svc.DummyLogin(oapi.PostDummyLoginJSONRequestBody{Role: oapi.PostDummyLoginJSONBodyRoleModerator})
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})
}
