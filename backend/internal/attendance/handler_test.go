package attendance

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestHandlerRejectsMissingTokenAndAdminRole(t *testing.T) {
	handler := protectedHandler(newFakeAttendanceHTTPService(), auth.Claims{Role: user.RoleUser})

	missingTokenRequest := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/today", nil)
	missingTokenResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingTokenResponse, missingTokenRequest)
	if missingTokenResponse.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", missingTokenResponse.Code, http.StatusUnauthorized)
	}

	adminHandler := protectedHandler(newFakeAttendanceHTTPService(), auth.Claims{Role: user.RoleAdmin})
	adminRequest := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/today", nil)
	adminRequest.Header.Set("Authorization", "Bearer valid-token")
	adminResponse := httptest.NewRecorder()
	adminHandler.ServeHTTP(adminResponse, adminRequest)
	if adminResponse.Code != http.StatusForbidden {
		t.Fatalf("admin status = %d, want %d", adminResponse.Code, http.StatusForbidden)
	}
}

func TestHandlerTodayCheckInAndCheckOut(t *testing.T) {
	service := newFakeAttendanceHTTPService()
	handler := protectedHandler(service, userHTTPClaims())

	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{name: "today", method: http.MethodGet, path: "/api/v1/attendance/today", want: http.StatusOK},
		{name: "check in", method: http.MethodPost, path: "/api/v1/attendance/check-in", want: http.StatusCreated},
		{name: "check out", method: http.MethodPost, path: "/api/v1/attendance/check-out", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", response.Code, tt.want, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestHandlerConflictAndInvalidHistoryQuery(t *testing.T) {
	service := newFakeAttendanceHTTPService()
	service.checkInErr = ErrAlreadyCheckedIn
	handler := protectedHandler(service, userHTTPClaims())

	checkInRequest := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/check-in", nil)
	checkInRequest.Header.Set("Authorization", "Bearer valid-token")
	checkInResponse := httptest.NewRecorder()
	handler.ServeHTTP(checkInResponse, checkInRequest)
	if checkInResponse.Code != http.StatusConflict {
		t.Fatalf("check-in status = %d, want %d", checkInResponse.Code, http.StatusConflict)
	}

	historyRequest := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/history?page=abc", nil)
	historyRequest.Header.Set("Authorization", "Bearer valid-token")
	historyResponse := httptest.NewRecorder()
	handler.ServeHTTP(historyResponse, historyRequest)
	if historyResponse.Code != http.StatusBadRequest {
		t.Fatalf("history status = %d, want %d", historyResponse.Code, http.StatusBadRequest)
	}
}

func TestHandlerHistoryDoesNotExposeOtherUserData(t *testing.T) {
	service := newFakeAttendanceHTTPService()
	service.history = HistoryList{
		Items: []HistoryItem{{
			ID:             "record-id",
			AttendanceDate: "2026-08-05",
			Schedule:       testSchedule(),
			CheckInAt:      time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC),
			State:          StateCheckedIn,
		}},
		Page:       1,
		PageSize:   10,
		TotalItems: 1,
		TotalPages: 1,
	}
	handler := protectedHandler(service, userHTTPClaims())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/history", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if strings.Contains(body, "other-user") || strings.Contains(body, "user_id") {
		t.Fatalf("history response exposes user data: %s", body)
	}
}

func TestHandlerRejectsNonEmptyCheckInBody(t *testing.T) {
	handler := protectedHandler(newFakeAttendanceHTTPService(), userHTTPClaims())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/check-in", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func protectedHandler(service *fakeAttendanceHTTPService, claims auth.Claims) http.Handler {
	handler := NewHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/attendance/today", handler.Today)
	mux.HandleFunc("/api/v1/attendance/check-in", handler.CheckIn)
	mux.HandleFunc("/api/v1/attendance/check-out", handler.CheckOut)
	mux.HandleFunc("/api/v1/attendance/history", handler.History)
	return auth.Authenticate(fakeAttendanceVerifier{claims: claims}, auth.RequireRole(user.RoleUser, mux))
}

type fakeAttendanceVerifier struct {
	claims auth.Claims
}

func (v fakeAttendanceVerifier) VerifyAccessToken(token string) (auth.Claims, error) {
	if token == "" || token == "expired-token" {
		return auth.Claims{}, auth.ErrInvalidToken
	}
	return v.claims, nil
}

type fakeAttendanceHTTPService struct {
	todayErr    error
	checkInErr  error
	checkOutErr error
	historyErr  error
	history     HistoryList
}

func newFakeAttendanceHTTPService() *fakeAttendanceHTTPService {
	return &fakeAttendanceHTTPService{
		history: HistoryList{
			Items:      []HistoryItem{},
			Page:       1,
			PageSize:   10,
			TotalItems: 0,
			TotalPages: 0,
		},
	}
}

func (s *fakeAttendanceHTTPService) Today(_ context.Context, _ auth.Claims) (DailyStatus, error) {
	if s.todayErr != nil {
		return DailyStatus{}, s.todayErr
	}
	return DailyStatus{AttendanceDate: "2026-08-05", Schedule: testSchedule(), State: StateNotCheckedIn, CanCheckIn: true}, nil
}

func (s *fakeAttendanceHTTPService) CheckIn(_ context.Context, _ auth.Claims) (DailyStatus, error) {
	if s.checkInErr != nil {
		return DailyStatus{}, s.checkInErr
	}
	return DailyStatus{AttendanceDate: "2026-08-05", Schedule: testSchedule(), State: StateCheckedIn, CanCheckOut: true}, nil
}

func (s *fakeAttendanceHTTPService) CheckOut(_ context.Context, _ auth.Claims) (DailyStatus, error) {
	if s.checkOutErr != nil {
		return DailyStatus{}, s.checkOutErr
	}
	checkOutAt := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	return DailyStatus{AttendanceDate: "2026-08-05", Schedule: testSchedule(), CheckOutAt: &checkOutAt, State: StateCompleted}, nil
}

func (s *fakeAttendanceHTTPService) History(_ context.Context, _ auth.Claims, _ HistoryFilter) (HistoryList, error) {
	if s.historyErr != nil {
		return HistoryList{}, s.historyErr
	}
	return s.history, nil
}

func testSchedule() WorkSchedule {
	return WorkSchedule{ID: "schedule-id", Name: "Jadwal Kerja Dummy TI", StartTime: "08:00", EndTime: "17:00", GraceMinutes: 15, IsActive: true}
}

func userHTTPClaims() auth.Claims {
	return auth.Claims{
		Role: user.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "00000000-0000-4000-8000-000000000001",
		},
	}
}

func decodeAttendanceResponse(t *testing.T, body string) attendanceResponse {
	t.Helper()
	var response attendanceResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

var _ = errors.Is
