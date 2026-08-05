package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"r3-ti-faceattend/backend/internal/security"
)

const (
	defaultEmployeePageSize = 10
	maxEmployeePageSize     = 100
	minEmployeePasswordLen  = 8
)

var (
	ErrInvalidEmployeeInput  = errors.New("employee input is invalid")
	ErrEmployeePasswordShort = errors.New("employee password must be at least 8 characters")
	ErrInvalidEmployeeStatus = errors.New("employee account status is invalid")
	ErrEmployeeInternal      = errors.New("employee operation failed")
)

type EmployeeRepository interface {
	ListEmployees(ctx context.Context, filter EmployeeListFilter) ([]User, error)
	CountEmployees(ctx context.Context, filter EmployeeListFilter) (int, error)
	FindEmployeeByID(ctx context.Context, id string) (User, error)
	CreateEmployee(ctx context.Context, u User) (User, error)
	UpdateEmployee(ctx context.Context, u User) (User, error)
	UpdateEmployeeStatus(ctx context.Context, id string, status AccountStatus) (User, error)
}

type EmployeeCreateInput struct {
	EmployeeNumber  string
	Name            string
	Email           string
	InitialPassword string
	Phone           *string
	Position        *string
}

type EmployeeUpdateInput struct {
	EmployeeNumber string
	Name           string
	Email          string
	Phone          *string
	Position       *string
}

type EmployeeService struct {
	repo   EmployeeRepository
	hasher security.PasswordHasher
}

func NewEmployeeService(repo EmployeeRepository, hasher security.PasswordHasher) EmployeeService {
	return EmployeeService{repo: repo, hasher: hasher}
}

func (s EmployeeService) List(ctx context.Context, filter EmployeeListFilter) (EmployeeList, error) {
	normalized, err := normalizeEmployeeListFilter(filter)
	if err != nil {
		return EmployeeList{}, err
	}

	total, err := s.repo.CountEmployees(ctx, normalized)
	if err != nil {
		return EmployeeList{}, ErrEmployeeInternal
	}

	users, err := s.repo.ListEmployees(ctx, normalized)
	if err != nil {
		return EmployeeList{}, ErrEmployeeInternal
	}

	items := make([]EmployeeProfile, 0, len(users))
	for _, u := range users {
		items = append(items, safeEmployee(u))
	}

	return EmployeeList{
		Items:      items,
		Page:       normalized.Page,
		PageSize:   normalized.PageSize,
		TotalItems: total,
		TotalPages: totalPages(total, normalized.PageSize),
	}, nil
}

func (s EmployeeService) Create(ctx context.Context, input EmployeeCreateInput) (EmployeeProfile, error) {
	normalized, err := normalizeEmployeeCreateInput(input)
	if err != nil {
		return EmployeeProfile{}, err
	}

	passwordHash, err := s.hasher.Hash(normalized.InitialPassword)
	if err != nil {
		return EmployeeProfile{}, ErrEmployeeInternal
	}

	created, err := s.repo.CreateEmployee(ctx, User{
		ID:             newEmployeeUUID(),
		EmployeeNumber: normalized.EmployeeNumber,
		Name:           normalized.Name,
		Email:          normalized.Email,
		PasswordHash:   passwordHash,
		Phone:          normalized.Phone,
		Position:       normalized.Position,
		Role:           RoleUser,
		AccountStatus:  AccountStatusActive,
	})
	if err != nil {
		return EmployeeProfile{}, mapEmployeeRepositoryError(err)
	}

	return safeEmployee(created), nil
}

func (s EmployeeService) Detail(ctx context.Context, id string) (EmployeeProfile, error) {
	if !validUUID(id) {
		return EmployeeProfile{}, ErrInvalidEmployeeInput
	}

	u, err := s.repo.FindEmployeeByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return EmployeeProfile{}, mapEmployeeRepositoryError(err)
	}

	return safeEmployee(u), nil
}

func (s EmployeeService) Update(ctx context.Context, id string, input EmployeeUpdateInput) (EmployeeProfile, error) {
	if !validUUID(id) {
		return EmployeeProfile{}, ErrInvalidEmployeeInput
	}

	normalized, err := normalizeEmployeeUpdateInput(input)
	if err != nil {
		return EmployeeProfile{}, err
	}

	updated, err := s.repo.UpdateEmployee(ctx, User{
		ID:             strings.TrimSpace(id),
		EmployeeNumber: normalized.EmployeeNumber,
		Name:           normalized.Name,
		Email:          normalized.Email,
		Phone:          normalized.Phone,
		Position:       normalized.Position,
		Role:           RoleUser,
	})
	if err != nil {
		return EmployeeProfile{}, mapEmployeeRepositoryError(err)
	}

	return safeEmployee(updated), nil
}

func (s EmployeeService) UpdateStatus(ctx context.Context, id string, status AccountStatus) (EmployeeProfile, error) {
	if !validUUID(id) {
		return EmployeeProfile{}, ErrInvalidEmployeeInput
	}
	if !validAccountStatus(status) {
		return EmployeeProfile{}, ErrInvalidEmployeeStatus
	}

	updated, err := s.repo.UpdateEmployeeStatus(ctx, strings.TrimSpace(id), status)
	if err != nil {
		return EmployeeProfile{}, mapEmployeeRepositoryError(err)
	}

	return safeEmployee(updated), nil
}

func normalizeEmployeeListFilter(filter EmployeeListFilter) (EmployeeListFilter, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = defaultEmployeePageSize
	}
	if pageSize > maxEmployeePageSize {
		pageSize = maxEmployeePageSize
	}

	status := AccountStatus(strings.TrimSpace(string(filter.Status)))
	if status != "" && !validAccountStatus(status) {
		return EmployeeListFilter{}, ErrInvalidEmployeeStatus
	}

	return EmployeeListFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   strings.TrimSpace(filter.Search),
		Status:   status,
	}, nil
}

func normalizeEmployeeCreateInput(input EmployeeCreateInput) (EmployeeCreateInput, error) {
	normalized := EmployeeCreateInput{
		EmployeeNumber:  normalizeEmployeeNumber(input.EmployeeNumber),
		Name:            strings.TrimSpace(input.Name),
		Email:           normalizeEmail(input.Email),
		InitialPassword: input.InitialPassword,
		Phone:           normalizeOptionalString(input.Phone),
		Position:        normalizeOptionalString(input.Position),
	}

	if normalized.EmployeeNumber == "" || normalized.Name == "" || normalized.Email == "" || normalized.InitialPassword == "" {
		return EmployeeCreateInput{}, ErrInvalidEmployeeInput
	}
	if len(normalized.InitialPassword) < minEmployeePasswordLen {
		return EmployeeCreateInput{}, ErrEmployeePasswordShort
	}

	return normalized, nil
}

func normalizeEmployeeUpdateInput(input EmployeeUpdateInput) (EmployeeUpdateInput, error) {
	normalized := EmployeeUpdateInput{
		EmployeeNumber: normalizeEmployeeNumber(input.EmployeeNumber),
		Name:           strings.TrimSpace(input.Name),
		Email:          normalizeEmail(input.Email),
		Phone:          normalizeOptionalString(input.Phone),
		Position:       normalizeOptionalString(input.Position),
	}

	if normalized.EmployeeNumber == "" || normalized.Name == "" || normalized.Email == "" {
		return EmployeeUpdateInput{}, ErrInvalidEmployeeInput
	}

	return normalized, nil
}

func normalizeEmployeeNumber(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
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

func validAccountStatus(status AccountStatus) bool {
	switch status {
	case AccountStatusActive, AccountStatusInactive, AccountStatusSuspended:
		return true
	default:
		return false
	}
}

func totalPages(totalItems int, pageSize int) int {
	if totalItems == 0 {
		return 0
	}

	return (totalItems + pageSize - 1) / pageSize
}

func mapEmployeeRepositoryError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrEmailConflict), errors.Is(err, ErrEmployeeNumberConflict):
		return err
	default:
		return ErrEmployeeInternal
	}
}

func newEmployeeUUID() string {
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
			if !isHex(id[i]) {
				return false
			}
		}
	}

	return true
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'a' && b <= 'f') ||
		(b >= 'A' && b <= 'F')
}
