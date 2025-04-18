package service

import (
	"context"

	"github.com/google/uuid"
	pvz_errors "github.com/whaleship/pvz/internal/errors"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/repository"
	"github.com/whaleship/pvz/internal/utils"
)

type AuthService interface {
	RegisterUser(ctx context.Context, req oapi.PostRegisterJSONRequestBody) (oapi.User, error)
	LoginUser(ctx context.Context, req oapi.PostLoginJSONRequestBody) (string, error)
	DummyLogin(req oapi.PostDummyLoginJSONRequestBody) (string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) RegisterUser(ctx context.Context, req oapi.PostRegisterJSONRequestBody) (oapi.User, error) {
	if req.Role != oapi.Employee && req.Role != oapi.Moderator {
		return oapi.User{}, pvz_errors.ErrInvalidRole
	}
	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return oapi.User{}, err
	}
	newUserID := uuid.New()
	if err := s.userRepo.InsertUser(ctx, newUserID, string(req.Email), hashedPass, string(req.Role)); err != nil {
		return oapi.User{}, err
	}
	return oapi.User{
		Id:    &newUserID,
		Email: req.Email,
		Role:  oapi.UserRole(req.Role),
	}, nil
}

func (s *authService) LoginUser(ctx context.Context, req oapi.PostLoginJSONRequestBody) (string, error) {
	id, hashed, role, err := s.userRepo.GetUserByEmail(ctx, string(req.Email))
	if err != nil {
		return "", err
	}
	if !utils.IsCorrectPassword(hashed, req.Password) {
		return "", pvz_errors.ErrInvalidPassword
	}
	token, err := utils.GenerateJWT(id, role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) DummyLogin(req oapi.PostDummyLoginJSONRequestBody) (string, error) {
	role := req.Role
	if role != oapi.PostDummyLoginJSONBodyRoleModerator && role != oapi.PostDummyLoginJSONBodyRoleEmployee {
		return "", pvz_errors.ErrInvalidRole
	}

	token, err := utils.GenerateJWT(uuid.New(), string(role))
	if err != nil {
		return "", err
	}
	return token, nil
}
