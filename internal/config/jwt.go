package config

import (
	"os"
	"sync"
)

var jwtOnce sync.Once

func initJWTSecret() {
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		secretStr = "secret"
	}
	jwtSecret = []byte(secretStr)
}

func GetJWTSecret() []byte {
	jwtOnce.Do(initJWTSecret)
	return jwtSecret
}
