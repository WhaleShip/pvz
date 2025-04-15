package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/utils"
)

type AuthService interface {
	DummyLogin(req gen.PostDummyLoginJSONRequestBody) (string, error)
}

type authService struct {
}

func NewAuthService() AuthService {
	return &authService{}
}

func (s *authService) DummyLogin(req gen.PostDummyLoginJSONRequestBody) (string, error) {
	role := req.Role
	if role != gen.PostDummyLoginJSONBodyRoleModerator && role != gen.PostDummyLoginJSONBodyRoleEmployee {
		return "", errors.New("invalid role")
	}

	token, err := utils.GenerateJWT(uuid.New(), string(role))
	if err != nil {
		return "", err
	}
	return token, nil
}
