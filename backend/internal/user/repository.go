package user

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByEmployeeNumber(ctx context.Context, employeeNumber string) (User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByEmployeeNumber(ctx context.Context, employeeNumber string) (bool, error)
}
