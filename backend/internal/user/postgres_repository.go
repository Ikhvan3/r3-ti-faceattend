package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, u User) error {
	const query = `
		INSERT INTO users (
			id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err := r.pool.Exec(ctx, query,
		u.ID,
		u.EmployeeNumber,
		u.Name,
		normalizeEmail(u.Email),
		u.PasswordHash,
		u.Phone,
		u.Position,
		u.Role,
		u.AccountStatus,
	)
	if err != nil {
		return sanitizePostgresError(err)
	}

	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	return r.findOne(ctx, query, strings.TrimSpace(id))
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`

	return r.findOne(ctx, query, normalizeEmail(email))
}

func (r *PostgresRepository) FindByEmployeeNumber(ctx context.Context, employeeNumber string) (User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE employee_number = $1
	`

	return r.findOne(ctx, query, strings.TrimSpace(employeeNumber))
}

func (r *PostgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`

	return r.exists(ctx, query, normalizeEmail(email))
}

func (r *PostgresRepository) ExistsByEmployeeNumber(ctx context.Context, employeeNumber string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE employee_number = $1)`

	return r.exists(ctx, query, strings.TrimSpace(employeeNumber))
}

func (r *PostgresRepository) findOne(ctx context.Context, query string, arg string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&u.ID,
		&u.EmployeeNumber,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Phone,
		&u.Position,
		&u.Role,
		&u.AccountStatus,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, sanitizePostgresError(err)
	}

	return u, nil
}

func (r *PostgresRepository) exists(ctx context.Context, query string, arg string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, query, arg).Scan(&exists); err != nil {
		return false, sanitizePostgresError(err)
	}

	return exists, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sanitizePostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errors.New("user already exists")
		case "23514":
			return errors.New("user violates schema constraints")
		}
	}

	return fmt.Errorf("user repository operation failed")
}
