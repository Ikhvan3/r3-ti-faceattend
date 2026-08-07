package attendance

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

func TestServiceTodayStates(t *testing.T) {
	checkInAt := time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC)
	checkOutAt := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		record      *AttendanceRecord
		wantState   AttendanceState
		canCheckIn  bool
		canCheckOut bool
	}{
		{name: "belum check-in", wantState: StateNotCheckedIn, canCheckIn: true},
		{name: "sudah check-in", record: &AttendanceRecord{ID: "record-id", CheckInAt: checkInAt}, wantState: StateCheckedIn, canCheckOut: true},
		{name: "selesai", record: &AttendanceRecord{ID: "record-id", CheckInAt: checkInAt, CheckOutAt: &checkOutAt}, wantState: StateCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeAttendanceRepository()
			repo.record = tt.record
			service := newTestService(repo, time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC))

			status, err := service.Today(context.Background(), userClaims())
			if err != nil {
				t.Fatalf("Today() error = %v", err)
			}

			if status.State != tt.wantState {
				t.Fatalf("state = %s, want %s", status.State, tt.wantState)
			}
			if status.CanCheckIn != tt.canCheckIn || status.CanCheckOut != tt.canCheckOut {
				t.Fatalf("can check in/out = %v/%v, want %v/%v", status.CanCheckIn, status.CanCheckOut, tt.canCheckIn, tt.canCheckOut)
			}
		})
	}
}

func TestServiceCheckInSuccessAndDuplicate(t *testing.T) {
	repo := newFakeAttendanceRepository()
	service := newTestService(repo, time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC))

	status, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest())
	if err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}
	if status.State != StateCheckedIn || !status.CanCheckOut {
		t.Fatalf("status after check-in = %+v", status)
	}

	_, err = service.CheckIn(context.Background(), userClaims(), validLocationRequest())
	if !errors.Is(err, ErrAlreadyCheckedIn) {
		t.Fatalf("second CheckIn() error = %v, want %v", err, ErrAlreadyCheckedIn)
	}
}

func TestServiceConcurrentCheckInDoesNotCreateTwoRecords(t *testing.T) {
	repo := newFakeAttendanceRepository()
	service := newTestService(repo, time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC))

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest())
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var success int
	var conflict int
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrAlreadyCheckedIn):
			conflict++
		default:
			t.Fatalf("unexpected error = %v", err)
		}
	}

	if success != 1 || conflict != 1 || repo.createdRecords != 1 {
		t.Fatalf("success/conflict/created = %d/%d/%d, want 1/1/1", success, conflict, repo.createdRecords)
	}
}

func TestServiceCheckOutRules(t *testing.T) {
	now := time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC)

	t.Run("berhasil", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)
		if _, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest()); err != nil {
			t.Fatalf("CheckIn() error = %v", err)
		}

		status, err := service.CheckOut(context.Background(), userClaims(), validLocationRequest())
		if err != nil {
			t.Fatalf("CheckOut() error = %v", err)
		}
		if status.State != StateCompleted {
			t.Fatalf("state = %s, want %s", status.State, StateCompleted)
		}
	})

	t.Run("sebelum check-in ditolak", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		_, err := service.CheckOut(context.Background(), userClaims(), validLocationRequest())
		if !errors.Is(err, ErrNotCheckedIn) {
			t.Fatalf("CheckOut() error = %v, want %v", err, ErrNotCheckedIn)
		}
	})

	t.Run("ganda ditolak", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)
		if _, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest()); err != nil {
			t.Fatalf("CheckIn() error = %v", err)
		}
		if _, err := service.CheckOut(context.Background(), userClaims(), validLocationRequest()); err != nil {
			t.Fatalf("CheckOut() error = %v", err)
		}

		_, err := service.CheckOut(context.Background(), userClaims(), validLocationRequest())
		if !errors.Is(err, ErrAlreadyCheckedOut) {
			t.Fatalf("second CheckOut() error = %v, want %v", err, ErrAlreadyCheckedOut)
		}
	})
}

func TestServiceGeofenceEvidenceRules(t *testing.T) {
	now := time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC)

	t.Run("check-in menyimpan evidence lokasi", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		status, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest())
		if err != nil {
			t.Fatalf("CheckIn() error = %v", err)
		}
		if status.CheckInLocation == nil {
			t.Fatal("CheckInLocation = nil, want evidence")
		}
		if status.CheckInLocation.OfficeLocationID != repo.location.OfficeLocationID || !status.CheckInLocation.InsideGeofence {
			t.Fatalf("CheckInLocation = %+v", status.CheckInLocation)
		}
		if repo.record == nil || repo.record.CheckInLocation == nil {
			t.Fatal("repository record missing check-in evidence")
		}
	})

	t.Run("check-out menyimpan evidence lokasi", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)
		if _, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest()); err != nil {
			t.Fatalf("CheckIn() error = %v", err)
		}

		status, err := service.CheckOut(context.Background(), userClaims(), validLocationRequest())
		if err != nil {
			t.Fatalf("CheckOut() error = %v", err)
		}
		if status.CheckOutLocation == nil {
			t.Fatal("CheckOutLocation = nil, want evidence")
		}
	})

	t.Run("akurasi lebih dari batas ditolak", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		request := validLocationRequest()
		request.AccuracyMeters = 51
		_, err := service.CheckIn(context.Background(), userClaims(), request)
		if !errors.Is(err, ErrPoorAccuracy) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrPoorAccuracy)
		}
	})

	t.Run("di luar radius ditolak", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		request := AttendanceLocationRequest{Latitude: -6.98, Longitude: 110.50, AccuracyMeters: 12.5}
		_, err := service.CheckIn(context.Background(), userClaims(), request)
		if !errors.Is(err, ErrOutsideGeofence) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrOutsideGeofence)
		}
	})

	t.Run("assignment lokasi tidak ada", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		repo.locationErr = ErrLocationNotFound
		service := newTestService(repo, now)

		_, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest())
		if !errors.Is(err, ErrLocationNotFound) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrLocationNotFound)
		}
	})

	t.Run("lokasi kantor nonaktif", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		repo.location.IsActive = false
		service := newTestService(repo, now)

		_, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest())
		if !errors.Is(err, ErrInactiveLocation) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrInactiveLocation)
		}
	})

	t.Run("koordinat tidak valid", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		request := validLocationRequest()
		request.Latitude = 100
		_, err := service.CheckIn(context.Background(), userClaims(), request)
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrInvalidInput)
		}
	})
}

func TestServiceAttendanceStateCheckedBeforeGeofence(t *testing.T) {
	now := time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC)

	t.Run("check-in ganda tetap konflik", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)
		if _, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest()); err != nil {
			t.Fatalf("CheckIn() error = %v", err)
		}

		request := validLocationRequest()
		request.Latitude = 100
		_, err := service.CheckIn(context.Background(), userClaims(), request)
		if !errors.Is(err, ErrAlreadyCheckedIn) {
			t.Fatalf("CheckIn() error = %v, want %v", err, ErrAlreadyCheckedIn)
		}
	})

	t.Run("check-out sebelum check-in tetap konflik", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		service := newTestService(repo, now)

		request := validLocationRequest()
		request.Latitude = 100
		_, err := service.CheckOut(context.Background(), userClaims(), request)
		if !errors.Is(err, ErrNotCheckedIn) {
			t.Fatalf("CheckOut() error = %v, want %v", err, ErrNotCheckedIn)
		}
	})
}

func TestServiceScheduleAndUserRules(t *testing.T) {
	now := time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC)

	t.Run("tanpa jadwal", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		repo.schedule.ID = ""
		service := newTestService(repo, now)

		_, err := service.Today(context.Background(), userClaims())
		if !errors.Is(err, ErrScheduleNotFound) {
			t.Fatalf("Today() error = %v, want %v", err, ErrScheduleNotFound)
		}
	})

	t.Run("jadwal nonaktif", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		repo.schedule.IsActive = false
		service := newTestService(repo, now)

		_, err := service.Today(context.Background(), userClaims())
		if !errors.Is(err, ErrInactiveSchedule) {
			t.Fatalf("Today() error = %v, want %v", err, ErrInactiveSchedule)
		}
	})

	t.Run("akun nonaktif", func(t *testing.T) {
		repo := newFakeAttendanceRepository()
		repo.user.AccountStatus = user.AccountStatusInactive
		service := newTestService(repo, now)

		_, err := service.Today(context.Background(), userClaims())
		if !errors.Is(err, ErrInactiveAccount) {
			t.Fatalf("Today() error = %v, want %v", err, ErrInactiveAccount)
		}
	})
}

func TestServiceUsesAsiaJakartaAttendanceDate(t *testing.T) {
	repo := newFakeAttendanceRepository()
	service := newTestService(repo, time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC))

	status, err := service.Today(context.Background(), userClaims())
	if err != nil {
		t.Fatalf("Today() error = %v", err)
	}

	if status.AttendanceDate != "2026-08-05" {
		t.Fatalf("attendance date = %s, want 2026-08-05", status.AttendanceDate)
	}
}

func TestServiceUsesBusinessDateForLocationAssignment(t *testing.T) {
	repo := newFakeAttendanceRepository()
	service := newTestService(repo, time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC))

	if _, err := service.CheckIn(context.Background(), userClaims(), validLocationRequest()); err != nil {
		t.Fatalf("CheckIn() error = %v", err)
	}

	if got := repo.locationAttendanceDate.Format("2006-01-02"); got != "2026-08-05" {
		t.Fatalf("location assignment date = %s, want 2026-08-05", got)
	}
	if repo.locationAttendanceDate.Location().String() != "Asia/Jakarta" {
		t.Fatalf("location assignment timezone = %s, want Asia/Jakarta", repo.locationAttendanceDate.Location())
	}
}

func TestServiceHistoryIncludesLocationEvidence(t *testing.T) {
	repo := newFakeAttendanceRepository()
	repo.history = []HistoryRow{{
		Record: AttendanceRecord{
			ID:              "record-id",
			AttendanceDate:  dateInJakarta(2026, 8, 5),
			CheckInAt:       time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC),
			CheckInLocation: validLocationEvidence(),
		},
		Schedule: repo.schedule,
	}}
	service := newTestService(repo, time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC))

	history, err := service.History(context.Background(), userClaims(), HistoryFilter{})
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}

	if len(history.Items) != 1 || history.Items[0].CheckInLocation == nil {
		t.Fatalf("history evidence = %+v", history.Items)
	}
}

func TestServiceHistoryEmptyAndPagination(t *testing.T) {
	repo := newFakeAttendanceRepository()
	service := newTestService(repo, time.Date(2026, 8, 5, 2, 0, 0, 0, time.UTC))

	empty, err := service.History(context.Background(), userClaims(), HistoryFilter{})
	if err != nil {
		t.Fatalf("History() empty error = %v", err)
	}
	if len(empty.Items) != 0 || empty.Page != 1 || empty.PageSize != 10 || empty.TotalItems != 0 {
		t.Fatalf("empty history = %+v", empty)
	}

	repo.history = []HistoryRow{
		{Record: AttendanceRecord{ID: "2", AttendanceDate: dateInJakarta(2026, 8, 5), CheckInAt: time.Now()}, Schedule: repo.schedule},
		{Record: AttendanceRecord{ID: "1", AttendanceDate: dateInJakarta(2026, 8, 4), CheckInAt: time.Now()}, Schedule: repo.schedule},
	}
	paged, err := service.History(context.Background(), userClaims(), HistoryFilter{Page: 2, PageSize: 1})
	if err != nil {
		t.Fatalf("History() paged error = %v", err)
	}
	if len(paged.Items) != 1 || paged.Items[0].ID != "1" || paged.TotalPages != 2 {
		t.Fatalf("paged history = %+v", paged)
	}
}

func newTestService(repo *fakeAttendanceRepository, now time.Time) Service {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	service := NewService(repo, location, 50)
	service.now = func() time.Time { return now }
	return service
}

func validLocationRequest() AttendanceLocationRequest {
	return AttendanceLocationRequest{
		Latitude:          -6.98946,
		Longitude:         110.416735,
		AccuracyMeters:    12.5,
		VerificationGrant: "valid-grant",
	}
}

func validLocationEvidence() *AttendanceLocationEvidence {
	return &AttendanceLocationEvidence{
		OfficeLocationID:   "00000000-0000-4000-8000-000000000020",
		OfficeLocationName: "Kantor PTPN I Regional 3 Semarang",
		Latitude:           -6.98946,
		Longitude:          110.416735,
		AccuracyMeters:     12.5,
		DistanceMeters:     0,
		InsideGeofence:     true,
	}
}

func userClaims() auth.Claims {
	return auth.Claims{
		Role: user.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "00000000-0000-4000-8000-000000000001",
		},
	}
}

func newFakeAttendanceRepository() *fakeAttendanceRepository {
	return &fakeAttendanceRepository{
		user: user.User{
			ID:            "00000000-0000-4000-8000-000000000001",
			Role:          user.RoleUser,
			AccountStatus: user.AccountStatusActive,
		},
		schedule: WorkSchedule{
			ID:           "00000000-0000-4000-8000-000000000010",
			Name:         "Jadwal Kerja Dummy TI",
			StartTime:    "08:00",
			EndTime:      "17:00",
			GraceMinutes: 15,
			IsActive:     true,
		},
		location: AttendanceLocationTarget{
			OfficeLocationID:   "00000000-0000-4000-8000-000000000020",
			OfficeLocationName: "Kantor PTPN I Regional 3 Semarang",
			Latitude:           -6.98946,
			Longitude:          110.416735,
			RadiusMeters:       100,
			IsActive:           true,
		},
	}
}

type fakeAttendanceRepository struct {
	mu                     sync.Mutex
	user                   user.User
	schedule               WorkSchedule
	location               AttendanceLocationTarget
	locationErr            error
	locationAttendanceDate time.Time
	record                 *AttendanceRecord
	history                []HistoryRow
	createdRecords         int
}

func (r *fakeAttendanceRepository) Today(_ context.Context, _ string, _ time.Time) (TodayData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return TodayData{User: r.user, Schedule: r.schedule, Record: cloneRecord(r.record)}, nil
}

func (r *fakeAttendanceRepository) CurrentOfficeLocation(_ context.Context, _ string, attendanceDate time.Time) (AttendanceLocationTarget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.locationAttendanceDate = attendanceDate
	if r.locationErr != nil {
		return AttendanceLocationTarget{}, r.locationErr
	}
	return r.location, nil
}

func (r *fakeAttendanceRepository) CheckIn(_ context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string, evidence AttendanceLocationEvidence) (AttendanceRecord, error) {
	return r.CheckInWithGrant(context.Background(), userID, attendanceDate, now, recordID, evidence, "")
}

func (r *fakeAttendanceRepository) CheckInWithGrant(_ context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string, evidence AttendanceLocationEvidence, _ string) (AttendanceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record != nil {
		return AttendanceRecord{}, ErrAlreadyCheckedIn
	}
	record := AttendanceRecord{
		ID:              recordID,
		UserID:          userID,
		ScheduleID:      r.schedule.ID,
		AttendanceDate:  attendanceDate,
		CheckInAt:       now,
		CheckInLocation: cloneEvidence(&evidence),
	}
	r.record = &record
	r.createdRecords++
	return record, nil
}

func (r *fakeAttendanceRepository) CheckOut(_ context.Context, _ string, _ time.Time, now time.Time, evidence AttendanceLocationEvidence) (AttendanceRecord, error) {
	return r.CheckOutWithGrant(context.Background(), "", time.Time{}, now, evidence, "")
}

func (r *fakeAttendanceRepository) CheckOutWithGrant(_ context.Context, _ string, _ time.Time, now time.Time, evidence AttendanceLocationEvidence, _ string) (AttendanceRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.record == nil {
		return AttendanceRecord{}, ErrNotCheckedIn
	}
	if r.record.CheckOutAt != nil {
		return AttendanceRecord{}, ErrAlreadyCheckedOut
	}
	r.record.CheckOutAt = &now
	r.record.CheckOutLocation = cloneEvidence(&evidence)
	return *cloneRecord(r.record), nil
}

func (r *fakeAttendanceRepository) ListHistory(_ context.Context, _ string, filter HistoryFilter) ([]HistoryRow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := (filter.Page - 1) * filter.PageSize
	if start >= len(r.history) {
		return []HistoryRow{}, nil
	}
	end := start + filter.PageSize
	if end > len(r.history) {
		end = len(r.history)
	}
	return append([]HistoryRow(nil), r.history[start:end]...), nil
}

func (r *fakeAttendanceRepository) CountHistory(_ context.Context, _ string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.history), nil
}

func cloneRecord(record *AttendanceRecord) *AttendanceRecord {
	if record == nil {
		return nil
	}
	cloned := *record
	cloned.CheckInLocation = cloneEvidence(record.CheckInLocation)
	cloned.CheckOutLocation = cloneEvidence(record.CheckOutLocation)
	return &cloned
}

func cloneEvidence(evidence *AttendanceLocationEvidence) *AttendanceLocationEvidence {
	if evidence == nil {
		return nil
	}
	cloned := *evidence
	return &cloned
}

func dateInJakarta(year int, month time.Month, day int) time.Time {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}
