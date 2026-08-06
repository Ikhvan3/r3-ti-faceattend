package attendance

import (
	"context"
	"errors"
	"testing"
	"time"

	"r3-ti-faceattend/backend/internal/user"
)

const (
	testUserID       = "00000000-0000-4000-8000-000000000001"
	testAdminID      = "00000000-0000-4000-8000-000000000002"
	testScheduleID   = "00000000-0000-4000-8000-000000000010"
	testScheduleID2  = "00000000-0000-4000-8000-000000000011"
	testAssignmentID = "00000000-0000-4000-8000-000000000020"
)

func TestAdminScheduleServiceWorkScheduleRules(t *testing.T) {
	t.Run("create berhasil", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		schedule, err := service.CreateWorkSchedule(context.Background(), WorkScheduleInput{Name: " Jadwal Kerja Reguler ", StartTime: "08:00", EndTime: "17:00", GraceMinutes: 15})
		if err != nil {
			t.Fatalf("CreateWorkSchedule() error = %v", err)
		}
		if schedule.Name != "Jadwal Kerja Reguler" || !schedule.IsActive {
			t.Fatalf("schedule = %+v", schedule)
		}
	})

	t.Run("end sebelum start ditolak", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateWorkSchedule(context.Background(), WorkScheduleInput{Name: "Jadwal", StartTime: "17:00", EndTime: "08:00"})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("grace negatif ditolak", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateWorkSchedule(context.Background(), WorkScheduleInput{Name: "Jadwal", StartTime: "08:00", EndTime: "17:00", GraceMinutes: -1})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("list kosong berhasil", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		repo.schedules = map[string]WorkSchedule{}
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		result, err := service.ListWorkSchedules(context.Background(), WorkScheduleListFilter{})
		if err != nil {
			t.Fatalf("ListWorkSchedules() error = %v", err)
		}
		if len(result.Items) != 0 || result.Page != 1 || result.PageSize != 10 {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("update berhasil", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		schedule, err := service.UpdateWorkSchedule(context.Background(), testScheduleID, WorkScheduleInput{Name: "Jadwal Baru", StartTime: "07:30", EndTime: "16:30", GraceMinutes: 10})
		if err != nil {
			t.Fatalf("UpdateWorkSchedule() error = %v", err)
		}
		if schedule.Name != "Jadwal Baru" || !schedule.IsActive {
			t.Fatalf("schedule = %+v", schedule)
		}
	})

	t.Run("deactivate tanpa assignment berhasil", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		schedule, err := service.UpdateWorkScheduleStatus(context.Background(), testScheduleID, false)
		if err != nil {
			t.Fatalf("UpdateWorkScheduleStatus() error = %v", err)
		}
		if schedule.IsActive {
			t.Fatalf("schedule should be inactive")
		}
	})

	t.Run("deactivate dengan assignment aktif ditolak", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testScheduleID, "2026-08-01", nil)
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		_, err := service.UpdateWorkScheduleStatus(context.Background(), testScheduleID, false)
		if !errors.Is(err, ErrScheduleInUse) {
			t.Fatalf("error = %v, want %v", err, ErrScheduleInUse)
		}
	})

	t.Run("aktivasi kembali berhasil", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		schedule := repo.schedules[testScheduleID]
		schedule.IsActive = false
		repo.schedules[testScheduleID] = schedule
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		updated, err := service.UpdateWorkScheduleStatus(context.Background(), testScheduleID, true)
		if err != nil {
			t.Fatalf("UpdateWorkScheduleStatus() error = %v", err)
		}
		if !updated.IsActive {
			t.Fatalf("schedule should be active")
		}
	})
}

func TestAdminScheduleServiceAssignmentRules(t *testing.T) {
	t.Run("create berhasil", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		assignment, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: testScheduleID, EffectiveFrom: "2026-08-05"})
		if err != nil {
			t.Fatalf("CreateAssignment() error = %v", err)
		}
		if assignment.User.ID != testUserID || assignment.Schedule.ID != testScheduleID || assignment.EffectiveFrom != "2026-08-05" {
			t.Fatalf("assignment = %+v", assignment)
		}
	})

	t.Run("user tidak ditemukan", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: "00000000-0000-4000-8000-000000000099", ScheduleID: testScheduleID, EffectiveFrom: "2026-08-05"})
		if !errors.Is(err, ErrAssignmentInvalidUser) {
			t.Fatalf("error = %v, want %v", err, ErrAssignmentInvalidUser)
		}
	})

	t.Run("user admin ditolak", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testAdminID, ScheduleID: testScheduleID, EffectiveFrom: "2026-08-05"})
		if !errors.Is(err, ErrAssignmentInvalidUser) {
			t.Fatalf("error = %v, want %v", err, ErrAssignmentInvalidUser)
		}
	})

	t.Run("schedule tidak ditemukan", func(t *testing.T) {
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: "00000000-0000-4000-8000-000000000099", EffectiveFrom: "2026-08-05"})
		if !errors.Is(err, ErrScheduleNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrScheduleNotFound)
		}
	})

	t.Run("schedule nonaktif ditolak", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		schedule := repo.schedules[testScheduleID]
		schedule.IsActive = false
		repo.schedules[testScheduleID] = schedule
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: testScheduleID, EffectiveFrom: "2026-08-05"})
		if !errors.Is(err, ErrInactiveSchedule) {
			t.Fatalf("error = %v, want %v", err, ErrInactiveSchedule)
		}
	})

	t.Run("effective_to sebelum effective_from ditolak", func(t *testing.T) {
		to := "2026-08-04"
		service := newTestAdminScheduleService(newFakeAdminScheduleRepository(), dateInJakarta(2026, 8, 5))
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: testScheduleID, EffectiveFrom: "2026-08-05", EffectiveTo: &to})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("overlap ditolak dan berurutan diterima", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		to := "2026-08-31"
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testScheduleID, "2026-08-01", &to)
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		nextTo := "2026-09-10"
		_, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: testScheduleID2, EffectiveFrom: "2026-08-20", EffectiveTo: &nextTo})
		if !errors.Is(err, ErrAssignmentOverlap) {
			t.Fatalf("overlap error = %v, want %v", err, ErrAssignmentOverlap)
		}
		if _, err := service.CreateAssignment(context.Background(), AssignmentCreateInput{UserID: testUserID, ScheduleID: testScheduleID2, EffectiveFrom: "2026-09-01"}); err != nil {
			t.Fatalf("sequential assignment error = %v", err)
		}
	})

	t.Run("end assignment berhasil dan invalid ditolak", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		repo.assignments[testAssignmentID] = fakeAssignment(testAssignmentID, testUserID, testScheduleID, "2026-08-05", nil)
		service := newTestAdminScheduleService(repo, dateInJakarta(2026, 8, 5))
		assignment, err := service.EndAssignment(context.Background(), testAssignmentID, "2026-08-05")
		if err != nil {
			t.Fatalf("EndAssignment() error = %v", err)
		}
		if assignment.EffectiveTo == nil || *assignment.EffectiveTo != "2026-08-05" {
			t.Fatalf("assignment = %+v", assignment)
		}
		_, err = service.EndAssignment(context.Background(), testAssignmentID, "2026-08-04")
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("invalid end error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("list kosong dan status timezone", func(t *testing.T) {
		repo := newFakeAdminScheduleRepository()
		repo.assignments = map[string]ScheduleAssignmentRecord{}
		service := newTestAdminScheduleService(repo, time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC))
		empty, err := service.ListAssignments(context.Background(), AssignmentListFilter{})
		if err != nil {
			t.Fatalf("ListAssignments() error = %v", err)
		}
		if len(empty.Items) != 0 || empty.Page != 1 || empty.PageSize != 10 {
			t.Fatalf("empty = %+v", empty)
		}

		endedTo := "2026-08-04"
		repo.assignments["00000000-0000-4000-8000-000000000021"] = fakeAssignment("00000000-0000-4000-8000-000000000021", testUserID, testScheduleID, "2026-08-01", &endedTo)
		repo.assignments["00000000-0000-4000-8000-000000000022"] = fakeAssignment("00000000-0000-4000-8000-000000000022", testUserID, testScheduleID2, "2026-08-05", nil)
		repo.assignments["00000000-0000-4000-8000-000000000023"] = fakeAssignment("00000000-0000-4000-8000-000000000023", testUserID, testScheduleID2, "2026-08-06", nil)

		current, err := service.ListAssignments(context.Background(), AssignmentListFilter{Status: AssignmentStatusCurrent})
		if err != nil || len(current.Items) != 1 || current.Items[0].EffectiveFrom != "2026-08-05" {
			t.Fatalf("current = %+v err=%v", current, err)
		}
		upcoming, _ := service.ListAssignments(context.Background(), AssignmentListFilter{Status: AssignmentStatusUpcoming})
		ended, _ := service.ListAssignments(context.Background(), AssignmentListFilter{Status: AssignmentStatusEnded})
		if len(upcoming.Items) != 1 || len(ended.Items) != 1 {
			t.Fatalf("upcoming/ended = %d/%d", len(upcoming.Items), len(ended.Items))
		}
	})
}

func newTestAdminScheduleService(repo *fakeAdminScheduleRepository, now time.Time) AdminScheduleService {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	service := NewAdminScheduleService(repo, location)
	service.now = func() time.Time { return now }
	return service
}

func newFakeAdminScheduleRepository() *fakeAdminScheduleRepository {
	return &fakeAdminScheduleRepository{
		users: map[string]user.User{
			testUserID:  {ID: testUserID, EmployeeNumber: "EMP-DUMMY-001", Name: "Pegawai Dummy", Email: "pegawai.dummy@example.test", Role: user.RoleUser, AccountStatus: user.AccountStatusActive},
			testAdminID: {ID: testAdminID, EmployeeNumber: "ADMIN-DUMMY", Name: "Admin Dummy", Email: "admin.dummy@example.test", Role: user.RoleAdmin, AccountStatus: user.AccountStatusActive},
		},
		schedules: map[string]WorkSchedule{
			testScheduleID:  {ID: testScheduleID, Name: "Jadwal Reguler", StartTime: "08:00", EndTime: "17:00", GraceMinutes: 15, IsActive: true},
			testScheduleID2: {ID: testScheduleID2, Name: "Jadwal Pagi", StartTime: "07:00", EndTime: "16:00", GraceMinutes: 10, IsActive: true},
		},
		assignments: map[string]ScheduleAssignmentRecord{},
	}
}

type fakeAdminScheduleRepository struct {
	users       map[string]user.User
	schedules   map[string]WorkSchedule
	assignments map[string]ScheduleAssignmentRecord
}

func (r *fakeAdminScheduleRepository) ListWorkSchedules(_ context.Context, filter WorkScheduleListFilter) ([]WorkSchedule, error) {
	result := make([]WorkSchedule, 0, len(r.schedules))
	for _, schedule := range r.schedules {
		if filter.Status == ScheduleStatusActive && !schedule.IsActive {
			continue
		}
		if filter.Status == ScheduleStatusInactive && schedule.IsActive {
			continue
		}
		result = append(result, schedule)
	}
	return result, nil
}

func (r *fakeAdminScheduleRepository) CountWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) (int, error) {
	items, _ := r.ListWorkSchedules(ctx, filter)
	return len(items), nil
}

func (r *fakeAdminScheduleRepository) CreateWorkSchedule(_ context.Context, schedule WorkSchedule) (WorkSchedule, error) {
	r.schedules[schedule.ID] = schedule
	return schedule, nil
}

func (r *fakeAdminScheduleRepository) FindWorkScheduleByID(_ context.Context, id string) (WorkSchedule, error) {
	schedule, ok := r.schedules[id]
	if !ok {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	return schedule, nil
}

func (r *fakeAdminScheduleRepository) UpdateWorkSchedule(_ context.Context, schedule WorkSchedule) (WorkSchedule, error) {
	current, ok := r.schedules[schedule.ID]
	if !ok {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	schedule.IsActive = current.IsActive
	r.schedules[schedule.ID] = schedule
	return schedule, nil
}

func (r *fakeAdminScheduleRepository) UpdateWorkScheduleStatus(_ context.Context, id string, isActive bool) (WorkSchedule, error) {
	schedule, ok := r.schedules[id]
	if !ok {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	schedule.IsActive = isActive
	r.schedules[id] = schedule
	return schedule, nil
}

func (r *fakeAdminScheduleRepository) HasActiveOrFutureAssignments(_ context.Context, scheduleID string, businessDate time.Time) (bool, error) {
	for _, assignment := range r.assignments {
		if assignment.Schedule.ID == scheduleID && (assignment.EffectiveTo == nil || !assignment.EffectiveTo.Before(businessDate)) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeAdminScheduleRepository) ListAssignments(_ context.Context, filter AssignmentListFilter, businessDate time.Time) ([]ScheduleAssignmentRecord, error) {
	var result []ScheduleAssignmentRecord
	for _, assignment := range r.assignments {
		if filter.Status == AssignmentStatusCurrent && !(assignment.EffectiveFrom.Before(businessDate) || assignment.EffectiveFrom.Equal(businessDate)) {
			continue
		}
		if filter.Status == AssignmentStatusCurrent && assignment.EffectiveTo != nil && assignment.EffectiveTo.Before(businessDate) {
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

func (r *fakeAdminScheduleRepository) CountAssignments(ctx context.Context, filter AssignmentListFilter, businessDate time.Time) (int, error) {
	items, _ := r.ListAssignments(ctx, filter, businessDate)
	return len(items), nil
}

func (r *fakeAdminScheduleRepository) FindAssignmentByID(_ context.Context, id string) (ScheduleAssignmentRecord, error) {
	assignment, ok := r.assignments[id]
	if !ok {
		return ScheduleAssignmentRecord{}, ErrAssignmentNotFound
	}
	return assignment, nil
}

func (r *fakeAdminScheduleRepository) FindUserByID(_ context.Context, id string) (user.User, error) {
	u, ok := r.users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func (r *fakeAdminScheduleRepository) HasOverlappingAssignment(_ context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error) {
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

func (r *fakeAdminScheduleRepository) CreateAssignment(_ context.Context, assignmentID string, userID string, scheduleID string, effectiveFrom time.Time, effectiveTo *time.Time) (ScheduleAssignmentRecord, error) {
	if overlap, _ := r.HasOverlappingAssignment(context.Background(), userID, "", effectiveFrom, effectiveTo); overlap {
		return ScheduleAssignmentRecord{}, ErrAssignmentOverlap
	}
	assignment := ScheduleAssignmentRecord{ID: assignmentID, User: r.users[userID], Schedule: r.schedules[scheduleID], EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo}
	r.assignments[assignmentID] = assignment
	return assignment, nil
}

func (r *fakeAdminScheduleRepository) EndAssignment(_ context.Context, id string, effectiveTo time.Time) (ScheduleAssignmentRecord, error) {
	assignment, ok := r.assignments[id]
	if !ok {
		return ScheduleAssignmentRecord{}, ErrAssignmentNotFound
	}
	assignment.EffectiveTo = &effectiveTo
	r.assignments[id] = assignment
	return assignment, nil
}

func fakeAssignment(id string, userID string, scheduleID string, from string, to *string) ScheduleAssignmentRecord {
	location, _ := time.LoadLocation("Asia/Jakarta")
	effectiveFrom, _ := parseRequiredDate(from, location)
	var effectiveTo *time.Time
	if to != nil {
		parsed, _ := parseRequiredDate(*to, location)
		effectiveTo = &parsed
	}
	repo := newFakeAdminScheduleRepository()
	return ScheduleAssignmentRecord{ID: id, User: repo.users[userID], Schedule: repo.schedules[scheduleID], EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo}
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
