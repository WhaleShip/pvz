package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/repository"
	"github.com/whaleship/pvz/internal/utils"
)

type AuthService interface {
	RegisterUser(ctx context.Context, req gen.PostRegisterJSONRequestBody) (gen.User, error)
	LoginUser(ctx context.Context, req gen.PostLoginJSONRequestBody) (string, error)
	DummyLogin(req gen.PostDummyLoginJSONRequestBody) (string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) RegisterUser(ctx context.Context, req gen.PostRegisterJSONRequestBody) (gen.User, error) {
	if req.Role != gen.Employee && req.Role != gen.Moderator {
		return gen.User{}, errors.New("invalid role")
	}
	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return gen.User{}, err
	}
	newUserID := uuid.New()
	if err := s.userRepo.InsertUser(ctx, newUserID, string(req.Email), hashedPass, string(req.Role)); err != nil {
		return gen.User{}, err
	}
	return gen.User{
		Id:    &newUserID,
		Email: req.Email,
		Role:  gen.UserRole(req.Role),
	}, nil
}

func (s *authService) LoginUser(ctx context.Context, req gen.PostLoginJSONRequestBody) (string, error) {
	id, hashed, role, err := s.userRepo.GetUserByEmail(ctx, string(req.Email))
	if err != nil {
		return "", err
	}
	if !utils.IsCorrectPassword(hashed, req.Password) {
		return "", errors.New("invalid password")
	}
	token, err := utils.GenerateJWT(id, role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) DummyLogin(req gen.PostDummyLoginJSONRequestBody) (string, error) {
	role := req.Role
	if role != gen.PostDummyLoginJSONBodyRoleModerator && role != gen.PostDummyLoginJSONBodyRoleEmployee {
		return "", errors.New("invalid token")
	}

	token, err := utils.GenerateJWT(uuid.New(), string(role))
	if err != nil {
		return "", err
	}
	return token, nil
}
