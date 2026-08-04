package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"r3-ti-faceattend/backend/internal/config"
	"r3-ti-faceattend/backend/internal/health"
)

type fakeDatabase struct{}

func (fakeDatabase) Ping(_ context.Context) error {
	return nil
}

func TestHealthRoutesUseConsistentResponse(t *testing.T) {
	handler := newHTTPHandler(config.Config{AppEnv: "test"}, fakeDatabase{})

	paths := []string{"/health", "/api/v1/health"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %q", response.Header().Get("Content-Type"))
			}

			var body health.Response
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if body.Status != "ok" {
				t.Fatalf("expected overall status ok, got %q", body.Status)
			}
			if body.Checks["api"].Status != "ok" {
				t.Fatalf("expected api status ok, got %q", body.Checks["api"].Status)
			}
			if body.Checks["database"].Status != "ok" {
				t.Fatalf("expected database status ok, got %q", body.Checks["database"].Status)
			}
		})
	}
}

func TestHealthRouteRejectsNonGetMethod(t *testing.T) {
	handler := newHTTPHandler(config.Config{AppEnv: "test"}, fakeDatabase{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, response.Header().Get("Allow"))
	}
}

func TestWaitForServerReturnsListenError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("listen failed")
	errCh := make(chan error, 1)
	errCh <- expectedErr

	err := waitForServer(ctx, &http.Server{}, errCh)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected listen error %v, got %v", expectedErr, err)
	}
}

func TestWaitForServerIgnoresServerClosed(t *testing.T) {
	ctx := context.Background()
	errCh := make(chan error, 1)
	errCh <- http.ErrServerClosed

	err := waitForServer(ctx, &http.Server{}, errCh)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWaitForServerShutsDownWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errCh := make(chan error)

	err := waitForServer(ctx, &http.Server{}, errCh)

	if err != nil {
		t.Fatalf("expected nil shutdown error, got %v", err)
	}
}
