package location

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestLocationHandlerAuthAndMethods(t *testing.T) {
	handler := protectedLocationHandler(newFakeHTTPService(), auth.Claims{Role: user.RoleAdmin})

	missing := httptest.NewRequest(http.MethodGet, "/api/v1/admin/office-locations", nil)
	missingResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingResponse, missing)
	if missingResponse.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", missingResponse.Code)
	}

	userHandler := protectedLocationHandler(newFakeHTTPService(), auth.Claims{Role: user.RoleUser})
	userRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/office-locations", nil)
	userRequest.Header.Set("Authorization", "Bearer valid-token")
	userResponse := httptest.NewRecorder()
	userHandler.ServeHTTP(userResponse, userRequest)
	if userResponse.Code != http.StatusForbidden {
		t.Fatalf("user status = %d", userResponse.Code)
	}

	method := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/office-locations", nil)
	method.Header.Set("Authorization", "Bearer valid-token")
	methodResponse := httptest.NewRecorder()
	handler.ServeHTTP(methodResponse, method)
	if methodResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d", methodResponse.Code)
	}
}

func TestLocationHandlerSuccessAndBadRequests(t *testing.T) {
	handler := protectedLocationHandler(newFakeHTTPService(), auth.Claims{Role: user.RoleAdmin})
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{name: "list office", method: http.MethodGet, path: "/api/v1/admin/office-locations", want: http.StatusOK},
		{name: "create office", method: http.MethodPost, path: "/api/v1/admin/office-locations", body: `{"name":"Kantor","address":null,"latitude":-6.1,"longitude":106.8,"radius_meters":100}`, want: http.StatusCreated},
		{name: "detail office", method: http.MethodGet, path: "/api/v1/admin/office-locations/" + testOfficeID, want: http.StatusOK},
		{name: "update office", method: http.MethodPut, path: "/api/v1/admin/office-locations/" + testOfficeID, body: `{"name":"Kantor","address":"Alamat","latitude":-6.1,"longitude":106.8,"radius_meters":100}`, want: http.StatusOK},
		{name: "status office", method: http.MethodPatch, path: "/api/v1/admin/office-locations/" + testOfficeID + "/status", body: `{"is_active":false}`, want: http.StatusOK},
		{name: "list assignment", method: http.MethodGet, path: "/api/v1/admin/location-assignments", want: http.StatusOK},
		{name: "create assignment", method: http.MethodPost, path: "/api/v1/admin/location-assignments", body: `{"user_id":"` + testUserID + `","office_location_id":"` + testOfficeID + `","effective_from":"2026-08-06","effective_to":null}`, want: http.StatusCreated},
		{name: "detail assignment", method: http.MethodGet, path: "/api/v1/admin/location-assignments/" + testAssignmentID, want: http.StatusOK},
		{name: "end assignment", method: http.MethodPatch, path: "/api/v1/admin/location-assignments/" + testAssignmentID + "/end", body: `{"effective_to":"2026-08-31"}`, want: http.StatusOK},
		{name: "malformed json", method: http.MethodPost, path: "/api/v1/admin/office-locations", body: `{`, want: http.StatusBadRequest},
		{name: "unknown field", method: http.MethodPost, path: "/api/v1/admin/office-locations", body: `{"name":"Kantor","latitude":-6.1,"longitude":106.8,"radius_meters":100,"is_active":true}`, want: http.StatusBadRequest},
		{name: "invalid uuid", method: http.MethodGet, path: "/api/v1/admin/office-locations/not-a-uuid", want: http.StatusBadRequest},
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

func TestLocationHandlerErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*fakeHTTPService)
		path  string
		body  string
		want  int
	}{
		{name: "not found", setup: func(s *fakeHTTPService) { s.officeDetailErr = ErrOfficeNotFound }, path: "/api/v1/admin/office-locations/" + testOfficeID, want: http.StatusNotFound},
		{name: "deactivate conflict", setup: func(s *fakeHTTPService) { s.officeStatusErr = ErrOfficeInUse }, path: "/api/v1/admin/office-locations/" + testOfficeID + "/status", body: `{"is_active":false}`, want: http.StatusConflict},
		{name: "overlap", setup: func(s *fakeHTTPService) { s.assignmentCreateErr = ErrAssignmentOverlap }, path: "/api/v1/admin/location-assignments", body: `{"user_id":"` + testUserID + `","office_location_id":"` + testOfficeID + `","effective_from":"2026-08-06","effective_to":null}`, want: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeHTTPService()
			tt.setup(service)
			handler := protectedLocationHandler(service, auth.Claims{Role: user.RoleAdmin})
			method := http.MethodGet
			if tt.body != "" {
				method = http.MethodPost
			}
			if strings.HasSuffix(tt.path, "/status") {
				method = http.MethodPatch
			}
			request := httptest.NewRequest(method, tt.path, strings.NewReader(tt.body))
			request.Header.Set("Authorization", "Bearer valid-token")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestLocationRequirementAuth(t *testing.T) {
	handler := protectedUserLocationRequirementHandler(newFakeHTTPService(), auth.Claims{Role: user.RoleAdmin})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/location-requirement", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("admin user endpoint status = %d", response.Code)
	}

	handler = protectedUserLocationRequirementHandler(newFakeHTTPService(), auth.Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: testUserID}, Role: user.RoleUser})
	request = httptest.NewRequest(http.MethodGet, "/api/v1/attendance/location-requirement", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("user endpoint status = %d, body=%s", response.Code, response.Body.String())
	}
}

func protectedLocationHandler(service *fakeHTTPService, claims auth.Claims) http.Handler {
	handler := NewHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/admin/office-locations", handler.OfficeCollection)
	mux.HandleFunc("/api/v1/admin/office-locations/", handler.OfficeResource)
	mux.HandleFunc("/api/v1/admin/location-assignments", handler.AssignmentCollection)
	mux.HandleFunc("/api/v1/admin/location-assignments/", handler.AssignmentResource)
	return auth.Authenticate(fakeVerifier{claims: claims}, auth.RequireRole(user.RoleAdmin, mux))
}

func protectedUserLocationRequirementHandler(service *fakeHTTPService, claims auth.Claims) http.Handler {
	handler := NewHandler(service)
	return auth.Authenticate(fakeVerifier{claims: claims}, auth.RequireRole(user.RoleUser, http.HandlerFunc(handler.LocationRequirement)))
}

type fakeVerifier struct {
	claims auth.Claims
}

func (v fakeVerifier) VerifyAccessToken(token string) (auth.Claims, error) {
	if token == "" || token == "expired-token" {
		return auth.Claims{}, auth.ErrInvalidToken
	}
	if v.claims.Subject == "" {
		v.claims.RegisteredClaims = jwt.RegisteredClaims{Subject: testAdminID}
	}
	return v.claims, nil
}

type fakeHTTPService struct {
	officeDetailErr     error
	officeStatusErr     error
	assignmentCreateErr error
}

func newFakeHTTPService() *fakeHTTPService {
	return &fakeHTTPService{}
}

func (s *fakeHTTPService) ListOfficeLocations(_ context.Context, _ OfficeLocationListFilter) (OfficeLocationList, error) {
	return OfficeLocationList{Items: []OfficeLocation{}, Page: 1, PageSize: 10}, nil
}

func (s *fakeHTTPService) CreateOfficeLocation(_ context.Context, _ OfficeLocationInput) (OfficeLocation, error) {
	return testOffice(), nil
}

func (s *fakeHTTPService) OfficeLocationDetail(_ context.Context, id string) (OfficeLocation, error) {
	if !validUUID(id) {
		return OfficeLocation{}, ErrInvalidInput
	}
	if s.officeDetailErr != nil {
		return OfficeLocation{}, s.officeDetailErr
	}
	return testOffice(), nil
}

func (s *fakeHTTPService) UpdateOfficeLocation(_ context.Context, _ string, _ OfficeLocationInput) (OfficeLocation, error) {
	return testOffice(), nil
}

func (s *fakeHTTPService) UpdateOfficeLocationStatus(_ context.Context, _ string, _ bool) (OfficeLocation, error) {
	if s.officeStatusErr != nil {
		return OfficeLocation{}, s.officeStatusErr
	}
	office := testOffice()
	office.IsActive = false
	return office, nil
}

func (s *fakeHTTPService) ListLocationAssignments(_ context.Context, _ LocationAssignmentListFilter) (LocationAssignmentList, error) {
	return LocationAssignmentList{Items: []LocationAssignment{}, Page: 1, PageSize: 10}, nil
}

func (s *fakeHTTPService) CreateLocationAssignment(_ context.Context, _ LocationAssignmentInput) (LocationAssignment, error) {
	if s.assignmentCreateErr != nil {
		return LocationAssignment{}, s.assignmentCreateErr
	}
	return testLocationAssignment(), nil
}

func (s *fakeHTTPService) LocationAssignmentDetail(_ context.Context, _ string) (LocationAssignment, error) {
	return testLocationAssignment(), nil
}

func (s *fakeHTTPService) EndLocationAssignment(_ context.Context, _ string, _ string) (LocationAssignment, error) {
	return testLocationAssignment(), nil
}

func (s *fakeHTTPService) LocationRequirement(_ context.Context, _ auth.Claims) (LocationRequirement, error) {
	assignment := testLocationAssignment()
	return LocationRequirement{Assignment: assignment, Office: assignment.Office}, nil
}

func testOffice() OfficeLocation {
	return OfficeLocation{ID: testOfficeID, Name: "Kantor Regional 3", Latitude: -6.1, Longitude: 106.8, RadiusMeters: 100, IsActive: true}
}

func testLocationAssignment() LocationAssignment {
	return LocationAssignment{ID: testAssignmentID, User: user.EmployeeProfile{ID: testUserID, EmployeeNumber: "EMP-DUMMY-001", Name: "Pegawai Dummy", Email: "pegawai.dummy@example.test", Role: user.RoleUser, AccountStatus: user.AccountStatusActive}, Office: testOffice(), EffectiveFrom: "2026-08-06", Status: AssignmentStatusCurrent}
}
