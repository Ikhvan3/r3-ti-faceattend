package audit

import (
	"context"
	"errors"
	"testing"
	"time"
)

const testAuditEntityID = "00000000-0000-4000-8000-000000000099"

type fakeRepository struct {
	items []Log
	total int
	err   error
	last  listQuery
}

func (f *fakeRepository) List(_ context.Context, filter listQuery) ([]Log, error) {
	f.last = filter
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeRepository) Count(_ context.Context, filter listQuery) (int, error) {
	f.last = filter
	if f.err != nil {
		return 0, f.err
	}
	return f.total, nil
}

func TestServiceListNormalizesPaginationRangeAndEntity(t *testing.T) {
	repo := &fakeRepository{total: 41, items: []Log{{ID: "audit-1"}}}
	location := time.FixedZone("Asia/Jakarta", 7*60*60)
	service := NewService(repo, location)

	result, err := service.List(context.Background(), ListFilter{
		Action:     ActionAttendanceCorrected,
		EntityType: EntityAttendanceRecord,
		EntityID:   testAuditEntityID,
		DateFrom:   "2026-08-01",
		DateTo:     "2026-08-08",
		Page:       2,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Page != 2 || result.PageSize != 20 || result.TotalPages != 3 || result.TotalItems != 41 {
		t.Fatalf("result = %+v", result)
	}
	if repo.last.EntityID != testAuditEntityID {
		t.Fatalf("entity id = %q, want %q", repo.last.EntityID, testAuditEntityID)
	}
	if repo.last.DateFrom == nil || repo.last.DateTo == nil {
		t.Fatal("date range was not normalized")
	}
	if got := repo.last.DateFrom.In(location).Format("2006-01-02"); got != "2026-08-01" {
		t.Fatalf("date from = %s", got)
	}
	if got := repo.last.DateTo.In(location).Format("2006-01-02"); got != "2026-08-09" {
		t.Fatalf("date to exclusive = %s", got)
	}
}

func TestServiceListRejectsInvalidFilter(t *testing.T) {
	service := NewService(&fakeRepository{}, time.UTC)
	tests := []ListFilter{
		{Action: Action("UNKNOWN")},
		{EntityType: EntityType("UNKNOWN")},
		{EntityID: "not-a-uuid"},
		{DateFrom: "2026-08-01"},
		{DateFrom: "2026-08-08", DateTo: "2026-08-01"},
		{DateFrom: "2026-01-01", DateTo: "2026-08-08"},
	}
	for _, filter := range tests {
		if _, err := service.List(context.Background(), filter); !errors.Is(err, ErrInvalidFilter) {
			t.Fatalf("List(%+v) error = %v, want ErrInvalidFilter", filter, err)
		}
	}
}
