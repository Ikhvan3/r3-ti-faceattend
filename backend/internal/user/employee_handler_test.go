package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmployeeHandlerListSuccess(t *testing.T) {
	repo := newEmployeeFakeRepository()
	repo.users = append(repo.users, employeeUser("00000000-0000-4000-8000-000000000001", "EMP-001", "one.ti@example.test"))
	handler := newEmployeeTestHandler(repo)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/employees?page=1&page_size=10", nil)
	response := httptest.NewRecorder()

	handler.Collection(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"items"`) {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

func TestEmployeeHandlerCreateReturns201AndHidesPasswordHash(t *testing.T) {
	handler := newEmployeeTestHandler(newEmployeeFakeRepository())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/employees", strings.NewReader(validEmployeeCreateJSON()))
	response := httptest.NewRecorder()

	handler.Collection(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if strings.Contains(response.Body.String(), "password_hash") {
		t.Fatal("response contains password_hash")
	}
}

func TestEmployeeHandlerRejectsMalformedJSONUnknownFieldAndWrongMethod(t *testing.T) {
	handler := newEmployeeTestHandler(newEmployeeFakeRepository())
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "malformed", method: http.MethodPost, path: "/api/v1/admin/employees", body: `{`, status: http.StatusBadRequest},
		{name: "unknown field", method: http.MethodPost, path: "/api/v1/admin/employees", body: `{"employee_number":"EMP-001","name":"Pegawai Dummy","email":"dummy.ti@example.test","initial_password":"password-dummy","role":"ADMIN"}`, status: http.StatusBadRequest},
		{name: "wrong collection method", method: http.MethodDelete, path: "/api/v1/admin/employees", body: `{}`, status: http.StatusMethodNotAllowed},
		{name: "wrong resource method", method: http.MethodPost, path: "/api/v1/admin/employees/00000000-0000-4000-8000-000000000001", body: `{}`, status: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			if tt.path == "/api/v1/admin/employees" {
				handler.Collection(response, request)
			} else {
				handler.Resource(response, request)
			}

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
		})
	}
}

func TestEmployeeHandlerInvalidUUIDNotFoundAndConflict(t *testing.T) {
	repo := newEmployeeFakeRepository()
	existing := employeeUser("00000000-0000-4000-8000-000000000001", "EMP-001", "dummy.ti@example.test")
	repo.users = append(repo.users, existing)
	handler := newEmployeeTestHandler(repo)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "invalid uuid", method: http.MethodGet, path: "/api/v1/admin/employees/not-a-uuid", status: http.StatusBadRequest},
		{name: "not found", method: http.MethodGet, path: "/api/v1/admin/employees/00000000-0000-4000-8000-000000000099", status: http.StatusNotFound},
		{name: "conflict", method: http.MethodPost, path: "/api/v1/admin/employees", body: validEmployeeCreateJSON(), status: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			if tt.path == "/api/v1/admin/employees" {
				handler.Collection(response, request)
			} else {
				handler.Resource(response, request)
			}

			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.status, response.Body.String())
			}
		})
	}
}

func newEmployeeTestHandler(repo *employeeFakeRepository) EmployeeHandler {
	return NewEmployeeHandler(NewEmployeeService(repo, employeeFakeHasher{}))
}

func validEmployeeCreateJSON() string {
	return `{"employee_number":"EMP-001","name":"Pegawai Dummy","email":"dummy.ti@example.test","initial_password":"password-dummy","phone":null,"position":"Staf TI"}`
}
