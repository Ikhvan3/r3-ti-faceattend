package location

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

const (
	testUserID       = "00000000-0000-4000-8000-000000000001"
	testAdminID      = "00000000-0000-4000-8000-000000000002"
	testOfficeID     = "00000000-0000-4000-8000-000000000030"
	testOfficeID2    = "00000000-0000-4000-8000-000000000031"
	testAssignmentID = "00000000-0000-4000-8000-000000000040"
)

func TestOfficeLocationServiceRules(t *testing.T) {
	t.Run("create berhasil", func(t *testing.T) {
		service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
		office, err := service.CreateOfficeLocation(context.Background(), OfficeLocationInput{Name: " Kantor Regional 3 ", Latitude: -6.1, Longitude: 106.8, RadiusMeters: 100})
		if err != nil {
			t.Fatalf("CreateOfficeLocation() error = %v", err)
		}
		if office.Name != "Kantor Regional 3" || !office.IsActive {
			t.Fatalf("office = %+v", office)
		}
	})

	tests := []struct {
		name  string
		input OfficeLocationInput
	}{
		{name: "latitude invalid", input: OfficeLocationInput{Name: "Kantor", Latitude: -91, Longitude: 106, RadiusMeters: 100}},
		{name: "longitude invalid", input: OfficeLocationInput{Name: "Kantor", Latitude: -6, Longitude: 181, RadiusMeters: 100}},
		{name: "radius terlalu kecil", input: OfficeLocationInput{Name: "Kantor", Latitude: -6, Longitude: 106, RadiusMeters: 9}},
		{name: "radius terlalu besar", input: OfficeLocationInput{Name: "Kantor", Latitude: -6, Longitude: 106, RadiusMeters: 2001}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
			_, err := service.CreateOfficeLocation(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
			}
		})
	}

	t.Run("update berhasil", func(t *testing.T) {
		service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
		office, err := service.UpdateOfficeLocation(context.Background(), testOfficeID, OfficeLocationInput{Name: "Kantor Baru", Latitude: -6.2, Longitude: 106.9, RadiusMeters: 150})
		if err != nil {
			t.Fatalf("UpdateOfficeLocation() error = %v", err)
		}
		if office.Name != "Kantor Baru" || !office.IsActive {
			t.Fatalf("office = %+v", office)
		}
	})

	t.Run("deactivate tanpa assignment berhasil", func(t *testing.T) {
		service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
		office, err := service.UpdateOfficeLocationStatus(context.Background(), testOfficeID, false)
		if err != nil {
			t.Fatalf("UpdateOfficeLocationStatus() error = %v", err)
		}
		if office.IsActive {
			t.Fatalf("office should be inactive")
		}
	})

	t.Run("deactivate dengan assignment current ditolak", func(t *testing.T) {
		repo := newFakeRepository()
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testOfficeID, "2026-08-01", nil)
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		_, err := service.UpdateOfficeLocationStatus(context.Background(), testOfficeID, false)
		if !errors.Is(err, ErrOfficeInUse) {
			t.Fatalf("error = %v, want %v", err, ErrOfficeInUse)
		}
	})

	t.Run("activate kembali berhasil dan list kosong", func(t *testing.T) {
		repo := newFakeRepository()
		office := repo.offices[testOfficeID]
		office.IsActive = false
		repo.offices[testOfficeID] = office
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		updated, err := service.UpdateOfficeLocationStatus(context.Background(), testOfficeID, true)
		if err != nil || !updated.IsActive {
			t.Fatalf("updated = %+v err=%v", updated, err)
		}
		repo.offices = map[string]OfficeLocation{}
		list, err := service.ListOfficeLocations(context.Background(), OfficeLocationListFilter{})
		if err != nil || len(list.Items) != 0 || list.Page != 1 || list.PageSize != 10 {
			t.Fatalf("list = %+v err=%v", list, err)
		}
		if list.Items == nil {
			t.Fatal("list.Items is nil, want empty slice")
		}
	})
}

func TestLocationAssignmentServiceRules(t *testing.T) {
	t.Run("create berhasil", func(t *testing.T) {
		service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
		assignment, err := service.CreateLocationAssignment(context.Background(), LocationAssignmentInput{UserID: testUserID, OfficeLocationID: testOfficeID, EffectiveFrom: "2026-08-06"})
		if err != nil {
			t.Fatalf("CreateLocationAssignment() error = %v", err)
		}
		if assignment.User.ID != testUserID || assignment.Office.ID != testOfficeID || assignment.Status != AssignmentStatusCurrent {
			t.Fatalf("assignment = %+v", assignment)
		}
	})

	t.Run("invalid dependencies ditolak", func(t *testing.T) {
		service := newTestService(newFakeRepository(), dateInJakarta(2026, 8, 6))
		cases := []struct {
			name  string
			input LocationAssignmentInput
			want  error
		}{
			{name: "user tidak ditemukan", input: LocationAssignmentInput{UserID: "00000000-0000-4000-8000-000000000099", OfficeLocationID: testOfficeID, EffectiveFrom: "2026-08-06"}, want: ErrInvalidUser},
			{name: "admin ditolak", input: LocationAssignmentInput{UserID: testAdminID, OfficeLocationID: testOfficeID, EffectiveFrom: "2026-08-06"}, want: ErrInvalidUser},
			{name: "lokasi tidak ditemukan", input: LocationAssignmentInput{UserID: testUserID, OfficeLocationID: "00000000-0000-4000-8000-000000000099", EffectiveFrom: "2026-08-06"}, want: ErrOfficeNotFound},
			{name: "tanggal invalid", input: LocationAssignmentInput{UserID: testUserID, OfficeLocationID: testOfficeID, EffectiveFrom: "invalid"}, want: ErrInvalidInput},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := service.CreateLocationAssignment(context.Background(), tt.input)
				if !errors.Is(err, tt.want) {
					t.Fatalf("error = %v, want %v", err, tt.want)
				}
			})
		}
	})

	t.Run("lokasi nonaktif ditolak", func(t *testing.T) {
		repo := newFakeRepository()
		office := repo.offices[testOfficeID]
		office.IsActive = false
		repo.offices[testOfficeID] = office
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		_, err := service.CreateLocationAssignment(context.Background(), LocationAssignmentInput{UserID: testUserID, OfficeLocationID: testOfficeID, EffectiveFrom: "2026-08-06"})
		if !errors.Is(err, ErrInactiveOffice) {
			t.Fatalf("error = %v, want %v", err, ErrInactiveOffice)
		}
	})

	t.Run("overlap ditolak dan berurutan diterima", func(t *testing.T) {
		repo := newFakeRepository()
		to := "2026-08-31"
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testOfficeID, "2026-08-01", &to)
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		nextTo := "2026-09-10"
		_, err := service.CreateLocationAssignment(context.Background(), LocationAssignmentInput{UserID: testUserID, OfficeLocationID: testOfficeID2, EffectiveFrom: "2026-08-20", EffectiveTo: &nextTo})
		if !errors.Is(err, ErrAssignmentOverlap) {
			t.Fatalf("overlap error = %v, want %v", err, ErrAssignmentOverlap)
		}
		if _, err := service.CreateLocationAssignment(context.Background(), LocationAssignmentInput{UserID: testUserID, OfficeLocationID: testOfficeID2, EffectiveFrom: "2026-09-01"}); err != nil {
			t.Fatalf("sequential assignment error = %v", err)
		}
	})

	t.Run("end assignment berhasil dan status timezone", func(t *testing.T) {
		repo := newFakeRepository()
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testOfficeID, "2026-08-06", nil)
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		ended, err := service.EndLocationAssignment(context.Background(), testAssignmentID, "2026-08-06")
		if err != nil || ended.EffectiveTo == nil || *ended.EffectiveTo != "2026-08-06" {
			t.Fatalf("ended = %+v err=%v", ended, err)
		}

		endedTo := "2026-08-05"
		repo.assignments["00000000-0000-4000-8000-000000000041"] = fakeAssignment("00000000-0000-4000-8000-000000000041", testUserID, testOfficeID, "2026-08-01", &endedTo)
		repo.assignments["00000000-0000-4000-8000-000000000042"] = fakeAssignment("00000000-0000-4000-8000-000000000042", testUserID, testOfficeID2, "2026-08-07", nil)
		listCurrent, _ := service.ListLocationAssignments(context.Background(), LocationAssignmentListFilter{Status: AssignmentStatusCurrent})
		listUpcoming, _ := service.ListLocationAssignments(context.Background(), LocationAssignmentListFilter{Status: AssignmentStatusUpcoming})
		listEnded, _ := service.ListLocationAssignments(context.Background(), LocationAssignmentListFilter{Status: AssignmentStatusEnded})
		if len(listCurrent.Items) != 1 || len(listUpcoming.Items) != 1 || len(listEnded.Items) != 1 {
			t.Fatalf("status counts = current:%d upcoming:%d ended:%d", len(listCurrent.Items), len(listUpcoming.Items), len(listEnded.Items))
		}
	})

	t.Run("location requirement user berhasil", func(t *testing.T) {
		repo := newFakeRepository()
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testOfficeID, "2026-08-06", nil)
		service := newTestService(repo, dateInJakarta(2026, 8, 6))
		requirement, err := service.LocationRequirement(context.Background(), auth.Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: testUserID}, Role: user.RoleUser})
		if err != nil || requirement.Office.ID != testOfficeID {
			t.Fatalf("requirement = %+v err=%v", requirement, err)
		}
	})
}

func newTestService(repo *fakeRepository, now time.Time) Service {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	service := NewService(repo, repo, location)
	service.now = func() time.Time { return now }
	return service
}

func dateInJakarta(year int, month time.Month, day int) time.Time {
	location, _ := time.LoadLocation("Asia/Jakarta")
	return time.Date(year, month, day, 8, 0, 0, 0, location)
}

type fakeRepository struct {
	users       map[string]user.User
	offices     map[string]OfficeLocation
	assignments map[string]LocationAssignmentRecord
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		users: map[string]user.User{
			testUserID:  {ID: testUserID, EmployeeNumber: "EMP-DUMMY-001", Name: "Pegawai Dummy", Email: "pegawai.dummy@example.test", Role: user.RoleUser, AccountStatus: user.AccountStatusActive},
			testAdminID: {ID: testAdminID, EmployeeNumber: "ADMIN-DUMMY", Name: "Admin Dummy", Email: "admin.dummy@example.test", Role: user.RoleAdmin, AccountStatus: user.AccountStatusActive},
		},
		offices: map[string]OfficeLocation{
			testOfficeID:  {ID: testOfficeID, Name: "Kantor Regional 3", Latitude: -6.1, Longitude: 106.8, RadiusMeters: 100, IsActive: true},
			testOfficeID2: {ID: testOfficeID2, Name: "Kantor Cadangan", Latitude: -6.2, Longitude: 106.9, RadiusMeters: 100, IsActive: true},
		},
		assignments: map[string]LocationAssignmentRecord{},
	}
}

func (r *fakeRepository) ListOfficeLocations(_ context.Context, filter OfficeLocationListFilter) ([]OfficeLocation, error) {
	var result []OfficeLocation
	for _, office := range r.offices {
		if filter.Status == OfficeStatusActive && !office.IsActive {
			continue
		}
		if filter.Status == OfficeStatusInactive && office.IsActive {
			continue
		}
		result = append(result, office)
	}
	return result, nil
}

func (r *fakeRepository) CountOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) (int, error) {
	items, _ := r.ListOfficeLocations(ctx, filter)
	return len(items), nil
}

func (r *fakeRepository) CreateOfficeLocation(_ context.Context, office OfficeLocation) (OfficeLocation, error) {
	r.offices[office.ID] = office
	return office, nil
}

func (r *fakeRepository) FindOfficeLocationByID(_ context.Context, id string) (OfficeLocation, error) {
	office, ok := r.offices[id]
	if !ok {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	return office, nil
}

func (r *fakeRepository) UpdateOfficeLocation(_ context.Context, office OfficeLocation) (OfficeLocation, error) {
	current, ok := r.offices[office.ID]
	if !ok {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	office.IsActive = current.IsActive
	r.offices[office.ID] = office
	return office, nil
}

func (r *fakeRepository) UpdateOfficeLocationStatus(_ context.Context, id string, isActive bool) (OfficeLocation, error) {
	office, ok := r.offices[id]
	if !ok {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	office.IsActive = isActive
	r.offices[id] = office
	return office, nil
}

func (r *fakeRepository) HasActiveOrFutureLocationAssignments(_ context.Context, officeID string, businessDate time.Time) (bool, error) {
	for _, assignment := range r.assignments {
		if assignment.Office.ID == officeID && (assignment.EffectiveTo == nil || !assignment.EffectiveTo.Before(businessDate)) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepository) ListLocationAssignments(_ context.Context, filter LocationAssignmentListFilter, businessDate time.Time) ([]LocationAssignmentRecord, error) {
	var result []LocationAssignmentRecord
	for _, assignment := range r.assignments {
		if filter.Status == AssignmentStatusCurrent && (assignment.EffectiveFrom.After(businessDate) || (assignment.EffectiveTo != nil && assignment.EffectiveTo.Before(businessDate))) {
			continue
		}
		if filter.Status == AssignmentStatusUpcoming && !assignment.EffectiveFrom.After(businessDate) {
			continue
		}
		if filter.Status == AssignmentStatusEnded && (assignment.EffectiveTo == nil || !assignment.EffectiveTo.Before(businessDate)) {
			continue
		}
		result = append(result, assignment)
	}
	return result, nil
}

func (r *fakeRepository) CountLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter, businessDate time.Time) (int, error) {
	items, _ := r.ListLocationAssignments(ctx, filter, businessDate)
	return len(items), nil
}

func (r *fakeRepository) FindLocationAssignmentByID(_ context.Context, id string) (LocationAssignmentRecord, error) {
	assignment, ok := r.assignments[id]
	if !ok {
		return LocationAssignmentRecord{}, ErrAssignmentNotFound
	}
	return assignment, nil
}

func (r *fakeRepository) FindUserByID(_ context.Context, id string) (user.User, error) {
	u, ok := r.users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func (r *fakeRepository) HasOverlappingLocationAssignment(_ context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error) {
	for _, assignment := range r.assignments {
		if assignment.User.ID != userID || assignment.ID == assignmentID {
			continue
		}
		if rangesOverlap(effectiveFrom, effectiveTo, assignment.EffectiveFrom, assignment.EffectiveTo) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepository) CreateLocationAssignment(_ context.Context, assignmentID string, userID string, officeID string, effectiveFrom time.Time, effectiveTo *time.Time) (LocationAssignmentRecord, error) {
	if overlap, _ := r.HasOverlappingLocationAssignment(context.Background(), userID, "", effectiveFrom, effectiveTo); overlap {
		return LocationAssignmentRecord{}, ErrAssignmentOverlap
	}
	assignment := LocationAssignmentRecord{ID: assignmentID, User: r.users[userID], Office: r.offices[officeID], EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo}
	r.assignments[assignmentID] = assignment
	return assignment, nil
}

func (r *fakeRepository) EndLocationAssignment(_ context.Context, id string, effectiveTo time.Time) (LocationAssignmentRecord, error) {
	assignment, ok := r.assignments[id]
	if !ok {
		return LocationAssignmentRecord{}, ErrAssignmentNotFound
	}
	assignment.EffectiveTo = &effectiveTo
	r.assignments[id] = assignment
	return assignment, nil
}

func (r *fakeRepository) FindCurrentLocationAssignment(_ context.Context, userID string, businessDate time.Time) (LocationAssignmentRecord, error) {
	for _, assignment := range r.assignments {
		if assignment.User.ID == userID && !assignment.EffectiveFrom.After(businessDate) && (assignment.EffectiveTo == nil || !assignment.EffectiveTo.Before(businessDate)) && assignment.Office.IsActive {
			return assignment, nil
		}
	}
	return LocationAssignmentRecord{}, ErrOfficeNotFound
}

func fakeAssignment(id string, userID string, officeID string, from string, to *string) LocationAssignmentRecord {
	repo := newFakeRepository()
	location, _ := time.LoadLocation("Asia/Jakarta")
	effectiveFrom, _ := parseRequiredDate(from, location)
	var effectiveTo *time.Time
	if to != nil {
		parsed, _ := parseRequiredDate(*to, location)
		effectiveTo = &parsed
	}
	return LocationAssignmentRecord{ID: id, User: repo.users[userID], Office: repo.offices[officeID], EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo}
}

func rangesOverlap(aFrom time.Time, aTo *time.Time, bFrom time.Time, bTo *time.Time) bool {
	if aTo != nil && aTo.Before(bFrom) {
		return false
	}
	if bTo != nil && bTo.Before(aFrom) {
		return false
	}
	return true
}
