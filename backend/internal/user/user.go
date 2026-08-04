package user

import "time"

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "ACTIVE"
	AccountStatusInactive  AccountStatus = "INACTIVE"
	AccountStatusSuspended AccountStatus = "SUSPENDED"
)

type User struct {
	ID             string
	EmployeeNumber string
	Name           string
	Email          string
	PasswordHash   string
	Phone          *string
	Position       *string
	Role           Role
	AccountStatus  AccountStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
