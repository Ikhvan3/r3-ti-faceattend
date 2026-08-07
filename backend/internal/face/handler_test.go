package face

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestHandlerAuth(t *testing.T) {
	handler := protectedHandler(newFakeHTTPService(), userClaims(testUserID))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/face/status", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", response.Code)
	}

	adminHandler := protectedHandler(newFakeHTTPService(), adminClaims())
	request = httptest.NewRequest(http.MethodGet, "/api/v1/face/status", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	adminHandler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("admin status = %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/face/verify", strings.NewReader(validVerifyJSON()))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("verify missing token status = %d", response.Code)
	}
}

func TestHandlerStatusDoesNotExposeEmbedding(t *testing.T) {
	service := newFakeHTTPService()
	enrolledAt := time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC)
	service.status = StatusResponse{Enrolled: true, FaceStatus: FaceStatusEnrolled, EmbeddingModel: "test-face-model", EmbeddingVersion: "v1", EnrolledAt: &enrolledAt}
	handler := protectedHandler(service, userClaims(testUserID))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/face/status", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, `"embedding":`) || strings.Contains(body, "0.1") {
		t.Fatalf("status response exposes embedding: %s", body)
	}
}

func TestHandlerEnrollValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "malformed", body: `{"embedding":`, want: http.StatusBadRequest},
		{name: "unknown field", body: `{"embedding":[0.1,0.2,0.3],"embedding_model":"test-face-model","embedding_version":"v1","user_id":"other"}`, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := protectedHandler(newFakeHTTPService(), userClaims(testUserID))
			request := httptest.NewRequest(http.MethodPost, "/api/v1/face/enroll", strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d want %d body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestHandlerEnrollSuccessAndDuplicate(t *testing.T) {
	service := newFakeHTTPService()
	handler := protectedHandler(service, userClaims(testUserID))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/face/enroll", strings.NewReader(validEnrollJSON()))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if len(service.lastInput.Embedding) != 3 {
		t.Fatalf("embedding length passed to service = %d", len(service.lastInput.Embedding))
	}
	if strings.Contains(response.Body.String(), `"embedding":`) || strings.Contains(response.Body.String(), "0.1") {
		t.Fatalf("enroll response exposes embedding: %s", response.Body.String())
	}

	service.enrollErr = ErrAlreadyEnrolled
	request = httptest.NewRequest(http.MethodPost, "/api/v1/face/enroll", strings.NewReader(validEnrollJSON()))
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d", response.Code)
	}
}

func TestHandlerVerifyValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "malformed", body: `{"embedding":`, want: http.StatusBadRequest},
		{name: "unknown field", body: `{"embedding":[0.1,0.2,0.3],"embedding_model":"test-face-model","embedding_version":"v1","threshold":0.9}`, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := protectedHandler(newFakeHTTPService(), userClaims(testUserID))
			request := httptest.NewRequest(http.MethodPost, "/api/v1/face/verify", strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d want %d body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestHandlerVerifyResponses(t *testing.T) {
	service := newFakeHTTPService()
	handler := protectedHandler(service, userClaims(testUserID))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/face/verify", strings.NewReader(validVerifyJSON()))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
	if len(service.lastVerificationInput.Embedding) != 3 {
		t.Fatalf("embedding length passed to service = %d", len(service.lastVerificationInput.Embedding))
	}
	body := response.Body.String()
	if !strings.Contains(body, `"verified":true`) {
		t.Fatalf("verify response = %s, want verified true", body)
	}
	if strings.Contains(body, `"embedding":`) || strings.Contains(body, `"threshold"`) || strings.Contains(body, "0.1") {
		t.Fatalf("verify response exposes private data: %s", body)
	}

	service.verifyResult = VerificationResponse{Verified: false}
	request = httptest.NewRequest(http.MethodPost, "/api/v1/face/verify", strings.NewReader(validVerifyJSON()))
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("mismatch status = %d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"verified":false`) {
		t.Fatalf("verify response = %s, want verified false", response.Body.String())
	}
}

func TestHandlerVerifyErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not enrolled conflict", err: ErrNotEnrolled, want: http.StatusConflict},
		{name: "invalid input", err: ErrInvalidInput, want: http.StatusBadRequest},
		{name: "wrong model", err: ErrUnsupportedModel, want: http.StatusBadRequest},
		{name: "wrong dimension", err: ErrInvalidDimension, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeHTTPService()
			service.verifyErr = tt.err
			handler := protectedHandler(service, userClaims(testUserID))
			request := httptest.NewRequest(http.MethodPost, "/api/v1/face/verify", strings.NewReader(validVerifyJSON()))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d want %d body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestHandlerReset(t *testing.T) {
	service := newFakeHTTPService()
	handler := protectedHandler(service, userClaims(testUserID))

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/face/enrollment", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}

	service.resetErr = ErrProfileNotFound
	request = httptest.NewRequest(http.MethodDelete, "/api/v1/face/enrollment", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", response.Code)
	}
}

func protectedHandler(service *fakeHTTPService, claims auth.Claims) http.Handler {
	handler := NewHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/face/status", handler.Status)
	mux.HandleFunc("/api/v1/face/enroll", handler.Enroll)
	mux.HandleFunc("/api/v1/face/verify", handler.Verify)
	mux.HandleFunc("/api/v1/face/verify-for-attendance", handler.VerifyForAttendance)
	mux.HandleFunc("/api/v1/face/enrollment", handler.Enrollment)
	return auth.Authenticate(fakeVerifier{claims: claims}, auth.RequireRole(user.RoleUser, mux))
}

type fakeVerifier struct {
	claims auth.Claims
}

func (v fakeVerifier) VerifyAccessToken(token string) (auth.Claims, error) {
	if token == "" || token == "expired-token" {
		return auth.Claims{}, auth.ErrInvalidToken
	}
	return v.claims, nil
}

type fakeHTTPService struct {
	status                StatusResponse
	statusErr             error
	enrollErr             error
	verifyResult          VerificationResponse
	verifyErr             error
	resetErr              error
	lastInput             EnrollmentInput
	lastVerificationInput VerificationInput
	lastAttendanceInput   AttendanceVerificationInput
}

func newFakeHTTPService() *fakeHTTPService {
	return &fakeHTTPService{
		status:       StatusResponse{Enrolled: false, FaceStatus: FaceStatusNotEnrolled},
		verifyResult: VerificationResponse{Verified: true},
	}
}

func (s *fakeHTTPService) Status(_ context.Context, _ auth.Claims) (StatusResponse, error) {
	if s.statusErr != nil {
		return StatusResponse{}, s.statusErr
	}
	return s.status, nil
}

func (s *fakeHTTPService) Enroll(_ context.Context, _ auth.Claims, input EnrollmentInput) (StatusResponse, error) {
	s.lastInput = input
	if s.enrollErr != nil {
		return StatusResponse{}, s.enrollErr
	}
	now := time.Date(2026, 8, 7, 1, 0, 0, 0, time.UTC)
	return StatusResponse{Enrolled: true, FaceStatus: FaceStatusEnrolled, EmbeddingModel: input.EmbeddingModel, EmbeddingVersion: input.EmbeddingVersion, EnrolledAt: &now}, nil
}

func (s *fakeHTTPService) Verify(_ context.Context, _ auth.Claims, input VerificationInput) (VerificationResponse, error) {
	s.lastVerificationInput = input
	if s.verifyErr != nil {
		return VerificationResponse{}, s.verifyErr
	}
	return s.verifyResult, nil
}

func (s *fakeHTTPService) VerifyForAttendance(_ context.Context, _ auth.Claims, input AttendanceVerificationInput) (AttendanceVerificationResponse, error) {
	s.lastAttendanceInput = input
	if s.verifyErr != nil {
		return AttendanceVerificationResponse{}, s.verifyErr
	}
	return AttendanceVerificationResponse{
		VerificationGrant: "valid-grant",
		ExpiresAt:         time.Date(2026, 8, 7, 1, 2, 0, 0, time.UTC),
	}, nil
}

func (s *fakeHTTPService) Reset(_ context.Context, _ auth.Claims) error {
	if s.resetErr != nil {
		return s.resetErr
	}
	return nil
}

func validEnrollJSON() string {
	body := map[string]any{
		"embedding":         []float64{0.1, 0.2, 0.3},
		"embedding_model":   "test-face-model",
		"embedding_version": "v1",
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func validVerifyJSON() string {
	body := map[string]any{
		"embedding":         []float64{0.1, 0.2, 0.3},
		"embedding_model":   "test-face-model",
		"embedding_version": "v1",
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

var _ = errors.Is
