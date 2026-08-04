package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv   string
	AppPort  string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load() Config {
	return Config{
		AppEnv:  env("APP_ENV", "local"),
		AppPort: env("APP_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			Name:     env("DB_NAME", "r3_ti_faceattend"),
			User:     env("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
