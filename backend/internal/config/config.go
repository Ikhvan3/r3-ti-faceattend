package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv   string
	AppPort  string
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type AuthConfig struct {
	AccessTokenSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	TokenIssuer       string
	TokenAudience     string
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
		Auth: AuthConfig{
			AccessTokenSecret: os.Getenv("AUTH_ACCESS_TOKEN_SECRET"),
			AccessTokenTTL:    time.Duration(envInt("AUTH_ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
			RefreshTokenTTL:   time.Duration(envInt("AUTH_REFRESH_TOKEN_TTL_HOURS", 168)) * time.Hour,
			TokenIssuer:       env("AUTH_TOKEN_ISSUER", "r3-ti-faceattend-api"),
			TokenAudience:     env("AUTH_TOKEN_AUDIENCE", "r3-ti-faceattend-client"),
		},
	}
}

func (c AuthConfig) Validate() error {
	if strings.TrimSpace(c.AccessTokenSecret) == "" {
		return errors.New("AUTH_ACCESS_TOKEN_SECRET is required")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("AUTH_ACCESS_TOKEN_TTL_MINUTES must be positive")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("AUTH_REFRESH_TOKEN_TTL_HOURS must be positive")
	}
	if strings.TrimSpace(c.TokenIssuer) == "" {
		return errors.New("AUTH_TOKEN_ISSUER is required")
	}
	if strings.TrimSpace(c.TokenAudience) == "" {
		return errors.New("AUTH_TOKEN_AUDIENCE is required")
	}

	return nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func LoadAuthFromEnv() (AuthConfig, error) {
	cfg := Load().Auth
	if err := cfg.Validate(); err != nil {
		return AuthConfig{}, fmt.Errorf("invalid auth config: %w", err)
	}

	return cfg, nil
}
