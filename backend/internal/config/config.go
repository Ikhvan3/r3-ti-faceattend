package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv           string
	AppPort          string
	BusinessTimezone string
	Database         DatabaseConfig
	Auth             AuthConfig
	Geofence         GeofenceConfig
	FaceVerification FaceVerificationConfig
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

type GeofenceConfig struct {
	MaxAccuracyMeters float64
}

type FaceVerificationConfig struct {
	Threshold float64
}

func Load() Config {
	return Config{
		AppEnv:           env("APP_ENV", "local"),
		AppPort:          env("APP_PORT", "8080"),
		BusinessTimezone: env("BUSINESS_TIMEZONE", "Asia/Jakarta"),
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
		Geofence: GeofenceConfig{
			MaxAccuracyMeters: envFloat("GEOFENCE_MAX_ACCURACY_METERS", 50),
		},
		FaceVerification: FaceVerificationConfig{
			Threshold: envRequiredFloat("FACE_VERIFICATION_THRESHOLD"),
		},
	}
}

func (c Config) BusinessLocation() (*time.Location, error) {
	location, err := time.LoadLocation(strings.TrimSpace(c.BusinessTimezone))
	if err != nil {
		return nil, fmt.Errorf("invalid BUSINESS_TIMEZONE: %w", err)
	}

	return location, nil
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

func (c GeofenceConfig) Validate() error {
	if math.IsNaN(c.MaxAccuracyMeters) || math.IsInf(c.MaxAccuracyMeters, 0) || c.MaxAccuracyMeters <= 0 {
		return errors.New("GEOFENCE_MAX_ACCURACY_METERS must be finite and positive")
	}

	return nil
}

func (c FaceVerificationConfig) Validate() error {
	if math.IsNaN(c.Threshold) || math.IsInf(c.Threshold, 0) || c.Threshold < -1 || c.Threshold > 1 {
		return errors.New("FACE_VERIFICATION_THRESHOLD is required and must be finite between -1 and 1")
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

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return math.NaN()
	}

	return parsed
}

func envRequiredFloat(key string) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return math.NaN()
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return math.NaN()
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
