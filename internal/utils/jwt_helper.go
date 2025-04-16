package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/whaleship/pvz/internal/config"
)

type claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID, role string) (string, error) {
	now := time.Now()
	claims := claims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.TokenValidityPeriod)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.GetJWTSecret())
}

func ParseJWTToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return config.GetJWTSecret(), nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return uid, claims.Role, nil
}
