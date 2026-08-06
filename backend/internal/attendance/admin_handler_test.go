package attendance

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestAdminScheduleHandlerAuthAndMethods(t *testing.T) {
	handler := protectedAdminScheduleHandler(newFakeAdminScheduleHTTPService(), auth.Claims{Role: user.RoleAdmin})

	missingRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/work-schedules", nil)
	missingResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingResponse, missingRequest)
	if missingResponse.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", missingResponse.Code, http.StatusUnauthorized)
	}

	userHandler := protectedAdminScheduleHandler(newFakeAdminScheduleHTTPService(), auth.Claims{Role: user.RoleUser})
	userRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/work-schedules", nil)
	userRequest.Header.Set("Authorization", "Bearer valid-token")
	userResponse := httptest.NewRecorder()
	userHandler.ServeHTTP(userResponse, userRequest)
	if userResponse.Code != http.StatusForbidden {
		t.Fatalf("user status = %d, want %d", userResponse.Code, http.StatusForbidden)
	}

	methodRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/work-schedules", nil)
	methodRequest.Header.Set("Authorization", "Bearer valid-token")
	methodResponse := httptest.NewRecorder()
	handler.ServeHTTP(methodResponse, methodRequest)
	if methodResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d, want %d", methodResponse.Code, http.StatusMethodNotAllowed)
	}
}

func TestAdminScheduleHandlerSuccessAndBadRequests(t *testing.T) {
	handler := protectedAdminScheduleHandler(newFakeAdminScheduleHTTPService(), auth.Claims{Role: user.RoleAdmin})

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{name: "list schedule", method: http.MethodGet, path: "/api/v1/admin/work-schedules", want: http.StatusOK},
		{name: "create schedule", method: http.MethodPost, path: "/api/v1/admin/work-schedules", body: `{"name":"Jadwal","start_time":"08:00","end_time":"17:00","grace_minutes":15}`, want: http.StatusCreated},
		{name: "detail schedule", method: http.MethodGet, path: "/api/v1/admin/work-schedules/" + testScheduleID, want: http.StatusOK},
		{name: "status schedule", method: http.MethodPatch, path: "/api/v1/admin/work-schedules/" + testScheduleID + "/status", body: `{"is_active":false}`, want: http.StatusOK},
		{name: "list assignment", method: http.MethodGet, path: "/api/v1/admin/schedule-assignments", want: http.StatusOK},
		{name: "create assignment", method: http.MethodPost, path: "/api/v1/admin/schedule-assignments", body: `{"user_id":"` + testUserID + `","schedule_id":"` + testScheduleID + `","effective_from":"2026-08-05","effective_to":null}`, want: http.StatusCreated},
		{name: "detail assignment", method: http.MethodGet, path: "/api/v1/admin/schedule-assignments/" + testAssignmentID, want: http.StatusOK},
		{name: "end assignment", method: http.MethodPatch, path: "/api/v1/admin/schedule-assignments/" + testAssignmentID + "/end", body: `{"effective_to":"2026-08-31"}`, want: http.StatusOK},
		{name: "malformed json", method: http.MethodPost, path: "/api/v1/admin/work-schedules", body: `{`, want: http.StatusBadRequest},
		{name: "unknown field", method: http.MethodPost, path: "/api/v1/admin/work-schedules", body: `{"name":"Jadwal","start_time":"08:00","end_time":"17:00","grace_minutes":15,"is_active":true}`, want: http.StatusBadRequest},
		{name: "invalid uuid", method: http.MethodGet, path: "/api/v1/admin/work-schedules/not-a-uuid", want: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", response.Code, tt.want, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "password_hash") {
				t.Fatalf("response exposes password_hash: %s", response.Body.String())
			}
		})
	}
}

func TestAdminScheduleHandlerErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*fakeAdminScheduleHTTPService)
		path  string
		want  int
	}{
		{name: "not found", setup: func(s *fakeAdminScheduleHTTPService) { s.scheduleDetailErr = ErrScheduleNotFound }, path: "/api/v1/admin/work-schedules/" + testScheduleID, want: http.StatusNotFound},
		{name: "conflict", setup: func(s *fakeAdminScheduleHTTPService) { s.scheduleStatusErr = ErrScheduleInUse }, path: "/api/v1/admin/work-schedules/" + testScheduleID + "/status", want: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeAdminScheduleHTTPService()
			tt.setup(service)
			handler := protectedAdminScheduleHandler(service, auth.Claims{Role: user.RoleAdmin})
			method := http.MethodGet
			body := ""
			if strings.HasSuffix(tt.path, "/status") {
				method = http.MethodPatch
				body = `{"is_active":false}`
			}
			request := httptest.NewRequest(method, tt.path, strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func protectedAdminScheduleHandler(service *fakeAdminScheduleHTTPService, claims auth.Claims) http.Handler {
	handler := NewAdminScheduleHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/admin/work-schedules", handler.WorkScheduleCollection)
	mux.HandleFunc("/api/v1/admin/work-schedules/", handler.WorkScheduleResource)
	mux.HandleFunc("/api/v1/admin/schedule-assignments", handler.AssignmentCollection)
	mux.HandleFunc("/api/v1/admin/schedule-assignments/", handler.AssignmentResource)
	return auth.Authenticate(fakeAdminScheduleVerifier{claims: claims}, auth.RequireRole(user.RoleAdmin, mux))
}

type fakeAdminScheduleVerifier struct {
	claims auth.Claims
}

func (v fakeAdminScheduleVerifier) VerifyAccessToken(token string) (auth.Claims, error) {
	if token == "" || token == "expired-token" {
		return auth.Claims{}, auth.ErrInvalidToken
	}
	if v.claims.Subject == "" {
		v.claims.RegisteredClaims = jwt.RegisteredClaims{Subject: testAdminID}
	}
	return v.claims, nil
}

type fakeAdminScheduleHTTPService struct {
	scheduleDetailErr error
	scheduleStatusErr error
}

func newFakeAdminScheduleHTTPService() *fakeAdminScheduleHTTPService {
	return &fakeAdminScheduleHTTPService{}
}

func (s *fakeAdminScheduleHTTPService) ListWorkSchedules(_ context.Context, _ WorkScheduleListFilter) (WorkScheduleList, error) {
	return WorkScheduleList{Items: []WorkSchedule{}, Page: 1, PageSize: 10}, nil
}

func (s *fakeAdminScheduleHTTPService) CreateWorkSchedule(_ context.Context, _ WorkScheduleInput) (WorkSchedule, error) {
	return testAdminSchedule(), nil
}

func (s *fakeAdminScheduleHTTPService) WorkScheduleDetail(_ context.Context, id string) (WorkSchedule, error) {
	if !validAdminUUID(id) {
		return WorkSchedule{}, ErrInvalidInput
	}
	if s.scheduleDetailErr != nil {
		return WorkSchedule{}, s.scheduleDetailErr
	}
	return testAdminSchedule(), nil
}

func (s *fakeAdminScheduleHTTPService) UpdateWorkSchedule(_ context.Context, _ string, _ WorkScheduleInput) (WorkSchedule, error) {
	return testAdminSchedule(), nil
}

func (s *fakeAdminScheduleHTTPService) UpdateWorkScheduleStatus(_ context.Context, _ string, _ bool) (WorkSchedule, error) {
	if s.scheduleStatusErr != nil {
		return WorkSchedule{}, s.scheduleStatusErr
	}
	schedule := testAdminSchedule()
	schedule.IsActive = false
	return schedule, nil
}

func (s *fakeAdminScheduleHTTPService) ListAssignments(_ context.Context, _ AssignmentListFilter) (ScheduleAssignmentList, error) {
	return ScheduleAssignmentList{Items: []ScheduleAssignment{}, Page: 1, PageSize: 10}, nil
}

func (s *fakeAdminScheduleHTTPService) CreateAssignment(_ context.Context, _ AssignmentCreateInput) (ScheduleAssignment, error) {
	return testAdminAssignment(), nil
}

func (s *fakeAdminScheduleHTTPService) AssignmentDetail(_ context.Context, _ string) (ScheduleAssignment, error) {
	return testAdminAssignment(), nil
}

func (s *fakeAdminScheduleHTTPService) EndAssignment(_ context.Context, _ string, _ string) (ScheduleAssignment, error) {
	return testAdminAssignment(), nil
}

func testAdminSchedule() WorkSchedule {
	return WorkSchedule{ID: testScheduleID, Name: "Jadwal Reguler", StartTime: "08:00", EndTime: "17:00", GraceMinutes: 15, IsActive: true}
}

func testAdminAssignment() ScheduleAssignment {
	return ScheduleAssignment{
		ID:            testAssignmentID,
		User:          user.EmployeeProfile{ID: testUserID, EmployeeNumber: "EMP-DUMMY-001", Name: "Pegawai Dummy", Email: "pegawai.dummy@example.test", Role: user.RoleUser, AccountStatus: user.AccountStatusActive},
		Schedule:      testAdminSchedule(),
		EffectiveFrom: "2026-08-05",
	}
}

var _ = errors.Is
