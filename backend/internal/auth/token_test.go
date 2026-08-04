package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/user"
)

func TestAccessTokenValidTokenCanBeVerified(t *testing.T) {
	service := testAccessTokenService("secret", time.Now)

	token, _, err := service.Issue("user-id", "session-id", "admin@example.test", user.RoleAdmin)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := service.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject != "user-id" || claims.SessionID != "session-id" || claims.Role != user.RoleAdmin {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestAccessTokenRejectsWrongSignature(t *testing.T) {
	service := testAccessTokenService("secret", time.Now)
	other := testAccessTokenService("wrong-secret", time.Now)

	token, _, err := other.Issue("user-id", "session-id", "admin@example.test", user.RoleAdmin)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := service.Verify(token); err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func TestAccessTokenRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	service := testAccessTokenService("secret", func() time.Time { return now })
	service.ttl = time.Minute

	token, _, err := service.Issue("user-id", "session-id", "admin@example.test", user.RoleAdmin)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	service.now = func() time.Time { return now.Add(2 * time.Minute) }

	if _, err := service.Verify(token); err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func TestAccessTokenRejectsWrongIssuer(t *testing.T) {
	service := testAccessTokenService("secret", time.Now)
	other := service
	other.issuer = "other-issuer"

	token, _, err := other.Issue("user-id", "session-id", "admin@example.test", user.RoleAdmin)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := service.Verify(token); err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func TestAccessTokenRejectsWrongAudience(t *testing.T) {
	service := testAccessTokenService("secret", time.Now)
	other := service
	other.audience = "other-audience"

	token, _, err := other.Issue("user-id", "session-id", "admin@example.test", user.RoleAdmin)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := service.Verify(token); err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func TestAccessTokenRejectsWrongSigningMethod(t *testing.T) {
	service := testAccessTokenService("secret", time.Now)
	claims := Claims{
		SessionID: "session-id",
		Email:     "admin@example.test",
		Role:      user.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-id",
			Issuer:    service.issuer,
			Audience:  jwt.ClaimStrings{service.audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(service.secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := service.Verify(token); err == nil {
		t.Fatal("Verify() error = nil, want error")
	}
}

func testAccessTokenService(secret string, now func() time.Time) AccessTokenService {
	service := NewAccessTokenService(config.AuthConfig{
		AccessTokenSecret: secret,
		AccessTokenTTL:    15 * time.Minute,
		TokenIssuer:       "issuer",
		TokenAudience:     "audience",
	})
	service.now = now
	return service
}
