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

func (r *PostgresRepository) CurrentOfficeLocation(ctx context.Context, userID string, attendanceDate time.Time) (AttendanceLocationTarget, error) {
	const query = `
		SELECT ol.id, ol.name, ol.latitude, ol.longitude, ol.radius_meters, ol.is_active
		FROM employee_location_assignments ela
		JOIN office_locations ol ON ol.id = ela.office_location_id
		WHERE ela.user_id = $1
			AND ela.effective_from <= $2
			AND (ela.effective_to IS NULL OR ela.effective_to >= $2)
		ORDER BY ela.effective_from DESC, ela.created_at DESC
		LIMIT 1
	`

	var target AttendanceLocationTarget
	err := r.pool.QueryRow(ctx, query, userID, attendanceDate).Scan(
		&target.OfficeLocationID,
		&target.OfficeLocationName,
		&target.Latitude,
		&target.Longitude,
		&target.RadiusMeters,
		&target.IsActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AttendanceLocationTarget{}, ErrLocationNotFound
	}
	if err != nil {
		return AttendanceLocationTarget{}, sanitizeAttendanceError(err)
	}

	return target, nil
}

func (r *PostgresRepository) CheckIn(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string, evidence AttendanceLocationEvidence) (AttendanceRecord, error) {
	return r.checkIn(ctx, userID, attendanceDate, now, recordID, evidence, "")
}

func (r *PostgresRepository) CheckInWithGrant(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string, evidence AttendanceLocationEvidence, grantHash string) (AttendanceRecord, error) {
	return r.checkIn(ctx, userID, attendanceDate, now, recordID, evidence, grantHash)
}

func (r *PostgresRepository) checkIn(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, recordID string, evidence AttendanceLocationEvidence, grantHash string) (AttendanceRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttendanceRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	if grantHash != "" {
		if err := validateAndLockGrant(ctx, tx, userID, "CHECK_IN", grantHash, now); err != nil {
			return AttendanceRecord{}, err
		}
	}

	schedule, err := r.findSchedule(ctx, tx, userID, attendanceDate)
	if err != nil {
		return AttendanceRecord{}, err
	}

	const query = `
		INSERT INTO attendance_records (
			id, user_id, schedule_id, attendance_date, check_in_at,
			check_in_location_id, check_in_latitude, check_in_longitude,
			check_in_accuracy_meters, check_in_distance_meters,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id
	`

	var insertedID string
	err = tx.QueryRow(
		ctx,
		query,
		recordID,
		userID,
		schedule.ID,
		attendanceDate,
		now,
		evidence.OfficeLocationID,
		evidence.Latitude,
		evidence.Longitude,
		evidence.AccuracyMeters,
		evidence.DistanceMeters,
	).Scan(&insertedID)
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	if grantHash != "" {
		if err := consumeGrant(ctx, tx, grantHash, now); err != nil {
			return AttendanceRecord{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return AttendanceRecord{}, ErrInternal
	}

	return r.findRecordByID(ctx, r.pool, insertedID)
}

func (r *PostgresRepository) CheckOut(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, evidence AttendanceLocationEvidence) (AttendanceRecord, error) {
	return r.checkOut(ctx, userID, attendanceDate, now, evidence, "")
}

func (r *PostgresRepository) CheckOutWithGrant(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, evidence AttendanceLocationEvidence, grantHash string) (AttendanceRecord, error) {
	return r.checkOut(ctx, userID, attendanceDate, now, evidence, grantHash)
}

func (r *PostgresRepository) checkOut(ctx context.Context, userID string, attendanceDate time.Time, now time.Time, evidence AttendanceLocationEvidence, grantHash string) (AttendanceRecord, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AttendanceRecord{}, ErrInternal
	}
	defer rollback(ctx, tx)

	if grantHash != "" {
		if err := validateAndLockGrant(ctx, tx, userID, "CHECK_OUT", grantHash, now); err != nil {
			return AttendanceRecord{}, err
		}
	}

	const lockQuery = `
		SELECT ar.id, ar.user_id, ar.schedule_id, ar.attendance_date, ar.check_in_at, ar.check_out_at,
			ar.check_in_location_id, cil.name, ar.check_in_latitude, ar.check_in_longitude, ar.check_in_accuracy_meters, ar.check_in_distance_meters,
			ar.check_out_location_id, col.name, ar.check_out_latitude, ar.check_out_longitude, ar.check_out_accuracy_meters, ar.check_out_distance_meters,
			ar.created_at, ar.updated_at
		FROM attendance_records ar
		LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
		LEFT JOIN office_locations col ON col.id = ar.check_out_location_id
		WHERE ar.user_id = $1 AND ar.attendance_date = $2
		FOR UPDATE OF ar
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
			check_out_location_id = $4,
			check_out_latitude = $5,
			check_out_longitude = $6,
			check_out_accuracy_meters = $7,
			check_out_distance_meters = $8,
			updated_at = NOW()
		WHERE user_id = $1 AND attendance_date = $2
		RETURNING id
	`
	var updatedID string
	err = tx.QueryRow(
		ctx,
		updateQuery,
		userID,
		attendanceDate,
		now,
		evidence.OfficeLocationID,
		evidence.Latitude,
		evidence.Longitude,
		evidence.AccuracyMeters,
		evidence.DistanceMeters,
	).Scan(&updatedID)
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	if grantHash != "" {
		if err := consumeGrant(ctx, tx, grantHash, now); err != nil {
			return AttendanceRecord{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return AttendanceRecord{}, ErrInternal
	}

	return r.findRecordByID(ctx, r.pool, updatedID)
}

func validateAndLockGrant(ctx context.Context, tx pgx.Tx, userID string, purpose string, grantHash string, now time.Time) error {
	const query = `
		SELECT user_id, purpose, expires_at, consumed_at
		FROM face_verification_grants
		WHERE token_hash = $1
		FOR UPDATE
	`
	var storedUserID string
	var storedPurpose string
	var expiresAt time.Time
	var consumedAt *time.Time
	err := tx.QueryRow(ctx, query, grantHash).Scan(&storedUserID, &storedPurpose, &expiresAt, &consumedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidGrant
	}
	if err != nil {
		return sanitizeAttendanceError(err)
	}
	if storedUserID != userID || storedPurpose != purpose {
		return ErrInvalidGrant
	}
	if consumedAt != nil {
		return ErrConsumedGrant
	}
	if !now.Before(expiresAt) {
		return ErrExpiredGrant
	}
	return nil
}

func consumeGrant(ctx context.Context, tx pgx.Tx, grantHash string, now time.Time) error {
	tag, err := tx.Exec(ctx, `UPDATE face_verification_grants SET consumed_at = $2 WHERE token_hash = $1 AND consumed_at IS NULL`, grantHash, now)
	if err != nil {
		return sanitizeAttendanceError(err)
	}
	if tag.RowsAffected() != 1 {
		return ErrConsumedGrant
	}
	return nil
}

func (r *PostgresRepository) ListHistory(ctx context.Context, userID string, filter HistoryFilter) ([]HistoryRow, error) {
	const query = `
		SELECT ar.id, ar.user_id, ar.schedule_id, ar.attendance_date, ar.check_in_at, ar.check_out_at,
			ar.check_in_location_id, cil.name, ar.check_in_latitude, ar.check_in_longitude, ar.check_in_accuracy_meters, ar.check_in_distance_meters,
			ar.check_out_location_id, col.name, ar.check_out_latitude, ar.check_out_longitude, ar.check_out_accuracy_meters, ar.check_out_distance_meters,
			ar.created_at, ar.updated_at,
			ws.id, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'),
			ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at
		FROM attendance_records ar
		JOIN work_schedules ws ON ws.id = ar.schedule_id
		LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
		LEFT JOIN office_locations col ON col.id = ar.check_out_location_id
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
		var checkInLocationID *string
		var checkInLocationName *string
		var checkInLatitude *float64
		var checkInLongitude *float64
		var checkInAccuracy *float64
		var checkInDistance *float64
		var checkOutLocationID *string
		var checkOutLocationName *string
		var checkOutLatitude *float64
		var checkOutLongitude *float64
		var checkOutAccuracy *float64
		var checkOutDistance *float64
		if err := rows.Scan(
			&row.Record.ID,
			&row.Record.UserID,
			&row.Record.ScheduleID,
			&row.Record.AttendanceDate,
			&row.Record.CheckInAt,
			&row.Record.CheckOutAt,
			&checkInLocationID,
			&checkInLocationName,
			&checkInLatitude,
			&checkInLongitude,
			&checkInAccuracy,
			&checkInDistance,
			&checkOutLocationID,
			&checkOutLocationName,
			&checkOutLatitude,
			&checkOutLongitude,
			&checkOutAccuracy,
			&checkOutDistance,
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
		row.Record.CheckInLocation = locationEvidence(checkInLocationID, checkInLocationName, checkInLatitude, checkInLongitude, checkInAccuracy, checkInDistance)
		row.Record.CheckOutLocation = locationEvidence(checkOutLocationID, checkOutLocationName, checkOutLatitude, checkOutLongitude, checkOutAccuracy, checkOutDistance)
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
		SELECT ar.id, ar.user_id, ar.schedule_id, ar.attendance_date, ar.check_in_at, ar.check_out_at,
			ar.check_in_location_id, cil.name, ar.check_in_latitude, ar.check_in_longitude, ar.check_in_accuracy_meters, ar.check_in_distance_meters,
			ar.check_out_location_id, col.name, ar.check_out_latitude, ar.check_out_longitude, ar.check_out_accuracy_meters, ar.check_out_distance_meters,
			ar.created_at, ar.updated_at
		FROM attendance_records ar
		LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
		LEFT JOIN office_locations col ON col.id = ar.check_out_location_id
		WHERE ar.user_id = $1 AND ar.attendance_date = $2
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

func (r *PostgresRepository) findRecordByID(ctx context.Context, q queryer, id string) (AttendanceRecord, error) {
	const query = `
		SELECT ar.id, ar.user_id, ar.schedule_id, ar.attendance_date, ar.check_in_at, ar.check_out_at,
			ar.check_in_location_id, cil.name, ar.check_in_latitude, ar.check_in_longitude, ar.check_in_accuracy_meters, ar.check_in_distance_meters,
			ar.check_out_location_id, col.name, ar.check_out_latitude, ar.check_out_longitude, ar.check_out_accuracy_meters, ar.check_out_distance_meters,
			ar.created_at, ar.updated_at
		FROM attendance_records ar
		LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
		LEFT JOIN office_locations col ON col.id = ar.check_out_location_id
		WHERE ar.id = $1
	`
	record, err := scanRecord(q.QueryRow(ctx, query, id))
	if err != nil {
		return AttendanceRecord{}, sanitizeAttendanceError(err)
	}
	return record, nil
}

func scanRecord(row pgx.Row) (AttendanceRecord, error) {
	var record AttendanceRecord
	var checkInLocationID *string
	var checkInLocationName *string
	var checkInLatitude *float64
	var checkInLongitude *float64
	var checkInAccuracy *float64
	var checkInDistance *float64
	var checkOutLocationID *string
	var checkOutLocationName *string
	var checkOutLatitude *float64
	var checkOutLongitude *float64
	var checkOutAccuracy *float64
	var checkOutDistance *float64
	err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.ScheduleID,
		&record.AttendanceDate,
		&record.CheckInAt,
		&record.CheckOutAt,
		&checkInLocationID,
		&checkInLocationName,
		&checkInLatitude,
		&checkInLongitude,
		&checkInAccuracy,
		&checkInDistance,
		&checkOutLocationID,
		&checkOutLocationName,
		&checkOutLatitude,
		&checkOutLongitude,
		&checkOutAccuracy,
		&checkOutDistance,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return record, err
	}
	record.CheckInLocation = locationEvidence(checkInLocationID, checkInLocationName, checkInLatitude, checkInLongitude, checkInAccuracy, checkInDistance)
	record.CheckOutLocation = locationEvidence(checkOutLocationID, checkOutLocationName, checkOutLatitude, checkOutLongitude, checkOutAccuracy, checkOutDistance)
	return record, err
}

func locationEvidence(locationID *string, locationName *string, latitude *float64, longitude *float64, accuracy *float64, distance *float64) *AttendanceLocationEvidence {
	if locationID == nil || locationName == nil || latitude == nil || longitude == nil || accuracy == nil || distance == nil {
		return nil
	}
	return &AttendanceLocationEvidence{
		OfficeLocationID:   *locationID,
		OfficeLocationName: *locationName,
		Latitude:           *latitude,
		Longitude:          *longitude,
		AccuracyMeters:     *accuracy,
		DistanceMeters:     *distance,
		InsideGeofence:     true,
	}
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func sanitizeAttendanceError(err error) error {
	if errors.Is(err, ErrScheduleNotFound) || errors.Is(err, ErrNotCheckedIn) || errors.Is(err, ErrAlreadyCheckedOut) || errors.Is(err, ErrLocationNotFound) {
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
