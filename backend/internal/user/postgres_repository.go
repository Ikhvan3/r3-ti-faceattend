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

func (r *PostgresRepository) ListEmployees(ctx context.Context, filter EmployeeListFilter) ([]User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE role = 'USER'
			AND ($1 = '' OR (
				employee_number ILIKE '%' || $1 || '%' OR
				name ILIKE '%' || $1 || '%' OR
				email ILIKE '%' || $1 || '%' OR
				position ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR account_status = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query,
		strings.TrimSpace(filter.Search),
		string(filter.Status),
		filter.PageSize,
		(filter.Page-1)*filter.PageSize,
	)
	if err != nil {
		return nil, sanitizePostgresError(err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
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
		); err != nil {
			return nil, sanitizePostgresError(err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizePostgresError(err)
	}

	return users, nil
}

func (r *PostgresRepository) CountEmployees(ctx context.Context, filter EmployeeListFilter) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM users
		WHERE role = 'USER'
			AND ($1 = '' OR (
				employee_number ILIKE '%' || $1 || '%' OR
				name ILIKE '%' || $1 || '%' OR
				email ILIKE '%' || $1 || '%' OR
				position ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR account_status = $2)
	`

	var count int
	if err := r.pool.QueryRow(ctx, query, strings.TrimSpace(filter.Search), string(filter.Status)).Scan(&count); err != nil {
		return 0, sanitizePostgresError(err)
	}

	return count, nil
}

func (r *PostgresRepository) FindEmployeeByID(ctx context.Context, id string) (User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE id = $1 AND role = 'USER'
	`

	return r.findOne(ctx, query, strings.TrimSpace(id))
}

func (r *PostgresRepository) CreateEmployee(ctx context.Context, u User) (User, error) {
	const query = `
		INSERT INTO users (
			id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'USER', 'ACTIVE', NOW(), NOW())
		RETURNING id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
	`

	var created User
	err := r.pool.QueryRow(ctx, query,
		u.ID,
		u.EmployeeNumber,
		u.Name,
		normalizeEmail(u.Email),
		u.PasswordHash,
		u.Phone,
		u.Position,
	).Scan(
		&created.ID,
		&created.EmployeeNumber,
		&created.Name,
		&created.Email,
		&created.PasswordHash,
		&created.Phone,
		&created.Position,
		&created.Role,
		&created.AccountStatus,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return User{}, sanitizePostgresError(err)
	}

	return created, nil
}

func (r *PostgresRepository) UpdateEmployee(ctx context.Context, u User) (User, error) {
	const query = `
		UPDATE users
		SET employee_number = $2,
			name = $3,
			email = $4,
			phone = $5,
			position = $6,
			updated_at = NOW()
		WHERE id = $1 AND role = 'USER'
		RETURNING id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
	`

	return r.updateOne(ctx, query,
		strings.TrimSpace(u.ID),
		u.EmployeeNumber,
		u.Name,
		normalizeEmail(u.Email),
		u.Phone,
		u.Position,
	)
}

func (r *PostgresRepository) UpdateEmployeeStatus(ctx context.Context, id string, status AccountStatus) (User, error) {
	const query = `
		UPDATE users
		SET account_status = $2,
			updated_at = NOW()
		WHERE id = $1 AND role = 'USER'
		RETURNING id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
	`

	return r.updateOne(ctx, query, strings.TrimSpace(id), status)
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

func (r *PostgresRepository) updateOne(ctx context.Context, query string, args ...any) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
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

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sanitizePostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "users_email_lower_unique":
				return ErrEmailConflict
			case "users_employee_number_unique":
				return ErrEmployeeNumberConflict
			default:
				return errors.New("user already exists")
			}
		case "23514":
			return errors.New("user violates schema constraints")
		}
	}

	return fmt.Errorf("user repository operation failed")
}
