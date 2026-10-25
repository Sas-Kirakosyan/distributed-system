// Package config reads service configuration from environment variables.
package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string // HTTP port to listen on, e.g. "8080"
	DatabaseURL string // e.g. postgres://user:pass@postgres:5432/orders?sslmode=disable
}

// Load reads config from the environment. DATABASE_URL is required;
// PORT defaults to 8080.
func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
