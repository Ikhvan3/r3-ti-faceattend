package user

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEmployeeServiceCreateSuccessForcesUserActiveAndHashesPassword(t *testing.T) {
	repo := newEmployeeFakeRepository()
	service := NewEmployeeService(repo, employeeFakeHasher{})

	profile, err := service.Create(context.Background(), validEmployeeCreateInput())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if profile.Role != RoleUser {
		t.Fatalf("Role = %q, want %q", profile.Role, RoleUser)
	}
	if profile.AccountStatus != AccountStatusActive {
		t.Fatalf("AccountStatus = %q, want %q", profile.AccountStatus, AccountStatusActive)
	}
	if len(repo.users) != 1 {
		t.Fatalf("users = %d, want 1", len(repo.users))
	}
	if repo.users[0].PasswordHash == validEmployeeCreateInput().InitialPassword {
		t.Fatal("repository received plaintext password")
	}
	if repo.users[0].PasswordHash != "hashed:"+validEmployeeCreateInput().InitialPassword {
		t.Fatalf("PasswordHash = %q", repo.users[0].PasswordHash)
	}
}

func TestEmployeeServiceCreateDuplicateEmailAndEmployeeNumber(t *testing.T) {
	tests := []struct {
		name string
		user User
		want error
	}{
		{name: "email", user: employeeUser("00000000-0000-4000-8000-000000000001", "EMP-999", "dummy.ti@example.test"), want: ErrEmailConflict},
		{name: "employee number", user: employeeUser("00000000-0000-4000-8000-000000000002", "EMP-001", "other.ti@example.test"), want: ErrEmployeeNumberConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newEmployeeFakeRepository()
			repo.users = append(repo.users, tt.user)
			service := NewEmployeeService(repo, employeeFakeHasher{})

			_, err := service.Create(context.Background(), validEmployeeCreateInput())
			if !errors.Is(err, tt.want) {
				t.Fatalf("Create() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestEmployeeServiceCreateRejectsEmptyInput(t *testing.T) {
	service := NewEmployeeService(newEmployeeFakeRepository(), employeeFakeHasher{})

	_, err := service.Create(context.Background(), EmployeeCreateInput{})
	if !errors.Is(err, ErrInvalidEmployeeInput) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidEmployeeInput)
	}
}

func TestEmployeeServiceListPagination(t *testing.T) {
	repo := newEmployeeFakeRepository()
	repo.users = append(repo.users,
		employeeUser("00000000-0000-4000-8000-000000000001", "EMP-001", "one.ti@example.test"),
		employeeUser("00000000-0000-4000-8000-000000000002", "EMP-002", "two.ti@example.test"),
		employeeUser("00000000-0000-4000-8000-000000000003", "EMP-003", "three.ti@example.test"),
	)
	service := NewEmployeeService(repo, employeeFakeHasher{})

	result, err := service.List(context.Background(), EmployeeListFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	if result.TotalItems != 3 || result.TotalPages != 2 {
		t.Fatalf("pagination = %+v", result)
	}
}

func TestEmployeeServiceListEmptyReturnsEmptyItems(t *testing.T) {
	service := NewEmployeeService(newEmployeeFakeRepository(), employeeFakeHasher{})

	result, err := service.List(context.Background(), EmployeeListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Items == nil {
		t.Fatal("Items = nil, want empty slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("Items length = %d, want 0", len(result.Items))
	}
	if result.TotalItems != 0 || result.TotalPages != 0 {
		t.Fatalf("pagination = %+v, want zero totals", result)
	}
}

func TestEmployeeServiceUpdateSuccessAndRejectsAdmin(t *testing.T) {
	repo := newEmployeeFakeRepository()
	employee := employeeUser("00000000-0000-4000-8000-000000000001", "EMP-001", "old.ti@example.test")
	admin := employeeUser("00000000-0000-4000-8000-000000000002", "ADMIN-LOCAL", "admin.ti@example.test")
	admin.Role = RoleAdmin
	repo.users = append(repo.users, employee, admin)
	service := NewEmployeeService(repo, employeeFakeHasher{})

	updated, err := service.Update(context.Background(), employee.ID, EmployeeUpdateInput{
		EmployeeNumber: "EMP-001",
		Name:           "Pegawai Dummy Updated",
		Email:          "Updated.TI@Example.Test",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Email != "updated.ti@example.test" || updated.Name != "Pegawai Dummy Updated" {
		t.Fatalf("updated profile = %+v", updated)
	}

	_, err = service.Update(context.Background(), admin.ID, EmployeeUpdateInput{
		EmployeeNumber: "ADMIN-LOCAL",
		Name:           "Admin Lokal",
		Email:          "admin.ti@example.test",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("admin Update() error = %v, want %v", err, ErrNotFound)
	}
}

func TestEmployeeServiceUpdateStatusValidAndInvalid(t *testing.T) {
	repo := newEmployeeFakeRepository()
	employee := employeeUser("00000000-0000-4000-8000-000000000001", "EMP-001", "one.ti@example.test")
	repo.users = append(repo.users, employee)
	service := NewEmployeeService(repo, employeeFakeHasher{})

	updated, err := service.UpdateStatus(context.Background(), employee.ID, AccountStatusInactive)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if updated.AccountStatus != AccountStatusInactive {
		t.Fatalf("AccountStatus = %q, want %q", updated.AccountStatus, AccountStatusInactive)
	}

	_, err = service.UpdateStatus(context.Background(), employee.ID, AccountStatus("LOCKED"))
	if !errors.Is(err, ErrInvalidEmployeeStatus) {
		t.Fatalf("invalid UpdateStatus() error = %v, want %v", err, ErrInvalidEmployeeStatus)
	}
}

func validEmployeeCreateInput() EmployeeCreateInput {
	phone := "081234567890"
	position := "Staf TI"
	return EmployeeCreateInput{
		EmployeeNumber:  " emp-001 ",
		Name:            " Pegawai Dummy ",
		Email:           "Dummy.TI@Example.Test",
		InitialPassword: "password-dummy",
		Phone:           &phone,
		Position:        &position,
	}
}

type employeeFakeHasher struct{}

func (employeeFakeHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}
	return "hashed:" + password, nil
}

func (employeeFakeHasher) Compare(hash string, password string) bool {
	return hash == "hashed:"+password
}

type employeeFakeRepository struct {
	users []User
}

func newEmployeeFakeRepository() *employeeFakeRepository {
	return &employeeFakeRepository{}
}

func (r *employeeFakeRepository) ListEmployees(ctx context.Context, filter EmployeeListFilter) ([]User, error) {
	var filtered []User
	for _, u := range r.users {
		if u.Role != RoleUser {
			continue
		}
		if filter.Status != "" && u.AccountStatus != filter.Status {
			continue
		}
		if filter.Search != "" && !employeeMatchesSearch(u, filter.Search) {
			continue
		}
		filtered = append(filtered, u)
	}

	start := (filter.Page - 1) * filter.PageSize
	if start >= len(filtered) {
		return []User{}, nil
	}
	end := start + filter.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

func (r *employeeFakeRepository) CountEmployees(ctx context.Context, filter EmployeeListFilter) (int, error) {
	users, err := r.ListEmployees(ctx, EmployeeListFilter{
		Page:     1,
		PageSize: len(r.users) + 1,
		Search:   filter.Search,
		Status:   filter.Status,
	})
	if err != nil {
		return 0, err
	}

	return len(users), nil
}

func (r *employeeFakeRepository) FindEmployeeByID(ctx context.Context, id string) (User, error) {
	for _, u := range r.users {
		if u.ID == id && u.Role == RoleUser {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (r *employeeFakeRepository) CreateEmployee(ctx context.Context, u User) (User, error) {
	for _, existing := range r.users {
		if employeeEqualFold(existing.Email, u.Email) {
			return User{}, ErrEmailConflict
		}
		if existing.EmployeeNumber == u.EmployeeNumber {
			return User{}, ErrEmployeeNumberConflict
		}
	}

	u.Role = RoleUser
	u.AccountStatus = AccountStatusActive
	if u.CreatedAt.IsZero() {
		u.CreatedAt = employeeTestTime()
		u.UpdatedAt = employeeTestTime()
	}
	r.users = append(r.users, u)
	return u, nil
}

func (r *employeeFakeRepository) UpdateEmployee(ctx context.Context, input User) (User, error) {
	for _, existing := range r.users {
		if existing.ID != input.ID && employeeEqualFold(existing.Email, input.Email) {
			return User{}, ErrEmailConflict
		}
		if existing.ID != input.ID && existing.EmployeeNumber == input.EmployeeNumber {
			return User{}, ErrEmployeeNumberConflict
		}
	}

	for i, existing := range r.users {
		if existing.ID == input.ID && existing.Role == RoleUser {
			existing.EmployeeNumber = input.EmployeeNumber
			existing.Name = input.Name
			existing.Email = input.Email
			existing.Phone = input.Phone
			existing.Position = input.Position
			existing.UpdatedAt = employeeTestTime().Add(time.Minute)
			r.users[i] = existing
			return existing, nil
		}
	}

	return User{}, ErrNotFound
}

func (r *employeeFakeRepository) UpdateEmployeeStatus(ctx context.Context, id string, status AccountStatus) (User, error) {
	for i, existing := range r.users {
		if existing.ID == id && existing.Role == RoleUser {
			existing.AccountStatus = status
			existing.UpdatedAt = employeeTestTime().Add(time.Minute)
			r.users[i] = existing
			return existing, nil
		}
	}

	return User{}, ErrNotFound
}

func employeeUser(id string, employeeNumber string, email string) User {
	return User{
		ID:             id,
		EmployeeNumber: employeeNumber,
		Name:           "Pegawai Dummy",
		Email:          email,
		PasswordHash:   "hashed:password-dummy",
		Role:           RoleUser,
		AccountStatus:  AccountStatusActive,
		CreatedAt:      employeeTestTime(),
		UpdatedAt:      employeeTestTime(),
	}
}

func employeeTestTime() time.Time {
	return time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
}

func employeeMatchesSearch(u User, search string) bool {
	search = employeeLower(search)
	return stringsContains(employeeLower(u.EmployeeNumber), search) ||
		stringsContains(employeeLower(u.Name), search) ||
		stringsContains(employeeLower(u.Email), search) ||
		(u.Position != nil && stringsContains(employeeLower(*u.Position), search))
}

func employeeEqualFold(a string, b string) bool {
	return employeeLower(a) == employeeLower(b)
}

func employeeLower(value string) string {
	var b []byte
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= 'A' && ch <= 'Z' {
			ch += 'a' - 'A'
		}
		b = append(b, ch)
	}
	return string(b)
}

func stringsContains(value string, substr string) bool {
	if substr == "" {
		return true
	}
	if len(substr) > len(value) {
		return false
	}
	for i := 0; i <= len(value)-len(substr); i++ {
		if value[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
