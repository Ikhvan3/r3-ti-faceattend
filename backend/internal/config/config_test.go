package config

import (
	"math"
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
