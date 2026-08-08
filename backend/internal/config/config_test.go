package config

import (
	"math"
	"os"
	"path/filepath"
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

func TestGeofenceConfigValidateAcceptsFinitePositiveAccuracy(t *testing.T) {
	cfg := GeofenceConfig{MaxAccuracyMeters: 50}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestGeofenceConfigValidateRejectsInvalidAccuracy(t *testing.T) {
	tests := []float64{0, -1, math.Inf(1), math.NaN()}

	for _, value := range tests {
		t.Run("invalid", func(t *testing.T) {
			cfg := GeofenceConfig{MaxAccuracyMeters: value}
			if err := cfg.Validate(); err == nil {
				t.Fatalf("Validate() error = nil, want error for %v", value)
			}
		})
	}
}

func TestLoadDefaultsGeofenceMaxAccuracy(t *testing.T) {
	t.Setenv("GEOFENCE_MAX_ACCURACY_METERS", "")

	cfg := Load()

	if cfg.Geofence.MaxAccuracyMeters != 50 {
		t.Fatalf("MaxAccuracyMeters = %v, want 50", cfg.Geofence.MaxAccuracyMeters)
	}
}

func TestLoadRejectsInvalidGeofenceMaxAccuracyThroughValidate(t *testing.T) {
	t.Setenv("GEOFENCE_MAX_ACCURACY_METERS", "not-a-number")

	cfg := Load()

	if err := cfg.Geofence.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestFaceVerificationConfigValidateAcceptsCosineThreshold(t *testing.T) {
	cfg := FaceVerificationConfig{Threshold: 0.8}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestFaceVerificationConfigValidateRejectsMissingOrOutOfRangeThreshold(t *testing.T) {
	tests := []float64{math.NaN(), math.Inf(1), -1.1, 1.1}

	for _, value := range tests {
		t.Run("invalid", func(t *testing.T) {
			cfg := FaceVerificationConfig{Threshold: value}
			if err := cfg.Validate(); err == nil {
				t.Fatalf("Validate() error = nil, want error for %v", value)
			}
		})
	}
}

func TestLoadRequiresFaceVerificationThresholdThroughValidate(t *testing.T) {
	t.Setenv("FACE_VERIFICATION_THRESHOLD", "")

	cfg := Load()

	if err := cfg.FaceVerification.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestLoadReadsFaceVerificationThreshold(t *testing.T) {
	t.Setenv("FACE_VERIFICATION_THRESHOLD", "0.82")

	cfg := Load()

	if cfg.FaceVerification.Threshold != 0.82 {
		t.Fatalf("Threshold = %v, want 0.82", cfg.FaceVerification.Threshold)
	}
}

func TestLoadDotEnvReadsFileWithoutOverridingExistingEnv(t *testing.T) {
	originalDBHost, hadDBHost := os.LookupEnv("DB_HOST")
	if err := os.Unsetenv("DB_HOST"); err != nil {
		t.Fatalf("Unsetenv(DB_HOST) error = %v", err)
	}
	t.Cleanup(func() {
		if hadDBHost {
			_ = os.Setenv("DB_HOST", originalDBHost)
			return
		}
		_ = os.Unsetenv("DB_HOST")
	})

	t.Setenv("AUTH_ACCESS_TOKEN_SECRET", "from-shell")
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(
		envPath,
		[]byte("AUTH_ACCESS_TOKEN_SECRET=from-file\nDB_HOST=db.local\n"),
		0o600,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := LoadDotEnv(envPath); err != nil {
		t.Fatalf("LoadDotEnv() error = %v", err)
	}

	if got := os.Getenv("AUTH_ACCESS_TOKEN_SECRET"); got != "from-shell" {
		t.Fatalf("AUTH_ACCESS_TOKEN_SECRET = %q, want from-shell", got)
	}
	if got := os.Getenv("DB_HOST"); got != "db.local" {
		t.Fatalf("DB_HOST = %q, want db.local", got)
	}
}

func TestLoadDotEnvIgnoresMissingFile(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")

	if err := LoadDotEnv(envPath); err != nil {
		t.Fatalf("LoadDotEnv() error = %v", err)
	}
}
