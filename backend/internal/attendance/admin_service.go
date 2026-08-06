package attendance

import (
	"context"
	"errors"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/user"
)

const (
	defaultAdminPageSize = 10
	maxAdminPageSize     = 100
	maxGraceMinutes      = 240
)

type AdminScheduleRepository interface {
	ListWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) ([]WorkSchedule, error)
	CountWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) (int, error)
	CreateWorkSchedule(ctx context.Context, schedule WorkSchedule) (WorkSchedule, error)
	FindWorkScheduleByID(ctx context.Context, id string) (WorkSchedule, error)
	UpdateWorkSchedule(ctx context.Context, schedule WorkSchedule) (WorkSchedule, error)
	UpdateWorkScheduleStatus(ctx context.Context, id string, isActive bool) (WorkSchedule, error)
	HasActiveOrFutureAssignments(ctx context.Context, scheduleID string, businessDate time.Time) (bool, error)
	ListAssignments(ctx context.Context, filter AssignmentListFilter, businessDate time.Time) ([]ScheduleAssignmentRecord, error)
	CountAssignments(ctx context.Context, filter AssignmentListFilter, businessDate time.Time) (int, error)
	FindAssignmentByID(ctx context.Context, id string) (ScheduleAssignmentRecord, error)
	FindUserByID(ctx context.Context, id string) (user.User, error)
	HasOverlappingAssignment(ctx context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error)
	CreateAssignment(ctx context.Context, assignmentID string, userID string, scheduleID string, effectiveFrom time.Time, effectiveTo *time.Time) (ScheduleAssignmentRecord, error)
	EndAssignment(ctx context.Context, id string, effectiveTo time.Time) (ScheduleAssignmentRecord, error)
}

type AdminScheduleService struct {
	repo     AdminScheduleRepository
	location *time.Location
	now      func() time.Time
}

func NewAdminScheduleService(repo AdminScheduleRepository, location *time.Location) AdminScheduleService {
	return AdminScheduleService{repo: repo, location: location, now: time.Now}
}

func (s AdminScheduleService) ListWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) (WorkScheduleList, error) {
	normalized, err := normalizeWorkScheduleListFilter(filter)
	if err != nil {
		return WorkScheduleList{}, err
	}

	total, err := s.repo.CountWorkSchedules(ctx, normalized)
	if err != nil {
		return WorkScheduleList{}, ErrInternal
	}
	items, err := s.repo.ListWorkSchedules(ctx, normalized)
	if err != nil {
		return WorkScheduleList{}, ErrInternal
	}

	return WorkScheduleList{Items: items, Page: normalized.Page, PageSize: normalized.PageSize, TotalItems: total, TotalPages: totalPages(total, normalized.PageSize)}, nil
}

func (s AdminScheduleService) CreateWorkSchedule(ctx context.Context, input WorkScheduleInput) (WorkSchedule, error) {
	normalized, err := normalizeWorkScheduleInput(input)
	if err != nil {
		return WorkSchedule{}, err
	}

	created, err := s.repo.CreateWorkSchedule(ctx, WorkSchedule{
		ID:           newUUID(),
		Name:         normalized.Name,
		StartTime:    normalized.StartTime,
		EndTime:      normalized.EndTime,
		GraceMinutes: normalized.GraceMinutes,
		IsActive:     true,
	})
	if err != nil {
		return WorkSchedule{}, mapAdminRepositoryError(err)
	}

	return created, nil
}

func (s AdminScheduleService) WorkScheduleDetail(ctx context.Context, id string) (WorkSchedule, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return WorkSchedule{}, ErrInvalidInput
	}

	schedule, err := s.repo.FindWorkScheduleByID(ctx, id)
	if err != nil {
		return WorkSchedule{}, mapAdminRepositoryError(err)
	}

	return schedule, nil
}

func (s AdminScheduleService) UpdateWorkSchedule(ctx context.Context, id string, input WorkScheduleInput) (WorkSchedule, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return WorkSchedule{}, ErrInvalidInput
	}
	normalized, err := normalizeWorkScheduleInput(input)
	if err != nil {
		return WorkSchedule{}, err
	}

	updated, err := s.repo.UpdateWorkSchedule(ctx, WorkSchedule{
		ID:           id,
		Name:         normalized.Name,
		StartTime:    normalized.StartTime,
		EndTime:      normalized.EndTime,
		GraceMinutes: normalized.GraceMinutes,
	})
	if err != nil {
		return WorkSchedule{}, mapAdminRepositoryError(err)
	}

	return updated, nil
}

func (s AdminScheduleService) UpdateWorkScheduleStatus(ctx context.Context, id string, isActive bool) (WorkSchedule, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return WorkSchedule{}, ErrInvalidInput
	}
	if !isActive {
		used, err := s.repo.HasActiveOrFutureAssignments(ctx, id, s.businessDate())
		if err != nil {
			return WorkSchedule{}, mapAdminRepositoryError(err)
		}
		if used {
			return WorkSchedule{}, ErrScheduleInUse
		}
	}

	updated, err := s.repo.UpdateWorkScheduleStatus(ctx, id, isActive)
	if err != nil {
		return WorkSchedule{}, mapAdminRepositoryError(err)
	}

	return updated, nil
}

func (s AdminScheduleService) ListAssignments(ctx context.Context, filter AssignmentListFilter) (ScheduleAssignmentList, error) {
	normalized, err := normalizeAssignmentListFilter(filter)
	if err != nil {
		return ScheduleAssignmentList{}, err
	}
	businessDate := s.businessDate()

	total, err := s.repo.CountAssignments(ctx, normalized, businessDate)
	if err != nil {
		return ScheduleAssignmentList{}, ErrInternal
	}
	rows, err := s.repo.ListAssignments(ctx, normalized, businessDate)
	if err != nil {
		return ScheduleAssignmentList{}, ErrInternal
	}

	items := make([]ScheduleAssignment, 0, len(rows))
	for _, row := range rows {
		items = append(items, assignmentResponse(row))
	}

	return ScheduleAssignmentList{Items: items, Page: normalized.Page, PageSize: normalized.PageSize, TotalItems: total, TotalPages: totalPages(total, normalized.PageSize)}, nil
}

func (s AdminScheduleService) CreateAssignment(ctx context.Context, input AssignmentCreateInput) (ScheduleAssignment, error) {
	userID := strings.TrimSpace(input.UserID)
	scheduleID := strings.TrimSpace(input.ScheduleID)
	if !validAdminUUID(userID) || !validAdminUUID(scheduleID) {
		return ScheduleAssignment{}, ErrInvalidInput
	}
	effectiveFrom, effectiveTo, err := parseDateRange(input.EffectiveFrom, input.EffectiveTo, s.location)
	if err != nil {
		return ScheduleAssignment{}, err
	}

	u, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return ScheduleAssignment{}, mapAssignmentUserError(err)
	}
	if u.Role != user.RoleUser {
		return ScheduleAssignment{}, ErrAssignmentInvalidUser
	}
	schedule, err := s.repo.FindWorkScheduleByID(ctx, scheduleID)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}
	if !schedule.IsActive {
		return ScheduleAssignment{}, ErrInactiveSchedule
	}
	overlap, err := s.repo.HasOverlappingAssignment(ctx, userID, "", effectiveFrom, effectiveTo)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}
	if overlap {
		return ScheduleAssignment{}, ErrAssignmentOverlap
	}

	created, err := s.repo.CreateAssignment(ctx, newUUID(), userID, scheduleID, effectiveFrom, effectiveTo)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}

	return assignmentResponse(created), nil
}

func (s AdminScheduleService) AssignmentDetail(ctx context.Context, id string) (ScheduleAssignment, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return ScheduleAssignment{}, ErrInvalidInput
	}

	assignment, err := s.repo.FindAssignmentByID(ctx, id)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}

	return assignmentResponse(assignment), nil
}

func (s AdminScheduleService) EndAssignment(ctx context.Context, id string, effectiveToValue string) (ScheduleAssignment, error) {
	id = strings.TrimSpace(id)
	if !validAdminUUID(id) {
		return ScheduleAssignment{}, ErrInvalidInput
	}
	effectiveTo, err := parseRequiredDate(effectiveToValue, s.location)
	if err != nil {
		return ScheduleAssignment{}, err
	}

	current, err := s.repo.FindAssignmentByID(ctx, id)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}
	if effectiveTo.Before(current.EffectiveFrom) {
		return ScheduleAssignment{}, ErrInvalidInput
	}
	overlap, err := s.repo.HasOverlappingAssignment(ctx, current.User.ID, id, current.EffectiveFrom, &effectiveTo)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}
	if overlap {
		return ScheduleAssignment{}, ErrAssignmentOverlap
	}

	ended, err := s.repo.EndAssignment(ctx, id, effectiveTo)
	if err != nil {
		return ScheduleAssignment{}, mapAdminRepositoryError(err)
	}

	return assignmentResponse(ended), nil
}

func (s AdminScheduleService) businessDate() time.Time {
	now := s.now().In(s.location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location)
}

func normalizeWorkScheduleInput(input WorkScheduleInput) (WorkScheduleInput, error) {
	normalized := WorkScheduleInput{
		Name:         strings.Join(strings.Fields(input.Name), " "),
		StartTime:    strings.TrimSpace(input.StartTime),
		EndTime:      strings.TrimSpace(input.EndTime),
		GraceMinutes: input.GraceMinutes,
	}
	start, err := time.Parse("15:04", normalized.StartTime)
	if err != nil {
		return WorkScheduleInput{}, ErrInvalidInput
	}
	end, err := time.Parse("15:04", normalized.EndTime)
	if err != nil {
		return WorkScheduleInput{}, ErrInvalidInput
	}
	if normalized.Name == "" || !end.After(start) || normalized.GraceMinutes < 0 || normalized.GraceMinutes > maxGraceMinutes {
		return WorkScheduleInput{}, ErrInvalidInput
	}

	return normalized, nil
}

func normalizeWorkScheduleListFilter(filter WorkScheduleListFilter) (WorkScheduleListFilter, error) {
	page, pageSize := normalizeAdminPagination(filter.Page, filter.PageSize)
	status := ScheduleStatus(strings.TrimSpace(string(filter.Status)))
	if status != "" && status != ScheduleStatusActive && status != ScheduleStatusInactive {
		return WorkScheduleListFilter{}, ErrInvalidInput
	}

	return WorkScheduleListFilter{Page: page, PageSize: pageSize, Search: strings.TrimSpace(filter.Search), Status: status}, nil
}

func normalizeAssignmentListFilter(filter AssignmentListFilter) (AssignmentListFilter, error) {
	page, pageSize := normalizeAdminPagination(filter.Page, filter.PageSize)
	userID := strings.TrimSpace(filter.UserID)
	scheduleID := strings.TrimSpace(filter.ScheduleID)
	status := AssignmentStatus(strings.TrimSpace(string(filter.Status)))
	if userID != "" && !validAdminUUID(userID) {
		return AssignmentListFilter{}, ErrInvalidInput
	}
	if scheduleID != "" && !validAdminUUID(scheduleID) {
		return AssignmentListFilter{}, ErrInvalidInput
	}
	if status != "" && status != AssignmentStatusCurrent && status != AssignmentStatusUpcoming && status != AssignmentStatusEnded {
		return AssignmentListFilter{}, ErrInvalidInput
	}

	return AssignmentListFilter{Page: page, PageSize: pageSize, Search: strings.TrimSpace(filter.Search), UserID: userID, ScheduleID: scheduleID, Status: status}, nil
}

func normalizeAdminPagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultAdminPageSize
	}
	if pageSize > maxAdminPageSize {
		pageSize = maxAdminPageSize
	}

	return page, pageSize
}

func parseDateRange(from string, to *string, location *time.Location) (time.Time, *time.Time, error) {
	effectiveFrom, err := parseRequiredDate(from, location)
	if err != nil {
		return time.Time{}, nil, err
	}
	var effectiveTo *time.Time
	if to != nil {
		parsed, err := parseRequiredDate(*to, location)
		if err != nil {
			return time.Time{}, nil, err
		}
		if parsed.Before(effectiveFrom) {
			return time.Time{}, nil, ErrInvalidInput
		}
		effectiveTo = &parsed
	}

	return effectiveFrom, effectiveTo, nil
}

func parseRequiredDate(value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), location)
	if err != nil {
		return time.Time{}, ErrInvalidInput
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, location), nil
}

func assignmentResponse(row ScheduleAssignmentRecord) ScheduleAssignment {
	var effectiveTo *string
	if row.EffectiveTo != nil {
		value := formatDate(*row.EffectiveTo)
		effectiveTo = &value
	}

	return ScheduleAssignment{
		ID:            row.ID,
		User:          safeAssignmentUser(row.User),
		Schedule:      row.Schedule,
		EffectiveFrom: formatDate(row.EffectiveFrom),
		EffectiveTo:   effectiveTo,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func safeAssignmentUser(u user.User) user.EmployeeProfile {
	return user.EmployeeProfile{
		ID:             u.ID,
		EmployeeNumber: u.EmployeeNumber,
		Name:           u.Name,
		Email:          u.Email,
		Phone:          u.Phone,
		Position:       u.Position,
		Role:           u.Role,
		AccountStatus:  u.AccountStatus,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func mapAdminRepositoryError(err error) error {
	switch {
	case errors.Is(err, ErrScheduleNotFound), errors.Is(err, ErrAssignmentNotFound), errors.Is(err, ErrScheduleDuplicate), errors.Is(err, ErrScheduleInUse), errors.Is(err, ErrAssignmentOverlap), errors.Is(err, ErrInactiveSchedule):
		return err
	default:
		return ErrInternal
	}
}

func mapAssignmentUserError(err error) error {
	if errors.Is(err, user.ErrNotFound) {
		return ErrAssignmentInvalidUser
	}
	return ErrInternal
}

func validAdminUUID(id string) bool {
	id = strings.TrimSpace(id)
	if len(id) != 36 {
		return false
	}
	for i := 0; i < len(id); i++ {
		switch i {
		case 8, 13, 18, 23:
			if id[i] != '-' {
				return false
			}
		default:
			if !isAdminHex(id[i]) {
				return false
			}
		}
	}
	return true
}

func isAdminHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
