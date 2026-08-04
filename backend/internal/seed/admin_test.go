package seed

import (
	"context"
	"errors"
	"testing"

	"r3-ti-faceattend/backend/internal/user"
)

func TestAdminServiceValidatesRequiredInput(t *testing.T) {
	service := NewAdminService(newFakeUserRepository(), fakeHasher{})

	_, _, err := service.Seed(context.Background(), AdminInput{})
	if !errors.Is(err, ErrRequiredField) {
		t.Fatalf("Seed() error = %v, want %v", err, ErrRequiredField)
	}
}

func TestAdminServiceCreatesNewAdmin(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAdminService(repo, fakeHasher{})

	admin, created, err := service.Seed(context.Background(), validAdminInput())
	if err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if !created {
		t.Fatal("Seed() created = false, want true")
	}
	if admin.Role != user.RoleAdmin {
		t.Fatalf("Role = %q, want %q", admin.Role, user.RoleAdmin)
	}
	if admin.AccountStatus != user.AccountStatusActive {
		t.Fatalf("AccountStatus = %q, want %q", admin.AccountStatus, user.AccountStatusActive)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created users = %d, want 1", len(repo.created))
	}
}

func TestAdminServiceDoesNotCreateSameAdminTwice(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAdminService(repo, fakeHasher{})

	if _, created, err := service.Seed(context.Background(), validAdminInput()); err != nil || !created {
		t.Fatalf("first Seed() created = %v, error = %v", created, err)
	}
	if _, created, err := service.Seed(context.Background(), validAdminInput()); err != nil || created {
		t.Fatalf("second Seed() created = %v, error = %v, want idempotent no-op", created, err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created users = %d, want 1", len(repo.created))
	}
}

func TestAdminServiceFailsOnEmailConflict(t *testing.T) {
	repo := newFakeUserRepository()
	repo.users = append(repo.users, user.User{
		ID:             "existing-id",
		EmployeeNumber: "EMP-002",
		Name:           "Existing User",
		Email:          "admin@example.test",
		Role:           user.RoleUser,
		AccountStatus:  user.AccountStatusActive,
	})
	service := NewAdminService(repo, fakeHasher{})

	_, _, err := service.Seed(context.Background(), validAdminInput())
	if !errors.Is(err, ErrAdminEmailConflict) {
		t.Fatalf("Seed() error = %v, want %v", err, ErrAdminEmailConflict)
	}
}

func TestAdminServiceFailsOnEmployeeNumberConflict(t *testing.T) {
	repo := newFakeUserRepository()
	repo.users = append(repo.users, user.User{
		ID:             "existing-id",
		EmployeeNumber: "EMP-001",
		Name:           "Existing User",
		Email:          "different@example.test",
		Role:           user.RoleUser,
		AccountStatus:  user.AccountStatusActive,
	})
	service := NewAdminService(repo, fakeHasher{})

	_, _, err := service.Seed(context.Background(), validAdminInput())
	if !errors.Is(err, ErrAdminEmployeeConflict) {
		t.Fatalf("Seed() error = %v, want %v", err, ErrAdminEmployeeConflict)
	}
}

func TestAdminServiceDoesNotPassPlaintextPasswordAsHash(t *testing.T) {
	repo := newFakeUserRepository()
	service := NewAdminService(repo, fakeHasher{})
	input := validAdminInput()

	if _, _, err := service.Seed(context.Background(), input); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if repo.created[0].PasswordHash == input.Password {
		t.Fatal("repository received plaintext password as PasswordHash")
	}
	if repo.created[0].PasswordHash != "hashed:"+input.Password {
		t.Fatalf("PasswordHash = %q, want fake hash", repo.created[0].PasswordHash)
	}
}

func TestAdminServiceRejectsShortPassword(t *testing.T) {
	service := NewAdminService(newFakeUserRepository(), fakeHasher{})
	input := validAdminInput()
	input.Password = "short"

	_, _, err := service.Seed(context.Background(), input)
	if !errors.Is(err, ErrAdminPasswordTooShort) {
		t.Fatalf("Seed() error = %v, want %v", err, ErrAdminPasswordTooShort)
	}
}

func validAdminInput() AdminInput {
	return AdminInput{
		EmployeeNumber: "EMP-001",
		Name:           "Admin Local",
		Email:          "Admin@Example.Test",
		Password:       "admin-password",
		Position:       "Administrator",
	}
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}

	return "hashed:" + password, nil
}

func (fakeHasher) Compare(hash string, password string) bool {
	return hash == "hashed:"+password
}

type fakeUserRepository struct {
	users   []user.User
	created []user.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{}
}

func (r *fakeUserRepository) Create(ctx context.Context, u user.User) error {
	r.users = append(r.users, u)
	r.created = append(r.created, u)
	return nil
}

func (r *fakeUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	for _, u := range r.users {
		if equalFold(u.Email, email) {
			return u, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *fakeUserRepository) FindByEmployeeNumber(ctx context.Context, employeeNumber string) (user.User, error) {
	for _, u := range r.users {
		if u.EmployeeNumber == employeeNumber {
			return u, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (r *fakeUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, err := r.FindByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		return false, nil
	}

	return err == nil, err
}

func (r *fakeUserRepository) ExistsByEmployeeNumber(ctx context.Context, employeeNumber string) (bool, error) {
	_, err := r.FindByEmployeeNumber(ctx, employeeNumber)
	if errors.Is(err, user.ErrNotFound) {
		return false, nil
	}

	return err == nil, err
}

func equalFold(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if lowerASCII(a[i]) != lowerASCII(b[i]) {
			return false
		}
	}

	return true
}

func lowerASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}

	return b
}
