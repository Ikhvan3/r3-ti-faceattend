package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"r3-ti-faceattend/backend/internal/user"
)

func TestAuthenticateAcceptsValidBearerToken(t *testing.T) {
	verifier := fakeTokenVerifier{claims: Claims{Role: user.RoleAdmin}}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		if _, ok := ClaimsFromContext(r.Context()); !ok {
			t.Fatal("claims missing from context")
		}
	})
	handler := Authenticate(verifier, next)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !nextCalled {
		t.Fatal("next handler was not called")
	}
}

func TestAuthenticateRejectsMissingMalformedAndExpiredToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing", header: ""},
		{name: "malformed", header: "Token value"},
		{name: "empty bearer", header: "Bearer "},
		{name: "expired", header: "Bearer expired-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Authenticate(fakeTokenVerifier{err: ErrInvalidToken}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("next handler should not be called")
			}))
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRequireRoleAcceptsAdminAndRejectsUser(t *testing.T) {
	adminClaims := Claims{Role: user.RoleAdmin}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	adminRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	adminRequest = adminRequest.WithContext(contextWithClaims(adminRequest.Context(), adminClaims))
	adminResponse := httptest.NewRecorder()

	RequireRole(user.RoleAdmin, next).ServeHTTP(adminResponse, adminRequest)

	if adminResponse.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want %d", adminResponse.Code, http.StatusOK)
	}

	userRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	userRequest = userRequest.WithContext(contextWithClaims(userRequest.Context(), Claims{Role: user.RoleUser}))
	userResponse := httptest.NewRecorder()

	RequireRole(user.RoleAdmin, next).ServeHTTP(userResponse, userRequest)

	if userResponse.Code != http.StatusForbidden {
		t.Fatalf("user status = %d, want %d", userResponse.Code, http.StatusForbidden)
	}
}

type fakeTokenVerifier struct {
	claims Claims
	err    error
}

func (v fakeTokenVerifier) VerifyAccessToken(token string) (Claims, error) {
	if v.err != nil || token == "expired-token" {
		return Claims{}, ErrInvalidToken
	}
	return v.claims, nil
}

func contextWithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}
