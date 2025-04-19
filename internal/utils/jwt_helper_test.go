package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/whaleship/pvz/internal/config"
)

func TestGenerateJWT(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		token, err := GenerateJWT(userID, "xd")
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})
}

func TestParseJWTToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		role := "manager"
		token, err := GenerateJWT(userID, role)
		require.NoError(t, err)

		parsedID, parsedRole, err := ParseJWTToken(token)
		require.NoError(t, err)
		require.Equal(t, userID, parsedID)
		require.Equal(t, role, parsedRole)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, _, err := ParseJWTToken("vronge tokina")
		require.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "some-id",
			"role":    "chertila",
			"exp":     time.Now().Add(-time.Hour).Unix(),
			"iat":     time.Now().Add(-2 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(config.GetJWTSecret())
		require.NoError(t, err)

		_, _, err = ParseJWTToken(signed)
		require.Error(t, err)
	})
	t.Run("invalid userID", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "not-a-uuid",
			"role":    "some-role",
			"exp":     time.Now().Add(time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(config.GetJWTSecret())
		require.NoError(t, err)

		uid, role, err := ParseJWTToken(signed)
		require.Error(t, err)
		require.Equal(t, uuid.Nil, uid)
		require.Empty(t, role)
	})
}
