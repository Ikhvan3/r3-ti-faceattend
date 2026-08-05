package attendance

import (
	"context"
	"errors"
	"fmt"
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

func (r *PostgresRepository) Today(ctx context.Context, userID string, attendanceDate time.Time) (TodayData, error) {
	u, err := r.findUser(ctx, r.pool, userID)
	if err != nil {
		return TodayData{}, err
	}

	schedule, err := r.findSchedule(ctx, r.pool, userID, attendanceDate)
	if err != nil {
		return TodayData{}, err
	}

	record, err := r.findRecord(ctx, r.pool, userID, attendanceDate)
	if err != nil {
		return TodayData{}, err
	}

	return TodayData{User: u, Schedule: schedule, Record: record}, nil
}

func (r *PostgresRepository) CheckIn(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string) (AttendanceRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttendanceRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	schedule, err := r.findSchedule(ctx, tx, userID, attendanceDate)
	if err != nil {
		return AttendanceRecord{}, err
	}

	const query = `
		INSERT INTO attendance_records (
			id, user_id, schedule_id, attendance_date, check_in_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, user_id, schedule_id, attendance_date, check_in_at, check_out_at, created_at, updated_at
	`

	record, err := scanRecord(tx.QueryRow(ctx, query, recordID, userID, schedule.ID, attendanceDate, now))
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return AttendanceRecord{}, ErrInternal
	}

	return record, nil
}

func (r *PostgresRepository) CheckOut(ctx context.Context, userID string, attendanceDate time.Time, now time.Time) (AttendanceRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttendanceRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	const lockQuery = `
		SELECT id, user_id, schedule_id, attendance_date, check_in_at, check_out_at, created_at, updated_at
		FROM attendance_records
		WHERE user_id = $1 AND attendance_date = $2
		FOR UPDATE
	`
	record, err := scanRecord(tx.QueryRow(ctx, lockQuery, userID, attendanceDate))
	if errors.Is(err, pgx.ErrNoRows) {
		return AttendanceRecord{}, ErrNotCheckedIn
	}
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	if record.CheckOutAt != nil {
		return AttendanceRecord{}, ErrAlreadyCheckedOut
	}

	const updateQuery = `
		UPDATE attendance_records
		SET check_out_at = $3,
			updated_at = NOW()
		WHERE user_id = $1 AND attendance_date = $2
		RETURNING id, user_id, schedule_id, attendance_date, check_in_at, check_out_at, created_at, updated_at
	`
	updated, err := scanRecord(tx.QueryRow(ctx, updateQuery, userID, attendanceDate, now))
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return AttendanceRecord{}, ErrInternal
	}

	return updated, nil
}

func (r *PostgresRepository) ListHistory(ctx context.Context, userID string, filter HistoryFilter) ([]HistoryRow, error) {
	const query = `
		SELECT ar.id, ar.user_id, ar.schedule_id, ar.attendance_date, ar.check_in_at, ar.check_out_at, ar.created_at, ar.updated_at,
			ws.id, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'),
			ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at
		FROM attendance_records ar
		JOIN work_schedules ws ON ws.id = ar.schedule_id
		WHERE ar.user_id = $1
		ORDER BY ar.attendance_date DESC, ar.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, sanitizeAttendanceError(err)
	}
	defer rows.Close()

	var result []HistoryRow
	for rows.Next() {
		var row HistoryRow
		if err := rows.Scan(
			&row.Record.ID,
			&row.Record.UserID,
			&row.Record.ScheduleID,
			&row.Record.AttendanceDate,
			&row.Record.CheckInAt,
			&row.Record.CheckOutAt,
			&row.Record.CreatedAt,
			&row.Record.UpdatedAt,
			&row.Schedule.ID,
			&row.Schedule.Name,
			&row.Schedule.StartTime,
			&row.Schedule.EndTime,
			&row.Schedule.GraceMinutes,
			&row.Schedule.IsActive,
			&row.Schedule.CreatedAt,
			&row.Schedule.UpdatedAt,
		); err != nil {
			return nil, sanitizeAttendanceError(err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizeAttendanceError(err)
	}

	return result, nil
}

func (r *PostgresRepository) CountHistory(ctx context.Context, userID string) (int, error) {
	const query = `SELECT COUNT(*) FROM attendance_records WHERE user_id = $1`

	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, sanitizeAttendanceError(err)
	}

	return count, nil
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *PostgresRepository) findUser(ctx context.Context, q queryer, userID string) (user.User, error) {
	const query = `
		SELECT id, employee_number, name, email, password_hash, phone, position,
			role, account_status, created_at, updated_at
		FROM users
		WHERE id = $1 AND role = 'USER'
	`

	var u user.User
	err := q.QueryRow(ctx, query, userID).Scan(
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
		return user.User{}, ErrInactiveAccount
	}
	if err != nil {
		return user.User{}, sanitizeAttendanceError(err)
	}

	return u, nil
}

func (r *PostgresRepository) findSchedule(ctx context.Context, q queryer, userID string, attendanceDate time.Time) (WorkSchedule, error) {
	const query = `
		SELECT ws.id, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'),
			ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at
		FROM employee_schedule_assignments esa
		JOIN work_schedules ws ON ws.id = esa.schedule_id
		WHERE esa.user_id = $1
			AND esa.effective_from <= $2
			AND (esa.effective_to IS NULL OR esa.effective_to >= $2)
		ORDER BY esa.effective_from DESC, esa.created_at DESC
		LIMIT 1
	`

	var schedule WorkSchedule
	err := q.QueryRow(ctx, query, userID, attendanceDate).Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.GraceMinutes,
		&schedule.IsActive,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkSchedule{}, ErrScheduleNotFound
	}
	if err != nil {
		return WorkSchedule{}, sanitizeAttendanceError(err)
	}

	return schedule, nil
}

func (r *PostgresRepository) findRecord(ctx context.Context, q queryer, userID string, attendanceDate time.Time) (*AttendanceRecord, error) {
	const query = `
		SELECT id, user_id, schedule_id, attendance_date, check_in_at, check_out_at, created_at, updated_at
		FROM attendance_records
		WHERE user_id = $1 AND attendance_date = $2
	`

	record, err := scanRecord(q.QueryRow(ctx, query, userID, attendanceDate))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, sanitizeAttendanceError(err)
	}

	return &record, nil
}

func scanRecord(row pgx.Row) (AttendanceRecord, error) {
	var record AttendanceRecord
	err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.ScheduleID,
		&record.AttendanceDate,
		&record.CheckInAt,
		&record.CheckOutAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func sanitizeAttendanceError(err error) error {
	if errors.Is(err, ErrScheduleNotFound) || errors.Is(err, ErrNotCheckedIn) || errors.Is(err, ErrAlreadyCheckedOut) {
		return err
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "attendance_records_user_date_unique" {
				return ErrAlreadyCheckedIn
			}
			return errors.New("attendance already exists")
		case "23503":
			return ErrScheduleNotFound
		case "23514":
			return ErrInvalidInput
		}
	}

	return fmt.Errorf("attendance repository operation failed")
}
