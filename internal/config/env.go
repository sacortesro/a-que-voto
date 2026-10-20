package config

import (
	"errors"
	"os"
)

func RequireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", errors.New(key + " environment variable not set")
	}
	return value, nil
}
