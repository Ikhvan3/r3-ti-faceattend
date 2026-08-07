package attendance

import (
	"context"
	"errors"
	"strings"
	"time"
)

const maxAdminAttendanceRangeDays = 31

type AdminAttendanceMonitoringRepository interface {
	Summary(ctx context.Context, businessDate time.Time, timezone string) (adminAttendanceSummaryRow, error)
	List(ctx context.Context, query adminAttendanceQuery) ([]adminAttendanceRow, error)
	Count(ctx context.Context, query adminAttendanceQuery) (int, error)
	Detail(ctx context.Context, id string, timezone string) (adminAttendanceRow, error)
}

type AdminAttendanceMonitoringService struct {
	repo     AdminAttendanceMonitoringRepository
	location *time.Location
	now      func() time.Time
}

func NewAdminAttendanceMonitoringService(repo AdminAttendanceMonitoringRepository, location *time.Location) AdminAttendanceMonitoringService {
	return AdminAttendanceMonitoringService{repo: repo, location: location, now: time.Now}
}

func (s AdminAttendanceMonitoringService) Summary(ctx context.Context, dateValue string) (AdminAttendanceSummary, error) {
	businessDate, err := s.resolveBusinessDate(dateValue)
	if err != nil {
		return AdminAttendanceSummary{}, err
	}
	row, err := s.repo.Summary(ctx, businessDate, s.location.String())
	if err != nil {
		return AdminAttendanceSummary{}, ErrInternal
	}
	return AdminAttendanceSummary{
		Date:            formatDate(businessDate),
		ActiveEmployees: row.ActiveEmployees,
		CheckedIn:       row.CheckedIn,
		CheckedOut:      row.CheckedOut,
		NotCheckedIn:    row.NotCheckedIn,
		Late:            row.Late,
	}, nil
}

func (s AdminAttendanceMonitoringService) List(ctx context.Context, filter AdminAttendanceListFilter) (AdminAttendanceList, error) {
	query, err := s.normalizeListFilter(filter)
	if err != nil {
		return AdminAttendanceList{}, err
	}

	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return AdminAttendanceList{}, ErrInternal
	}
	rows, err := s.repo.List(ctx, query)
	if err != nil {
		return AdminAttendanceList{}, ErrInternal
	}

	items := make([]AdminAttendanceListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, AdminAttendanceListItem{
			ID:              row.ID,
			AttendanceDate:  formatDate(row.AttendanceDate),
			Employee:        row.Employee,
			Schedule:        row.Schedule,
			CheckInAt:       row.CheckInAt,
			CheckOutAt:      row.CheckOutAt,
			AttendanceState: row.AttendanceState,
			IsLate:          row.IsLate,
			OfficeLocation:  row.OfficeLocation,
		})
	}

	return AdminAttendanceList{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, query.PageSize),
	}, nil
}

func (s AdminAttendanceMonitoringService) Detail(ctx context.Context, id string) (AdminAttendanceDetail, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return AdminAttendanceDetail{}, ErrInvalidInput
	}

	row, err := s.repo.Detail(ctx, id, s.location.String())
	if errors.Is(err, ErrAdminAttendanceNotFound) {
		return AdminAttendanceDetail{}, err
	}
	if err != nil {
		return AdminAttendanceDetail{}, ErrInternal
	}
	if row.ID == nil {
		return AdminAttendanceDetail{}, ErrAdminAttendanceNotFound
	}

	return AdminAttendanceDetail{
		ID:               *row.ID,
		AttendanceDate:   formatDate(row.AttendanceDate),
		Employee:         row.Employee,
		Schedule:         row.Schedule,
		CheckInAt:        *row.CheckInAt,
		CheckOutAt:       row.CheckOutAt,
		AttendanceState:  row.AttendanceState,
		IsLate:           row.IsLate,
		CheckInLocation:  row.CheckInLocation,
		CheckOutLocation: row.CheckOutLocation,
	}, nil
}

func (s AdminAttendanceMonitoringService) normalizeListFilter(filter AdminAttendanceListFilter) (adminAttendanceQuery, error) {
	page, pageSize := normalizeAdminPagination(filter.Page, filter.PageSize)
	from, to, err := s.resolveDateRange(filter.DateFrom, filter.DateTo)
	if err != nil {
		return adminAttendanceQuery{}, err
	}

	employeeID := strings.TrimSpace(filter.EmployeeID)
	if employeeID != "" && !validAdminUUID(employeeID) {
		return adminAttendanceQuery{}, ErrInvalidInput
	}

	state := AdminAttendanceState(strings.TrimSpace(string(filter.AttendanceState)))
	if state != "" && state != AdminAttendanceStateNotCheckedIn && state != AdminAttendanceStateCheckedIn && state != AdminAttendanceStateCheckedOut {
		return adminAttendanceQuery{}, ErrInvalidInput
	}

	return adminAttendanceQuery{
		DateFrom:        from,
		DateTo:          to,
		EmployeeID:      employeeID,
		Search:          strings.TrimSpace(filter.Search),
		AttendanceState: state,
		IsLate:          filter.IsLate,
		Page:            page,
		PageSize:        pageSize,
		Timezone:        s.location.String(),
	}, nil
}

func (s AdminAttendanceMonitoringService) resolveBusinessDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		now := s.now().In(s.location)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location), nil
	}
	return parseRequiredDate(value, s.location)
}

func (s AdminAttendanceMonitoringService) resolveDateRange(fromValue, toValue string) (time.Time, time.Time, error) {
	fromValue = strings.TrimSpace(fromValue)
	toValue = strings.TrimSpace(toValue)
	if fromValue == "" && toValue == "" {
		date, err := s.resolveBusinessDate("")
		return date, date, err
	}
	if fromValue == "" || toValue == "" {
		return time.Time{}, time.Time{}, ErrAdminAttendanceRange
	}

	from, err := parseRequiredDate(fromValue, s.location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := parseRequiredDate(toValue, s.location)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, ErrAdminAttendanceRange
	}
	if int(to.Sub(from).Hours()/24)+1 > maxAdminAttendanceRangeDays {
		return time.Time{}, time.Time{}, ErrAdminAttendanceRange
	}
	return from, to, nil
}
