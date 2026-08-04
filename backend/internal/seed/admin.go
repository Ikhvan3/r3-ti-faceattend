package seed

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"r3-ti-faceattend/backend/internal/security"
	"r3-ti-faceattend/backend/internal/user"
)

const MinAdminPasswordLength = 8

var (
	ErrRequiredField           = errors.New("required seed admin field is missing")
	ErrAdminPasswordTooShort   = errors.New("admin password must be at least 8 characters")
	ErrAdminEmailConflict      = errors.New("admin email is already used by another user")
	ErrAdminEmployeeConflict   = errors.New("admin employee number is already used by another user")
	ErrAdminSeedAlreadyApplied = errors.New("admin seed already exists")
)

type AdminInput struct {
	EmployeeNumber string
	Name           string
	Email          string
	Password       string
	Position       string
}

type AdminService struct {
	repo   user.Repository
	hasher security.PasswordHasher
}

func NewAdminService(repo user.Repository, hasher security.PasswordHasher) AdminService {
	return AdminService{repo: repo, hasher: hasher}
}

func (s AdminService) Seed(ctx context.Context, input AdminInput) (user.User, bool, error) {
	normalized, err := validateAdminInput(input)
	if err != nil {
		return user.User{}, false, err
	}

	existingByEmail, emailFound, err := s.findByEmail(ctx, normalized.Email)
	if err != nil {
		return user.User{}, false, err
	}

	existingByEmployee, employeeFound, err := s.findByEmployeeNumber(ctx, normalized.EmployeeNumber)
	if err != nil {
		return user.User{}, false, err
	}

	switch {
	case emailFound && !sameAdmin(existingByEmail, normalized):
		return user.User{}, false, ErrAdminEmailConflict
	case employeeFound && !sameAdmin(existingByEmployee, normalized):
		return user.User{}, false, ErrAdminEmployeeConflict
	case emailFound && employeeFound:
		existingByEmail.PasswordHash = ""
		return existingByEmail, false, nil
	case emailFound:
		existingByEmail.PasswordHash = ""
		return existingByEmail, false, nil
	case employeeFound:
		existingByEmployee.PasswordHash = ""
		return existingByEmployee, false, nil
	}

	passwordHash, err := s.hasher.Hash(normalized.Password)
	if err != nil {
		return user.User{}, false, err
	}

	position := normalized.Position
	admin := user.User{
		ID:             newUUID(),
		EmployeeNumber: normalized.EmployeeNumber,
		Name:           normalized.Name,
		Email:          normalized.Email,
		PasswordHash:   passwordHash,
		Phone:          nil,
		Position:       &position,
		Role:           user.RoleAdmin,
		AccountStatus:  user.AccountStatusActive,
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return user.User{}, false, err
	}

	admin.PasswordHash = ""
	return admin, true, nil
}

func (s AdminService) findByEmail(ctx context.Context, email string) (user.User, bool, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		return user.User{}, false, nil
	}
	if err != nil {
		return user.User{}, false, err
	}

	return u, true, nil
}

func (s AdminService) findByEmployeeNumber(ctx context.Context, employeeNumber string) (user.User, bool, error) {
	u, err := s.repo.FindByEmployeeNumber(ctx, employeeNumber)
	if errors.Is(err, user.ErrNotFound) {
		return user.User{}, false, nil
	}
	if err != nil {
		return user.User{}, false, err
	}

	return u, true, nil
}

func validateAdminInput(input AdminInput) (AdminInput, error) {
	normalized := AdminInput{
		EmployeeNumber: strings.TrimSpace(input.EmployeeNumber),
		Name:           strings.TrimSpace(input.Name),
		Email:          strings.ToLower(strings.TrimSpace(input.Email)),
		Password:       input.Password,
		Position:       strings.TrimSpace(input.Position),
	}

	if normalized.EmployeeNumber == "" ||
		normalized.Name == "" ||
		normalized.Email == "" ||
		normalized.Password == "" ||
		normalized.Position == "" {
		return AdminInput{}, ErrRequiredField
	}
	if len(normalized.Password) < MinAdminPasswordLength {
		return AdminInput{}, ErrAdminPasswordTooShort
	}

	return normalized, nil
}

func sameAdmin(existing user.User, input AdminInput) bool {
	return existing.EmployeeNumber == input.EmployeeNumber &&
		strings.EqualFold(existing.Email, input.Email) &&
		existing.Role == user.RoleAdmin
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
