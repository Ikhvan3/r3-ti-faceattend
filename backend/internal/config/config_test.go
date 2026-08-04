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
