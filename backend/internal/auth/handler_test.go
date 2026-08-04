package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/user"
)

func TestHandlerLoginValidRequest(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	handler := NewHandler(newTestService(users, newAuthFakeSessionRepository()))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.test","password":"secret"}`))
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
	}
	if strings.Contains(response.Body.String(), "password_hash") {
		t.Fatal("response contains password_hash")
	}
}

func TestHandlerRejectsMalformedJSONEmptyFieldUnknownFieldAndWrongMethod(t *testing.T) {
	users := newAuthFakeUserRepository()
	users.users = append(users.users, activeAdmin())
	handler := NewHandler(newTestService(users, newAuthFakeSessionRepository()))
	tests := []struct {
		name   string
		method string
		body   string
		status int
	}{
		{name: "malformed", method: http.MethodPost, body: `{`, status: http.StatusBadRequest},
		{name: "empty field", method: http.MethodPost, body: `{"email":"","password":""}`, status: http.StatusUnauthorized},
		{name: "unknown field", method: http.MethodPost, body: `{"email":"admin@example.test","password":"secret","extra":true}`, status: http.StatusBadRequest},
		{name: "wrong method", method: http.MethodGet, body: `{}`, status: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/api/v1/auth/login", bytes.NewBufferString(tt.body))
			response := httptest.NewRecorder()

			handler.Login(response, request)

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestHandlerInternalErrorDoesNotLeak(t *testing.T) {
	service := NewHandler(Service{
		users:        failingUserRepository{},
		sessions:     newAuthFakeSessionRepository(),
		hasher:       authFakeHasher{},
		accessTokens: testAccessTokenService("secret", timeNow),
		refresh:      NewRefreshTokenGenerator(),
		refreshTTL:   24 * time.Hour,
		now:          timeNow,
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.test","password":"secret"}`))
	response := httptest.NewRecorder()

	service.Login(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "database") {
		t.Fatal("response leaked internal error")
	}
}

func TestHandlerMeReturnsProfile(t *testing.T) {
	users := newAuthFakeUserRepository()
	admin := activeAdmin()
	users.users = append(users.users, admin)
	handler := NewHandler(newTestService(users, newAuthFakeSessionRepository()))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request = request.WithContext(contextWithClaims(request.Context(), Claims{RegisteredClaims: jwtRegisteredClaims(admin.ID)}))
	response := httptest.NewRecorder()

	handler.Me(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if strings.Contains(response.Body.String(), "password_hash") {
		t.Fatal("response contains password_hash")
	}
}

func jwtRegisteredClaims(subject string) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{Subject: subject}
}

func timeNow() time.Time {
	return time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
}

type failingUserRepository struct{}

func (failingUserRepository) Create(ctx context.Context, u user.User) error {
	return errors.New("database exploded")
}
func (failingUserRepository) FindByID(ctx context.Context, id string) (user.User, error) {
	return user.User{}, errors.New("database exploded")
}
func (failingUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	return user.User{}, errors.New("database exploded")
}
func (failingUserRepository) FindByEmployeeNumber(ctx context.Context, employeeNumber string) (user.User, error) {
	return user.User{}, errors.New("database exploded")
}
func (failingUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, errors.New("database exploded")
}
func (failingUserRepository) ExistsByEmployeeNumber(ctx context.Context, employeeNumber string) (bool, error) {
	return false, errors.New("database exploded")
}
