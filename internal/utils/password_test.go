package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pass := "secure_password"
		hashed, err := HashPassword(pass)
		require.NoError(t, err)
		require.NotEmpty(t, hashed)
	})
}

func TestIsCorrectPassword(t *testing.T) {
	t.Run("correct", func(t *testing.T) {
		password := "12345"
		hashed, err := HashPassword(password)
		require.NoError(t, err)
		require.True(t, IsCorrectPassword(hashed, password))
	})

	t.Run("incorrect", func(t *testing.T) {
		password := "12345qwertyabvgdeyojziiyklmnoprstNUVOZMITEVAVITOPJPJPJufhcchshschayuya"
		hashed, err := HashPassword(password)
		require.NoError(t, err)
		require.False(t, IsCorrectPassword(hashed, "wrong_password"))
	})
}
