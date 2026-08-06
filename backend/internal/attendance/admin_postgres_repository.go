package attendance

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

type AdminPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAdminPostgresRepository(pool *pgxpool.Pool) *AdminPostgresRepository {
	return &AdminPostgresRepository{pool: pool}
}

func (r *AdminPostgresRepository) ListWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) ([]WorkSchedule, error) {
	const query = `
		SELECT id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
			grace_minutes, is_active, created_at, updated_at
		FROM work_schedules
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
			AND ($2 = '' OR is_active = ($2 = 'ACTIVE'))
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, strings.TrimSpace(filter.Search), string(filter.Status), filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}
	defer rows.Close()

	var schedules []WorkSchedule
	for rows.Next() {
		schedule, err := scanWorkSchedule(rows)
		if err != nil {
			return nil, sanitizeAdminPostgresError(err)
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}

	return schedules, nil
}

func (r *AdminPostgresRepository) CountWorkSchedules(ctx context.Context, filter WorkScheduleListFilter) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM work_schedules
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
			AND ($2 = '' OR is_active = ($2 = 'ACTIVE'))
	`
	var count int
	if err := r.pool.QueryRow(ctx, query, strings.TrimSpace(filter.Search), string(filter.Status)).Scan(&count); err != nil {
		return 0, sanitizeAdminPostgresError(err)
	}
	return count, nil
}

func (r *AdminPostgresRepository) CreateWorkSchedule(ctx context.Context, schedule WorkSchedule) (WorkSchedule, error) {
	const query = `
		INSERT INTO work_schedules (id, name, start_time, end_time, grace_minutes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3::time, $4::time, $5, TRUE, NOW(), NOW())
		RETURNING id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
			grace_minutes, is_active, created_at, updated_at
	`
	created, err := scanWorkSchedule(r.pool.QueryRow(ctx, query, schedule.ID, schedule.Name, schedule.StartTime, schedule.EndTime, schedule.GraceMinutes))
	if err != nil {
		return WorkSchedule{}, sanitizeAdminPostgresError(err)
	}
	return created, nil
}

func (r *AdminPostgresRepository) FindWorkScheduleByID(ctx context.Context, id string) (WorkSchedule, error) {
	const query = `
		SELECT id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
			grace_minutes, is_active, created_at, updated_at
		FROM work_schedules
		WHERE id = $1
	`
	schedule, err := scanWorkSchedule(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	if err != nil {
		return WorkSchedule{}, sanitizeAdminPostgresError(err)
	}
	return schedule, nil
}

func (r *AdminPostgresRepository) UpdateWorkSchedule(ctx context.Context, schedule WorkSchedule) (WorkSchedule, error) {
	const query = `
		UPDATE work_schedules
		SET name = $2,
			start_time = $3::time,
			end_time = $4::time,
			grace_minutes = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
			grace_minutes, is_active, created_at, updated_at
	`
	updated, err := scanWorkSchedule(r.pool.QueryRow(ctx, query, schedule.ID, schedule.Name, schedule.StartTime, schedule.EndTime, schedule.GraceMinutes))
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	if err != nil {
		return WorkSchedule{}, sanitizeAdminPostgresError(err)
	}
	return updated, nil
}

func (r *AdminPostgresRepository) UpdateWorkScheduleStatus(ctx context.Context, id string, isActive bool) (WorkSchedule, error) {
	const query = `
		UPDATE work_schedules
		SET is_active = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'),
			grace_minutes, is_active, created_at, updated_at
	`
	updated, err := scanWorkSchedule(r.pool.QueryRow(ctx, query, id, isActive))
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	if err != nil {
		return WorkSchedule{}, sanitizeAdminPostgresError(err)
	}
	return updated, nil
}

func (r *AdminPostgresRepository) HasActiveOrFutureAssignments(ctx context.Context, scheduleID string, businessDate time.Time) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM employee_schedule_assignments
			WHERE schedule_id = $1
				AND (effective_to IS NULL OR effective_to >= $2)
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, scheduleID, businessDate).Scan(&exists); err != nil {
		return false, sanitizeAdminPostgresError(err)
	}
	return exists, nil
}

func (r *AdminPostgresRepository) ListAssignments(ctx context.Context, filter AssignmentListFilter, businessDate time.Time) ([]ScheduleAssignmentRecord, error) {
	const query = `
		SELECT esa.id, esa.effective_from, esa.effective_to, esa.created_at, esa.updated_at,
			u.id, u.employee_number, u.name, u.email, u.password_hash, u.phone, u.position, u.role, u.account_status, u.created_at, u.updated_at,
			ws.id, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'), ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at
		FROM employee_schedule_assignments esa
		JOIN users u ON u.id = esa.user_id
		JOIN work_schedules ws ON ws.id = esa.schedule_id
		WHERE ($1 = '' OR (
				u.employee_number ILIKE '%' || $1 || '%' OR
				u.name ILIKE '%' || $1 || '%' OR
				u.email ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR esa.user_id = $2::uuid)
			AND ($3 = '' OR esa.schedule_id = $3::uuid)
			AND ($4 = '' OR
				($4 = 'CURRENT' AND esa.effective_from <= $5 AND (esa.effective_to IS NULL OR esa.effective_to >= $5)) OR
				($4 = 'UPCOMING' AND esa.effective_from > $5) OR
				($4 = 'ENDED' AND esa.effective_to IS NOT NULL AND esa.effective_to < $5)
			)
		ORDER BY esa.created_at DESC, esa.id DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.pool.Query(ctx, query,
		strings.TrimSpace(filter.Search),
		filter.UserID,
		filter.ScheduleID,
		string(filter.Status),
		businessDate,
		filter.PageSize,
		(filter.Page-1)*filter.PageSize,
	)
	if err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}
	defer rows.Close()

	var assignments []ScheduleAssignmentRecord
	for rows.Next() {
		assignment, err := scanAssignment(rows)
		if err != nil {
			return nil, sanitizeAdminPostgresError(err)
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}
	return assignments, nil
}

func (r *AdminPostgresRepository) CountAssignments(ctx context.Context, filter AssignmentListFilter, businessDate time.Time) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM employee_schedule_assignments esa
		JOIN users u ON u.id = esa.user_id
		WHERE ($1 = '' OR (
				u.employee_number ILIKE '%' || $1 || '%' OR
				u.name ILIKE '%' || $1 || '%' OR
				u.email ILIKE '%' || $1 || '%'
			))
			AND ($2 = '' OR esa.user_id = $2::uuid)
			AND ($3 = '' OR esa.schedule_id = $3::uuid)
			AND ($4 = '' OR
				($4 = 'CURRENT' AND esa.effective_from <= $5 AND (esa.effective_to IS NULL OR esa.effective_to >= $5)) OR
				($4 = 'UPCOMING' AND esa.effective_from > $5) OR
				($4 = 'ENDED' AND esa.effective_to IS NOT NULL AND esa.effective_to < $5)
			)
	`
	var count int
	if err := r.pool.QueryRow(ctx, query, strings.TrimSpace(filter.Search), filter.UserID, filter.ScheduleID, string(filter.Status), businessDate).Scan(&count); err != nil {
		return 0, sanitizeAdminPostgresError(err)
	}
	return count, nil
}

func (r *AdminPostgresRepository) FindAssignmentByID(ctx context.Context, id string) (ScheduleAssignmentRecord, error) {
	const query = `
		SELECT esa.id, esa.effective_from, esa.effective_to, esa.created_at, esa.updated_at,
			u.id, u.employee_number, u.name, u.email, u.password_hash, u.phone, u.position, u.role, u.account_status, u.created_at, u.updated_at,
			ws.id, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'), ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at
		FROM employee_schedule_assignments esa
		JOIN users u ON u.id = esa.user_id
		JOIN work_schedules ws ON ws.id = esa.schedule_id
		WHERE esa.id = $1
	`
	assignment, err := scanAssignment(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return ScheduleAssignmentRecord{}, ErrAssignmentNotFound
	}
	if err != nil {
		return ScheduleAssignmentRecord{}, sanitizeAdminPostgresError(err)
	}
	return assignment, nil
}

func (r *AdminPostgresRepository) FindUserByID(ctx context.Context, id string) (user.User, error) {
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
		return user.User{}, sanitizeAdminPostgresError(err)
	}
	return u, nil
}

func (r *AdminPostgresRepository) HasOverlappingAssignment(ctx context.Context, userID string, assignmentID string, effectiveFrom time.Time, effectiveTo *time.Time) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM employee_schedule_assignments
			WHERE user_id = $1
				AND ($2 = '' OR id <> $2::uuid)
				AND daterange(effective_from, COALESCE(effective_to, 'infinity'::date), '[]')
					&& daterange($3::date, COALESCE($4::date, 'infinity'::date), '[]')
		)
	`
	var exists bool
	if err := r.pool.QueryRow(ctx, query, userID, assignmentID, effectiveFrom, effectiveTo).Scan(&exists); err != nil {
		return false, sanitizeAdminPostgresError(err)
	}
	return exists, nil
}

func (r *AdminPostgresRepository) CreateAssignment(ctx context.Context, assignmentID string, userID string, scheduleID string, effectiveFrom time.Time, effectiveTo *time.Time) (ScheduleAssignmentRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ScheduleAssignmentRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	const insertQuery = `
		INSERT INTO employee_schedule_assignments (id, user_id, schedule_id, effective_from, effective_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	if _, err := tx.Exec(ctx, insertQuery, assignmentID, userID, scheduleID, effectiveFrom, effectiveTo); err != nil {
		return ScheduleAssignmentRecord{}, sanitizeAdminPostgresError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ScheduleAssignmentRecord{}, sanitizeAdminPostgresError(err)
	}

	return r.FindAssignmentByID(ctx, assignmentID)
}

func (r *AdminPostgresRepository) EndAssignment(ctx context.Context, id string, effectiveTo time.Time) (ScheduleAssignmentRecord, error) {
	const query = `
		UPDATE employee_schedule_assignments
		SET effective_to = $2,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id
	`
	var updatedID string
	if err := r.pool.QueryRow(ctx, query, id, effectiveTo).Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ScheduleAssignmentRecord{}, ErrAssignmentNotFound
		}
		return ScheduleAssignmentRecord{}, sanitizeAdminPostgresError(err)
	}
	return r.FindAssignmentByID(ctx, updatedID)
}

type workScheduleScanner interface {
	Scan(dest ...any) error
}

func scanWorkSchedule(row workScheduleScanner) (WorkSchedule, error) {
	var schedule WorkSchedule
	err := row.Scan(&schedule.ID, &schedule.Name, &schedule.StartTime, &schedule.EndTime, &schedule.GraceMinutes, &schedule.IsActive, &schedule.CreatedAt, &schedule.UpdatedAt)
	return schedule, err
}

func scanAssignment(row workScheduleScanner) (ScheduleAssignmentRecord, error) {
	var assignment ScheduleAssignmentRecord
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
		&assignment.Schedule.ID,
		&assignment.Schedule.Name,
		&assignment.Schedule.StartTime,
		&assignment.Schedule.EndTime,
		&assignment.Schedule.GraceMinutes,
		&assignment.Schedule.IsActive,
		&assignment.Schedule.CreatedAt,
		&assignment.Schedule.UpdatedAt,
	)
	return assignment, err
}

func sanitizeAdminPostgresError(err error) error {
	if errors.Is(err, ErrScheduleNotFound) || errors.Is(err, ErrAssignmentNotFound) {
		return err
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "work_schedules_name_unique" {
				return ErrScheduleDuplicate
			}
			return ErrAssignmentOverlap
		case "23P01":
			if pgErr.ConstraintName == "employee_schedule_assignments_user_period_no_overlap" {
				return ErrAssignmentOverlap
			}
			return ErrAssignmentOverlap
		case "23503":
			return ErrScheduleNotFound
		case "23514", "22P02":
			return ErrInvalidInput
		}
	}

	return fmt.Errorf("admin schedule repository operation failed")
}
