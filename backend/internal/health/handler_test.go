package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeDatabase struct {
	err error
}

func (db fakeDatabase) Ping(_ context.Context) error {
	return db.err
}

func TestHandlerReportsHealthyDatabase(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler("test", fakeDatabase{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	body := decodeResponse(t, response)
	if body.Status != "ok" {
		t.Fatalf("expected overall status ok, got %q", body.Status)
	}
	if body.Checks["database"].Status != "ok" {
		t.Fatalf("expected database status ok, got %q", body.Checks["database"].Status)
	}
	assertJSONContentType(t, response)
}

func TestHandlerReportsUnhealthyDatabase(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler := NewHandler("test", fakeDatabase{err: errors.New("connection refused")})
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}

	body := decodeResponse(t, response)
	if body.Status != "degraded" {
		t.Fatalf("expected overall status degraded, got %q", body.Status)
	}
	if body.Checks["api"].Status != "ok" {
		t.Fatalf("expected api status ok, got %q", body.Checks["api"].Status)
	}
	if body.Checks["database"].Status != "error" {
		t.Fatalf("expected database status error, got %q", body.Checks["database"].Status)
	}
	assertJSONContentType(t, response)
}

func TestHandlerRejectsNonGetMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler("test", fakeDatabase{}).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
	if response.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, response.Header().Get("Allow"))
	}

	body := decodeResponse(t, response)
	if body.Status != "error" {
		t.Fatalf("expected overall status error, got %q", body.Status)
	}
	assertJSONContentType(t, response)
}

func TestHandlerUsesJSONContentType(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewHandler("test", fakeDatabase{}).ServeHTTP(response, request)

	assertJSONContentType(t, response)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder) Response {
	t.Helper()

	var body Response
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return body
}

func assertJSONContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", response.Header().Get("Content-Type"))
	}
}
