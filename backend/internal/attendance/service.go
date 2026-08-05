package attendance

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

const (
	defaultHistoryPageSize = 10
	maxHistoryPageSize     = 100
)

type AttendanceRepository interface {
	Today(ctx context.Context, userID string, attendanceDate time.Time) (TodayData, error)
	CheckIn(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string) (AttendanceRecord, error)
	CheckOut(ctx context.Context, userID string, attendanceDate time.Time, now time.Time) (AttendanceRecord, error)
	ListHistory(ctx context.Context, userID string, filter HistoryFilter) ([]HistoryRow, error)
	CountHistory(ctx context.Context, userID string) (int, error)
}

type Service struct {
	repo     AttendanceRepository
	location *time.Location
	now      func() time.Time
}

func NewService(repo AttendanceRepository, location *time.Location) Service {
	return Service{repo: repo, location: location, now: time.Now}
}

func (s Service) Today(ctx context.Context, claims auth.Claims) (DailyStatus, error) {
	userID, attendanceDate, err := s.requestContext(claims)
	if err != nil {
		return DailyStatus{}, err
	}

	data, err := s.repo.Today(ctx, userID, attendanceDate)
	if err != nil {
		return DailyStatus{}, mapRepositoryError(err)
	}
	if err := validateTodayData(data); err != nil {
		return DailyStatus{}, err
	}

	return dailyStatus(attendanceDate, data.Schedule, data.Record), nil
}

func (s Service) CheckIn(ctx context.Context, claims auth.Claims) (DailyStatus, error) {
	userID, attendanceDate, err := s.requestContext(claims)
	if err != nil {
		return DailyStatus{}, err
	}

	data, err := s.repo.Today(ctx, userID, attendanceDate)
	if err != nil {
		return DailyStatus{}, mapRepositoryError(err)
	}
	if err := validateTodayData(data); err != nil {
		return DailyStatus{}, err
	}
	if data.Record != nil {
		return DailyStatus{}, ErrAlreadyCheckedIn
	}

	record, err := s.repo.CheckIn(ctx, userID, attendanceDate, s.now().UTC(), newUUID())
	if err != nil {
		return DailyStatus{}, mapRepositoryError(err)
	}

	return dailyStatus(attendanceDate, data.Schedule, &record), nil
}

func (s Service) CheckOut(ctx context.Context, claims auth.Claims) (DailyStatus, error) {
	userID, attendanceDate, err := s.requestContext(claims)
	if err != nil {
		return DailyStatus{}, err
	}

	data, err := s.repo.Today(ctx, userID, attendanceDate)
	if err != nil {
		return DailyStatus{}, mapRepositoryError(err)
	}
	if err := validateTodayData(data); err != nil {
		return DailyStatus{}, err
	}
	if data.Record == nil {
		return DailyStatus{}, ErrNotCheckedIn
	}
	if data.Record.CheckOutAt != nil {
		return DailyStatus{}, ErrAlreadyCheckedOut
	}

	record, err := s.repo.CheckOut(ctx, userID, attendanceDate, s.now().UTC())
	if err != nil {
		return DailyStatus{}, mapRepositoryError(err)
	}

	return dailyStatus(attendanceDate, data.Schedule, &record), nil
}

func (s Service) History(ctx context.Context, claims auth.Claims, filter HistoryFilter) (HistoryList, error) {
	userID, _, err := s.requestContext(claims)
	if err != nil {
		return HistoryList{}, err
	}

	normalized, err := normalizeHistoryFilter(filter)
	if err != nil {
		return HistoryList{}, err
	}

	total, err := s.repo.CountHistory(ctx, userID)
	if err != nil {
		return HistoryList{}, ErrInternal
	}

	rows, err := s.repo.ListHistory(ctx, userID, normalized)
	if err != nil {
		return HistoryList{}, ErrInternal
	}

	items := make([]HistoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyItem(row))
	}

	return HistoryList{
		Items:      items,
		Page:       normalized.Page,
		PageSize:   normalized.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, normalized.PageSize),
	}, nil
}

func (s Service) requestContext(claims auth.Claims) (string, time.Time, error) {
	if claims.Subject == "" || claims.Role != user.RoleUser {
		return "", time.Time{}, ErrInvalidInput
	}

	return claims.Subject, attendanceDate(s.now(), s.location), nil
}

func validateTodayData(data TodayData) error {
	if data.User.ID == "" || data.User.AccountStatus != user.AccountStatusActive {
		return ErrInactiveAccount
	}
	if data.Schedule.ID == "" {
		return ErrScheduleNotFound
	}
	if !data.Schedule.IsActive {
		return ErrInactiveSchedule
	}

	return nil
}

func attendanceDate(now time.Time, location *time.Location) time.Time {
	local := now.In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}

func dailyStatus(attendanceDate time.Time, schedule WorkSchedule, record *AttendanceRecord) DailyStatus {
	status := DailyStatus{
		AttendanceDate: formatDate(attendanceDate),
		Schedule:       schedule,
		State:          StateNotCheckedIn,
		CanCheckIn:     true,
		CanCheckOut:    false,
	}
	if record == nil {
		return status
	}

	status.CheckInAt = &record.CheckInAt
	status.CheckOutAt = record.CheckOutAt
	status.CanCheckIn = false
	if record.CheckOutAt == nil {
		status.State = StateCheckedIn
		status.CanCheckOut = true
		return status
	}

	status.State = StateCompleted
	status.CanCheckOut = false
	return status
}

func historyItem(row HistoryRow) HistoryItem {
	state := StateCheckedIn
	if row.Record.CheckOutAt != nil {
		state = StateCompleted
	}

	return HistoryItem{
		ID:             row.Record.ID,
		AttendanceDate: formatDate(row.Record.AttendanceDate),
		Schedule:       row.Schedule,
		CheckInAt:      row.Record.CheckInAt,
		CheckOutAt:     row.Record.CheckOutAt,
		State:          state,
	}
}

func normalizeHistoryFilter(filter HistoryFilter) (HistoryFilter, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = defaultHistoryPageSize
	}
	if pageSize > maxHistoryPageSize {
		pageSize = maxHistoryPageSize
	}

	return HistoryFilter{Page: page, PageSize: pageSize}, nil
}

func totalPages(totalItems int, pageSize int) int {
	if totalItems == 0 {
		return 0
	}

	return (totalItems + pageSize - 1) / pageSize
}

func formatDate(value time.Time) string {
	return value.Format("2006-01-02")
}

func mapRepositoryError(err error) error {
	switch err {
	case ErrInactiveAccount, ErrScheduleNotFound, ErrInactiveSchedule, ErrAlreadyCheckedIn, ErrNotCheckedIn, ErrAlreadyCheckedOut:
		return err
	default:
		return ErrInternal
	}
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("generate uuid: %w", err))
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
