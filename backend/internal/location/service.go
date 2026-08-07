package location

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
	minRadiusMeters = 10
	maxRadiusMeters = 2000
)

type OfficeLocationRepository interface {
	ListOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) ([]OfficeLocation, error)
	CountOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) (int, error)
	CreateOfficeLocation(ctx context.Context, office OfficeLocation) (OfficeLocation, error)
	FindOfficeLocationByID(ctx context.Context, id string) (OfficeLocation, error)
	UpdateOfficeLocation(ctx context.Context, office OfficeLocation) (OfficeLocation, error)
	UpdateOfficeLocationStatus(ctx context.Context, id string, isActive bool) (OfficeLocation, error)
	HasActiveOrFutureLocationAssignments(ctx context.Context, officeID string, businessDate time.Time) (bool, error)
}

type LocationAssignmentRepository interface {
	ListLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter, businessDate time.Time) ([]LocationAssignmentRecord, error)
	CountLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter, businessDate time.Time) (int, error)
	FindLocationAssignmentByID(ctx context.Context, id string) (LocationAssignmentRecord, error)
	FindUserByID(ctx context.Context, id string) (user.User, error)
	HasOverlappingLocationAssignment(ctx context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error)
	CreateLocationAssignment(ctx context.Context, assignmentID string, userID string, officeID string, effectiveFrom time.Time, effectiveTo *time.Time) (LocationAssignmentRecord, error)
	EndLocationAssignment(ctx context.Context, id string, effectiveTo time.Time) (LocationAssignmentRecord, error)
	FindCurrentLocationAssignment(ctx context.Context, userID string, businessDate time.Time) (LocationAssignmentRecord, error)
}

type Service struct {
	offices     OfficeLocationRepository
	assignments LocationAssignmentRepository
	location    *time.Location
	now         func() time.Time
}

func NewService(offices OfficeLocationRepository, assignments LocationAssignmentRepository, location *time.Location) Service {
	return Service{offices: offices, assignments: assignments, location: location, now: time.Now}
}

func (s Service) ListOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) (OfficeLocationList, error) {
	normalized, err := normalizeOfficeFilter(filter)
	if err != nil {
		return OfficeLocationList{}, err
	}
	total, err := s.offices.CountOfficeLocations(ctx, normalized)
	if err != nil {
		return OfficeLocationList{}, ErrInternal
	}
	items, err := s.offices.ListOfficeLocations(ctx, normalized)
	if err != nil {
		return OfficeLocationList{}, ErrInternal
	}
	if items == nil {
		items = []OfficeLocation{}
	}
	return OfficeLocationList{Items: items, Page: normalized.Page, PageSize: normalized.PageSize, TotalItems: total, TotalPages: totalPages(total, normalized.PageSize)}, nil
}

func (s Service) CreateOfficeLocation(ctx context.Context, input OfficeLocationInput) (OfficeLocation, error) {
	normalized, err := normalizeOfficeInput(input)
	if err != nil {
		return OfficeLocation{}, err
	}
	office, err := s.offices.CreateOfficeLocation(ctx, OfficeLocation{
		ID:           newUUID(),
		Name:         normalized.Name,
		Address:      normalized.Address,
		Latitude:     normalized.Latitude,
		Longitude:    normalized.Longitude,
		RadiusMeters: normalized.RadiusMeters,
		IsActive:     true,
	})
	if err != nil {
		return OfficeLocation{}, mapRepositoryError(err)
	}
	return office, nil
}

func (s Service) OfficeLocationDetail(ctx context.Context, id string) (OfficeLocation, error) {
	id = strings.TrimSpace(id)
	if !validUUID(id) {
		return OfficeLocation{}, ErrInvalidInput
	}
	office, err := s.offices.FindOfficeLocationByID(ctx, id)
	if err != nil {
		return OfficeLocation{}, mapRepositoryError(err)
	}
	return office, nil
}

func (s Service) UpdateOfficeLocation(ctx context.Context, id string, input OfficeLocationInput) (OfficeLocation, error) {
	id = strings.TrimSpace(id)
	if !validUUID(id) {
		return OfficeLocation{}, ErrInvalidInput
	}
	normalized, err := normalizeOfficeInput(input)
	if err != nil {
		return OfficeLocation{}, err
	}
	office, err := s.offices.UpdateOfficeLocation(ctx, OfficeLocation{
		ID:           id,
		Name:         normalized.Name,
		Address:      normalized.Address,
		Latitude:     normalized.Latitude,
		Longitude:    normalized.Longitude,
		RadiusMeters: normalized.RadiusMeters,
	})
	if err != nil {
		return OfficeLocation{}, mapRepositoryError(err)
	}
	return office, nil
}

func (s Service) UpdateOfficeLocationStatus(ctx context.Context, id string, isActive bool) (OfficeLocation, error) {
	id = strings.TrimSpace(id)
	if !validUUID(id) {
		return OfficeLocation{}, ErrInvalidInput
	}
	if !isActive {
		used, err := s.offices.HasActiveOrFutureLocationAssignments(ctx, id, s.businessDate())
		if err != nil {
			return OfficeLocation{}, mapRepositoryError(err)
		}
		if used {
			return OfficeLocation{}, ErrOfficeInUse
		}
	}
	office, err := s.offices.UpdateOfficeLocationStatus(ctx, id, isActive)
	if err != nil {
		return OfficeLocation{}, mapRepositoryError(err)
	}
	return office, nil
}

func (s Service) ListLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter) (LocationAssignmentList, error) {
	normalized, err := normalizeAssignmentFilter(filter)
	if err != nil {
		return LocationAssignmentList{}, err
	}
	businessDate := s.businessDate()
	total, err := s.assignments.CountLocationAssignments(ctx, normalized, businessDate)
	if err != nil {
		return LocationAssignmentList{}, ErrInternal
	}
	rows, err := s.assignments.ListLocationAssignments(ctx, normalized, businessDate)
	if err != nil {
		return LocationAssignmentList{}, ErrInternal
	}
	items := make([]LocationAssignment, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.assignmentResponse(row))
	}
	return LocationAssignmentList{Items: items, Page: normalized.Page, PageSize: normalized.PageSize, TotalItems: total, TotalPages: totalPages(total, normalized.PageSize)}, nil
}

func (s Service) CreateLocationAssignment(ctx context.Context, input LocationAssignmentInput) (LocationAssignment, error) {
	userID := strings.TrimSpace(input.UserID)
	officeID := strings.TrimSpace(input.OfficeLocationID)
	if !validUUID(userID) || !validUUID(officeID) {
		return LocationAssignment{}, ErrInvalidInput
	}
	effectiveFrom, effectiveTo, err := parseDateRange(input.EffectiveFrom, input.EffectiveTo, s.location)
	if err != nil {
		return LocationAssignment{}, err
	}
	u, err := s.assignments.FindUserByID(ctx, userID)
	if err != nil {
		return LocationAssignment{}, mapUserError(err)
	}
	if u.Role != user.RoleUser {
		return LocationAssignment{}, ErrInvalidUser
	}
	office, err := s.offices.FindOfficeLocationByID(ctx, officeID)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	if !office.IsActive {
		return LocationAssignment{}, ErrInactiveOffice
	}
	overlap, err := s.assignments.HasOverlappingLocationAssignment(ctx, userID, "", effectiveFrom, effectiveTo)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	if overlap {
		return LocationAssignment{}, ErrAssignmentOverlap
	}
	created, err := s.assignments.CreateLocationAssignment(ctx, newUUID(), userID, officeID, effectiveFrom, effectiveTo)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	return s.assignmentResponse(created), nil
}

func (s Service) LocationAssignmentDetail(ctx context.Context, id string) (LocationAssignment, error) {
	id = strings.TrimSpace(id)
	if !validUUID(id) {
		return LocationAssignment{}, ErrInvalidInput
	}
	assignment, err := s.assignments.FindLocationAssignmentByID(ctx, id)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	return s.assignmentResponse(assignment), nil
}

func (s Service) EndLocationAssignment(ctx context.Context, id string, effectiveToValue string) (LocationAssignment, error) {
	id = strings.TrimSpace(id)
	if !validUUID(id) {
		return LocationAssignment{}, ErrInvalidInput
	}
	effectiveTo, err := parseRequiredDate(effectiveToValue, s.location)
	if err != nil {
		return LocationAssignment{}, err
	}
	current, err := s.assignments.FindLocationAssignmentByID(ctx, id)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	if effectiveTo.Before(current.EffectiveFrom) {
		return LocationAssignment{}, ErrInvalidInput
	}
	overlap, err := s.assignments.HasOverlappingLocationAssignment(ctx, current.User.ID, id, current.EffectiveFrom, &effectiveTo)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	if overlap {
		return LocationAssignment{}, ErrAssignmentOverlap
	}
	ended, err := s.assignments.EndLocationAssignment(ctx, id, effectiveTo)
	if err != nil {
		return LocationAssignment{}, mapRepositoryError(err)
	}
	return s.assignmentResponse(ended), nil
}

func (s Service) LocationRequirement(ctx context.Context, claims auth.Claims) (LocationRequirement, error) {
	if claims.Subject == "" || claims.Role != user.RoleUser {
		return LocationRequirement{}, ErrInvalidInput
	}
	row, err := s.assignments.FindCurrentLocationAssignment(ctx, claims.Subject, s.businessDate())
	if err != nil {
		return LocationRequirement{}, mapRepositoryError(err)
	}
	if !row.Office.IsActive {
		return LocationRequirement{}, ErrOfficeNotFound
	}
	assignment := s.assignmentResponse(row)
	return LocationRequirement{Assignment: assignment, Office: row.Office}, nil
}

func normalizeOfficeInput(input OfficeLocationInput) (OfficeLocationInput, error) {
	name := strings.Join(strings.Fields(input.Name), " ")
	if name == "" || !ValidLatitude(input.Latitude) || !ValidLongitude(input.Longitude) || input.RadiusMeters < minRadiusMeters || input.RadiusMeters > maxRadiusMeters {
		return OfficeLocationInput{}, ErrInvalidInput
	}
	address := normalizeOptionalString(input.Address)
	return OfficeLocationInput{Name: name, Address: address, Latitude: input.Latitude, Longitude: input.Longitude, RadiusMeters: input.RadiusMeters}, nil
}

func normalizeOfficeFilter(filter OfficeLocationListFilter) (OfficeLocationListFilter, error) {
	page, pageSize := normalizePagination(filter.Page, filter.PageSize)
	status := OfficeStatus(strings.TrimSpace(string(filter.Status)))
	if status != "" && status != OfficeStatusActive && status != OfficeStatusInactive {
		return OfficeLocationListFilter{}, ErrInvalidInput
	}
	return OfficeLocationListFilter{Page: page, PageSize: pageSize, Search: strings.TrimSpace(filter.Search), Status: status}, nil
}

func normalizeAssignmentFilter(filter LocationAssignmentListFilter) (LocationAssignmentListFilter, error) {
	page, pageSize := normalizePagination(filter.Page, filter.PageSize)
	userID := strings.TrimSpace(filter.UserID)
	officeID := strings.TrimSpace(filter.OfficeLocationID)
	status := AssignmentStatus(strings.TrimSpace(string(filter.Status)))
	if userID != "" && !validUUID(userID) {
		return LocationAssignmentListFilter{}, ErrInvalidInput
	}
	if officeID != "" && !validUUID(officeID) {
		return LocationAssignmentListFilter{}, ErrInvalidInput
	}
	if status != "" && status != AssignmentStatusCurrent && status != AssignmentStatusUpcoming && status != AssignmentStatusEnded {
		return LocationAssignmentListFilter{}, ErrInvalidInput
	}
	return LocationAssignmentListFilter{Page: page, PageSize: pageSize, Search: strings.TrimSpace(filter.Search), UserID: userID, OfficeLocationID: officeID, Status: status}, nil
}

func normalizePagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
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

func (s Service) assignmentResponse(row LocationAssignmentRecord) LocationAssignment {
	var effectiveTo *string
	if row.EffectiveTo != nil {
		value := formatDate(*row.EffectiveTo)
		effectiveTo = &value
	}
	return LocationAssignment{
		ID:            row.ID,
		User:          safeEmployee(row.User),
		Office:        row.Office,
		EffectiveFrom: formatDate(row.EffectiveFrom),
		EffectiveTo:   effectiveTo,
		Status:        assignmentStatus(row, s.businessDate(), s.location),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func assignmentStatus(row LocationAssignmentRecord, businessDate time.Time, location *time.Location) AssignmentStatus {
	businessDay := normalizeDateOnly(businessDate, location)
	effectiveFrom := normalizeDateOnly(row.EffectiveFrom, location)
	if businessDay.Before(effectiveFrom) {
		return AssignmentStatusUpcoming
	}
	if row.EffectiveTo != nil {
		effectiveTo := normalizeDateOnly(*row.EffectiveTo, location)
		if businessDay.After(effectiveTo) {
			return AssignmentStatusEnded
		}
	}
	return AssignmentStatusCurrent
}

func safeEmployee(u user.User) user.EmployeeProfile {
	return user.EmployeeProfile{ID: u.ID, EmployeeNumber: u.EmployeeNumber, Name: u.Name, Email: u.Email, Phone: u.Phone, Position: u.Position, Role: u.Role, AccountStatus: u.AccountStatus, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

func (s Service) businessDate() time.Time {
	now := s.now().In(s.location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.location)
}

func normalizeDateOnly(value time.Time, location *time.Location) time.Time {
	inLocation := value.In(location)
	return time.Date(inLocation.Year(), inLocation.Month(), inLocation.Day(), 0, 0, 0, 0, location)
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

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func mapRepositoryError(err error) error {
	switch err {
	case ErrOfficeNotFound, ErrOfficeInUse, ErrInactiveOffice, ErrAssignmentNotFound, ErrAssignmentOverlap, ErrInvalidUser, ErrInvalidInput:
		return err
	default:
		return ErrInternal
	}
}

func mapUserError(err error) error {
	if err == user.ErrNotFound {
		return ErrInvalidUser
	}
	return ErrInternal
}

func validUUID(id string) bool {
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
			if !((id[i] >= '0' && id[i] <= '9') || (id[i] >= 'a' && id[i] <= 'f') || (id[i] >= 'A' && id[i] <= 'F')) {
				return false
			}
		}
	}
	return true
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("generate uuid: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
