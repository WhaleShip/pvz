package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pass := "secure_password"
		hashed := HashPassword(pass)
		require.NotEmpty(t, hashed)
	})
}

func TestIsCorrectPassword(t *testing.T) {
	t.Run("correct", func(t *testing.T) {
		password := "12345"
		hashed := HashPassword(password)
		require.True(t, IsCorrectPassword(hashed, password))
	})

	t.Run("incorrect", func(t *testing.T) {
		password := "12345qwertyabvgdeyojziiyklmnoprstNUVOZMITEVAVITOPJPJPJufhcchshschayuya"
		hashed := HashPassword(password)
		require.False(t, IsCorrectPassword(hashed, "wrong_password"))
	})
}
