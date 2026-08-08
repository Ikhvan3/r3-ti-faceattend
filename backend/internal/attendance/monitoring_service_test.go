package attendance

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAdminAttendanceMonitoringRepository struct {
	summaryRow adminAttendanceSummaryRow
	listRows   []adminAttendanceRow
	count      int
	detailRow  adminAttendanceRow
	detailErr  error
	lastQuery  adminAttendanceQuery
}

func (f *fakeAdminAttendanceMonitoringRepository) Summary(_ context.Context, _ time.Time, _ string) (adminAttendanceSummaryRow, error) {
	return f.summaryRow, nil
}

func (f *fakeAdminAttendanceMonitoringRepository) List(_ context.Context, query adminAttendanceQuery) ([]adminAttendanceRow, error) {
	f.lastQuery = query
	return f.listRows, nil
}

func (f *fakeAdminAttendanceMonitoringRepository) Count(_ context.Context, query adminAttendanceQuery) (int, error) {
	f.lastQuery = query
	return f.count, nil
}

func (f *fakeAdminAttendanceMonitoringRepository) Detail(_ context.Context, _ string, _ string) (adminAttendanceRow, error) {
	return f.detailRow, f.detailErr
}

func TestAdminAttendanceSummaryUsesBusinessDate(t *testing.T) {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	repo := &fakeAdminAttendanceMonitoringRepository{summaryRow: adminAttendanceSummaryRow{
		ActiveEmployees: 4,
		CheckedIn:       3,
		CheckedOut:      1,
		NotCheckedIn:    1,
		Late:            1,
	}}
	service := NewAdminAttendanceMonitoringService(repo, location)
	service.now = func() time.Time {
		return time.Date(2026, 8, 8, 23, 30, 0, 0, location)
	}

	result, err := service.Summary(context.Background(), "")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if result.Date != "2026-08-08" {
		t.Fatalf("Date = %q, want 2026-08-08", result.Date)
	}
	if result.ActiveEmployees != 4 || result.CheckedIn != 3 || result.NotCheckedIn != 1 || result.Late != 1 {
		t.Fatalf("unexpected summary: %+v", result)
	}
}

func TestAdminAttendanceListDefaultsToTodayAndPagination(t *testing.T) {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	repo := &fakeAdminAttendanceMonitoringRepository{count: 21}
	service := NewAdminAttendanceMonitoringService(repo, location)
	service.now = func() time.Time {
		return time.Date(2026, 8, 8, 8, 0, 0, 0, location)
	}

	result, err := service.List(context.Background(), AdminAttendanceListFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != defaultAdminPageSize || result.TotalPages != 3 {
		t.Fatalf("unexpected pagination: %+v", result)
	}
	if got := formatDate(repo.lastQuery.DateFrom); got != "2026-08-08" {
		t.Fatalf("DateFrom = %q", got)
	}
	if got := formatDate(repo.lastQuery.DateTo); got != "2026-08-08" {
		t.Fatalf("DateTo = %q", got)
	}
}

func TestAdminAttendanceListRejectsLongDateRange(t *testing.T) {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	service := NewAdminAttendanceMonitoringService(&fakeAdminAttendanceMonitoringRepository{}, location)

	_, err := service.List(context.Background(), AdminAttendanceListFilter{
		DateFrom: "2026-07-01",
		DateTo:   "2026-08-08",
	})
	if !errors.Is(err, ErrAdminAttendanceRange) {
		t.Fatalf("error = %v, want ErrAdminAttendanceRange", err)
	}
}

func TestAdminAttendanceListRejectsInvalidState(t *testing.T) {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	service := NewAdminAttendanceMonitoringService(&fakeAdminAttendanceMonitoringRepository{}, location)

	_, err := service.List(context.Background(), AdminAttendanceListFilter{
		AttendanceState: "LATE",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestAdminAttendanceDetailMapsNotFound(t *testing.T) {
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	repo := &fakeAdminAttendanceMonitoringRepository{detailErr: ErrAdminAttendanceNotFound}
	service := NewAdminAttendanceMonitoringService(repo, location)

	_, err := service.Detail(context.Background(), "11111111-1111-4111-8111-111111111111")
	if !errors.Is(err, ErrAdminAttendanceNotFound) {
		t.Fatalf("error = %v, want ErrAdminAttendanceNotFound", err)
	}
}
