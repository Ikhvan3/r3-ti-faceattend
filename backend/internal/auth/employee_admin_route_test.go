package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"r3-ti-faceattend/backend/internal/user"
)

func TestEmployeeAdminRouteRejectsMissingTokenAndUserRole(t *testing.T) {
	handler := newProtectedEmployeeHandler(employeeRouteVerifier{claims: Claims{Role: user.RoleUser}})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/employees", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/employees", nil)
	request.Header.Set("Authorization", "Bearer user-token")
	response = httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("user role status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func newProtectedEmployeeHandler(verifier employeeRouteVerifier) http.Handler {
	repo := employeeRouteRepository{}
	hasher := employeeRouteHasher{}
	handler := user.NewEmployeeHandler(user.NewEmployeeService(repo, hasher))
	mux := http.NewServeMux()
	mux.Handle("/api/v1/admin/employees", Authenticate(verifier, RequireRole(user.RoleAdmin, http.HandlerFunc(handler.Collection))))
	mux.Handle("/api/v1/admin/employees/", Authenticate(verifier, RequireRole(user.RoleAdmin, http.HandlerFunc(handler.Resource))))
	return mux
}

type employeeRouteVerifier struct {
	claims Claims
	err    error
}

func (v employeeRouteVerifier) VerifyAccessToken(token string) (Claims, error) {
	if v.err != nil {
		return Claims{}, v.err
	}
	return v.claims, nil
}

type employeeRouteHasher struct{}

func (employeeRouteHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}
	return "hashed:" + password, nil
}

func (employeeRouteHasher) Compare(hash string, password string) bool {
	return hash == "hashed:"+password
}

type employeeRouteRepository struct{}

func (employeeRouteRepository) ListEmployees(ctx context.Context, filter user.EmployeeListFilter) ([]user.User, error) {
	return []user.User{}, nil
}

func (employeeRouteRepository) CountEmployees(ctx context.Context, filter user.EmployeeListFilter) (int, error) {
	return 0, nil
}

func (employeeRouteRepository) FindEmployeeByID(ctx context.Context, id string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func (employeeRouteRepository) CreateEmployee(ctx context.Context, u user.User) (user.User, error) {
	return u, nil
}

func (employeeRouteRepository) UpdateEmployee(ctx context.Context, u user.User) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func (employeeRouteRepository) UpdateEmployeeStatus(ctx context.Context, id string, status user.AccountStatus) (user.User, error) {
	return user.User{}, user.ErrNotFound
}
