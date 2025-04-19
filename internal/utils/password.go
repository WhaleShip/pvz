package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func IsCorrectPassword(hashedPassword, password string) bool {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hashedPassword == hex.EncodeToString(hash.Sum(nil))
}
