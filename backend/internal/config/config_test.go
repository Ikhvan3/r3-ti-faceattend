package config

import (
	"testing"
	"time"
)

func TestAuthConfigValidateAcceptsValidConfig(t *testing.T) {
	cfg := AuthConfig{
		AccessTokenSecret: "secret",
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   24 * time.Hour,
		TokenIssuer:       "issuer",
		TokenAudience:     "audience",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestAuthConfigValidateRejectsEmptySecret(t *testing.T) {
	cfg := AuthConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		TokenIssuer:     "issuer",
		TokenAudience:   "audience",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestAuthConfigValidateRejectsInvalidTTL(t *testing.T) {
	cfg := AuthConfig{
		AccessTokenSecret: "secret",
		AccessTokenTTL:    0,
		RefreshTokenTTL:   24 * time.Hour,
		TokenIssuer:       "issuer",
		TokenAudience:     "audience",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestConfigBusinessLocation(t *testing.T) {
	cfg := Config{BusinessTimezone: "Asia/Jakarta"}

	location, err := cfg.BusinessLocation()
	if err != nil {
		t.Fatalf("BusinessLocation() error = %v", err)
	}
	if location.String() != "Asia/Jakarta" {
		t.Fatalf("location = %s, want Asia/Jakarta", location)
	}
}

func TestLoadDefaultsBusinessTimezoneToAsiaJakarta(t *testing.T) {
	t.Setenv("BUSINESS_TIMEZONE", "")

	cfg := Load()

	if cfg.BusinessTimezone != "Asia/Jakarta" {
		t.Fatalf("BusinessTimezone = %q, want Asia/Jakarta", cfg.BusinessTimezone)
	}
}

func TestConfigBusinessLocationRejectsInvalidTimezone(t *testing.T) {
	cfg := Config{BusinessTimezone: "Invalid/Timezone"}

	if _, err := cfg.BusinessLocation(); err == nil {
		t.Fatal("BusinessLocation() error = nil, want error")
	}
}
