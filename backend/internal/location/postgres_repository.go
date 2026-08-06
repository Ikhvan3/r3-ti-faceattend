package location

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"r3-ti-faceattend/backend/internal/user"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) ([]OfficeLocation, error) {
	const query = `
		SELECT id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at
		FROM office_locations
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR COALESCE(address, '') ILIKE '%' || $1 || '%')
			AND ($2 = '' OR is_active = ($2 = 'ACTIVE'))
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, strings.TrimSpace(filter.Search), string(filter.Status), filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, sanitizePostgresError(err)
	}
	defer rows.Close()

	var offices []OfficeLocation
	for rows.Next() {
		office, err := scanOffice(rows)
		if err != nil {
			return nil, sanitizePostgresError(err)
		}
		offices = append(offices, office)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizePostgresError(err)
	}
	return offices, nil
}

func (r *PostgresRepository) CountOfficeLocations(ctx context.Context, filter OfficeLocationListFilter) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM office_locations
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR COALESCE(address, '') ILIKE '%' || $1 || '%')
			AND ($2 = '' OR is_active = ($2 = 'ACTIVE'))
	`
	var count int
	if err := r.pool.QueryRow(ctx, query, strings.TrimSpace(filter.Search), string(filter.Status)).Scan(&count); err != nil {
		return 0, sanitizePostgresError(err)
	}
	return count, nil
}

func (r *PostgresRepository) CreateOfficeLocation(ctx context.Context, office OfficeLocation) (OfficeLocation, error) {
	const query = `
		INSERT INTO office_locations (id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW(), NOW())
		RETURNING id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at
	`
	created, err := scanOffice(r.pool.QueryRow(ctx, query, office.ID, office.Name, office.Address, office.Latitude, office.Longitude, office.RadiusMeters))
	if err != nil {
		return OfficeLocation{}, sanitizePostgresError(err)
	}
	return created, nil
}

func (r *PostgresRepository) FindOfficeLocationByID(ctx context.Context, id string) (OfficeLocation, error) {
	const query = `
		SELECT id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at
		FROM office_locations
		WHERE id = $1
	`
	office, err := scanOffice(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	if err != nil {
		return OfficeLocation{}, sanitizePostgresError(err)
	}
	return office, nil
}

func (r *PostgresRepository) UpdateOfficeLocation(ctx context.Context, office OfficeLocation) (OfficeLocation, error) {
	const query = `
		UPDATE office_locations
		SET name = $2,
			address = $3,
			latitude = $4,
			longitude = $5,
			radius_meters = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at
	`
	updated, err := scanOffice(r.pool.QueryRow(ctx, query, office.ID, office.Name, office.Address, office.Latitude, office.Longitude, office.RadiusMeters))
	if errors.Is(err, pgx.ErrNoRows) {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	if err != nil {
		return OfficeLocation{}, sanitizePostgresError(err)
	}
	return updated, nil
}

func (r *PostgresRepository) UpdateOfficeLocationStatus(ctx context.Context, id string, isActive bool) (OfficeLocation, error) {
	const query = `
		UPDATE office_locations
		SET is_active = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, address, latitude, longitude, radius_meters, is_active, created_at, updated_at
	`
	updated, err := scanOffice(r.pool.QueryRow(ctx, query, id, isActive))
	if errors.Is(err, pgx.ErrNoRows) {
		return OfficeLocation{}, ErrOfficeNotFound
	}
	if err != nil {
		return OfficeLocation{}, sanitizePostgresError(err)
	}
	return updated, nil
}

func (r *PostgresRepository) HasActiveOrFutureLocationAssignments(ctx context.Context, officeID string, businessDate time.Time) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM employee_location_assignments
			WHERE office_location_id = $1
				AND (effective_to IS NULL OR effective_to >= $2)
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, officeID, businessDate).Scan(&exists); err != nil {
		return false, sanitizePostgresError(err)
	}
	return exists, nil
}

func (r *PostgresRepository) ListLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter, businessDate time.Time) ([]LocationAssignmentRecord, error) {
	const query = `
		SELECT ela.id, ela.effective_from, ela.effective_to, ela.created_at, ela.updated_at,
			u.id, u.employee_number, u.name, u.email, u.password_hash, u.phone, u.position, u.role, u.account_status, u.created_at, u.updated_at,
			ol.id, ol.name, ol.address, ol.latitude, ol.longitude, ol.radius_meters, ol.is_active, ol.created_at, ol.updated_at
		FROM employee_location_assignments ela
		JOIN users u ON u.id = ela.user_id
		JOIN office_locations ol ON ol.id = ela.office_location_id
		WHERE ($1 = '' OR (
				u.employee_number ILIKE '%' || $1 || '%' OR
				u.name ILIKE '%' || $1 || '%' OR
				u.email ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR ela.user_id = $2::uuid)
			AND ($3 = '' OR ela.office_location_id = $3::uuid)
			AND ($4 = '' OR
				($4 = 'CURRENT' AND ela.effective_from <= $5 AND (ela.effective_to IS NULL OR ela.effective_to >= $5)) OR
				($4 = 'UPCOMING' AND ela.effective_from > $5) OR
				($4 = 'ENDED' AND ela.effective_to IS NOT NULL AND ela.effective_to < $5)
			)
		ORDER BY ela.created_at DESC, ela.id DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.pool.Query(ctx, query, strings.TrimSpace(filter.Search), filter.UserID, filter.OfficeLocationID, string(filter.Status), businessDate, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, sanitizePostgresError(err)
	}
	defer rows.Close()

	var assignments []LocationAssignmentRecord
	for rows.Next() {
		assignment, err := scanAssignment(rows)
		if err != nil {
			return nil, sanitizePostgresError(err)
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizePostgresError(err)
	}
	return assignments, nil
}

func (r *PostgresRepository) CountLocationAssignments(ctx context.Context, filter LocationAssignmentListFilter, businessDate time.Time) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM employee_location_assignments ela
		JOIN users u ON u.id = ela.user_id
		WHERE ($1 = '' OR (
				u.employee_number ILIKE '%' || $1 || '%' OR
				u.name ILIKE '%' || $1 || '%' OR
				u.email ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR ela.user_id = $2::uuid)
			AND ($3 = '' OR ela.office_location_id = $3::uuid)
			AND ($4 = '' OR
				($4 = 'CURRENT' AND ela.effective_from <= $5 AND (ela.effective_to IS NULL OR ela.effective_to >= $5)) OR
				($4 = 'UPCOMING' AND ela.effective_from > $5) OR
				($4 = 'ENDED' AND ela.effective_to IS NOT NULL AND ela.effective_to < $5)
			)
	`
	var count int
	if err := r.pool.QueryRow(ctx, query, strings.TrimSpace(filter.Search), filter.UserID, filter.OfficeLocationID, string(filter.Status), businessDate).Scan(&count); err != nil {
		return 0, sanitizePostgresError(err)
	}
	return count, nil
}

func (r *PostgresRepository) FindLocationAssignmentByID(ctx context.Context, id string) (LocationAssignmentRecord, error) {
	const query = `
		SELECT ela.id, ela.effective_from, ela.effective_to, ela.created_at, ela.updated_at,
			u.id, u.employee_number, u.name, u.email, u.password_hash, u.phone, u.position, u.role, u.account_status, u.created_at, u.updated_at,
			ol.id, ol.name, ol.address, ol.latitude, ol.longitude, ol.radius_meters, ol.is_active, ol.created_at, ol.updated_at
		FROM employee_location_assignments ela
		JOIN users u ON u.id = ela.user_id
		JOIN office_locations ol ON ol.id = ela.office_location_id
		WHERE ela.id = $1
	`
	assignment, err := scanAssignment(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return LocationAssignmentRecord{}, ErrAssignmentNotFound
	}
	if err != nil {
		return LocationAssignmentRecord{}, sanitizePostgresError(err)
	}
	return assignment, nil
}

func (r *PostgresRepository) FindCurrentLocationAssignment(ctx context.Context, userID string, businessDate time.Time) (LocationAssignmentRecord, error) {
	const query = `
		SELECT ela.id, ela.effective_from, ela.effective_to, ela.created_at, ela.updated_at,
			u.id, u.employee_number, u.name, u.email, u.password_hash, u.phone, u.position, u.role, u.account_status, u.created_at, u.updated_at,
			ol.id, ol.name, ol.address, ol.latitude, ol.longitude, ol.radius_meters, ol.is_active, ol.created_at, ol.updated_at
		FROM employee_location_assignments ela
		JOIN users u ON u.id = ela.user_id
		JOIN office_locations ol ON ol.id = ela.office_location_id
		WHERE ela.user_id = $1
			AND ela.effective_from <= $2
			AND (ela.effective_to IS NULL OR ela.effective_to >= $2)
			AND ol.is_active = TRUE
		ORDER BY ela.effective_from DESC, ela.created_at DESC
		LIMIT 1
	`
	assignment, err := scanAssignment(r.pool.QueryRow(ctx, query, userID, businessDate))
	if errors.Is(err, pgx.ErrNoRows) {
		return LocationAssignmentRecord{}, ErrOfficeNotFound
	}
	if err != nil {
		return LocationAssignmentRecord{}, sanitizePostgresError(err)
	}
	return assignment, nil
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id string) (user.User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u user.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.EmployeeNumber, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.Position, &u.Role, &u.AccountStatus, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, sanitizePostgresError(err)
	}
	return u, nil
}

func (r *PostgresRepository) HasOverlappingLocationAssignment(ctx context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM employee_location_assignments
			WHERE user_id = $1
				AND ($2 = '' OR id <> $2::uuid)
				AND daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]')
					&& daterange($3::date, COALESCE($4::date, 'infinity'::date), '[]')
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, userID, assignmentID, effectiveFrom, effectiveTo).Scan(&exists); err != nil {
		return false, sanitizePostgresError(err)
	}
	return exists, nil
}

func (r *PostgresRepository) CreateLocationAssignment(ctx context.Context, assignmentID string, userID string, officeID string, effectiveFrom time.Time, effectiveTo *time.Time) (LocationAssignmentRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return LocationAssignmentRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	const insertQuery = `
		INSERT INTO employee_location_assignments (id, user_id, office_location_id, effective_from, effective_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	if _, err := tx.Exec(ctx, insertQuery, assignmentID, userID, officeID, effectiveFrom, effectiveTo); err != nil {
		return LocationAssignmentRecord{}, sanitizePostgresError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return LocationAssignmentRecord{}, sanitizePostgresError(err)
	}
	return r.FindLocationAssignmentByID(ctx, assignmentID)
}

func (r *PostgresRepository) EndLocationAssignment(ctx context.Context, id string, effectiveTo time.Time) (LocationAssignmentRecord, error) {
	const query = `
		UPDATE employee_location_assignments
		SET effective_to = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`
	var updatedID string
	if err := r.pool.QueryRow(ctx, query, id, effectiveTo).Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LocationAssignmentRecord{}, ErrAssignmentNotFound
		}
		return LocationAssignmentRecord{}, sanitizePostgresError(err)
	}
	return r.FindLocationAssignmentByID(ctx, updatedID)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOffice(row scanner) (OfficeLocation, error) {
	var office OfficeLocation
	err := row.Scan(&office.ID, &office.Name, &office.Address, &office.Latitude, &office.Longitude, &office.RadiusMeters, &office.IsActive, &office.CreatedAt, &office.UpdatedAt)
	return office, err
}

func scanAssignment(row scanner) (LocationAssignmentRecord, error) {
	var assignment LocationAssignmentRecord
	err := row.Scan(
		&assignment.ID,
		&assignment.EffectiveFrom,
		&assignment.EffectiveTo,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
		&assignment.User.ID,
		&assignment.User.EmployeeNumber,
		&assignment.User.Name,
		&assignment.User.Email,
		&assignment.User.PasswordHash,
		&assignment.User.Phone,
		&assignment.User.Position,
		&assignment.User.Role,
		&assignment.User.AccountStatus,
		&assignment.User.CreatedAt,
		&assignment.User.UpdatedAt,
		&assignment.Office.ID,
		&assignment.Office.Name,
		&assignment.Office.Address,
		&assignment.Office.Latitude,
		&assignment.Office.Longitude,
		&assignment.Office.RadiusMeters,
		&assignment.Office.IsActive,
		&assignment.Office.CreatedAt,
		&assignment.Office.UpdatedAt,
	)
	return assignment, err
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func sanitizePostgresError(err error) error {
	if errors.Is(err, ErrOfficeNotFound) || errors.Is(err, ErrAssignmentNotFound) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01":
			if pgErr.ConstraintName == "employee_location_assignments_user_period_no_overlap" {
				return ErrAssignmentOverlap
			}
			return ErrAssignmentOverlap
		case "23503":
			return ErrOfficeNotFound
		case "23514", "22P02":
			return ErrInvalidInput
		}
	}
	return fmt.Errorf("location repository operation failed")
}
