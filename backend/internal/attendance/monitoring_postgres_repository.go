package attendance

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminAttendancePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAdminAttendancePostgresRepository(pool *pgxpool.Pool) *AdminAttendancePostgresRepository {
	return &AdminAttendancePostgresRepository{pool: pool}
}

func (r *AdminAttendancePostgresRepository) Summary(ctx context.Context, businessDate time.Time, timezone string) (adminAttendanceSummaryRow, error) {
	const query = `
		WITH scheduled AS (
			SELECT u.id, ws.start_time, ws.grace_minutes, ar.id AS attendance_id,
				ar.check_in_at, ar.check_out_at
			FROM employee_schedule_assignments esa
			JOIN users u ON u.id = esa.user_id
			JOIN work_schedules ws ON ws.id = esa.schedule_id
			LEFT JOIN attendance_records ar
				ON ar.user_id = u.id AND ar.attendance_date = $1::date
			WHERE u.role = 'USER'
				AND u.account_status = 'ACTIVE'
				AND ws.is_active = TRUE
				AND esa.effective_from <= $1::date
				AND (esa.effective_to IS NULL OR esa.effective_to >= $1::date)
		)
		SELECT COUNT(*)::int,
			COUNT(*) FILTER (WHERE attendance_id IS NOT NULL)::int,
			COUNT(*) FILTER (WHERE check_out_at IS NOT NULL)::int,
			COUNT(*) FILTER (WHERE attendance_id IS NULL)::int,
			COUNT(*) FILTER (
				WHERE attendance_id IS NOT NULL
				AND (check_in_at AT TIME ZONE $2)::time > start_time + make_interval(mins => grace_minutes)
			)::int
		FROM scheduled
	`

	var row adminAttendanceSummaryRow
	if err := r.pool.QueryRow(ctx, query, businessDate, timezone).Scan(
		&row.ActiveEmployees,
		&row.CheckedIn,
		&row.CheckedOut,
		&row.NotCheckedIn,
		&row.Late,
	); err != nil {
		return adminAttendanceSummaryRow{}, sanitizeAdminPostgresError(err)
	}
	return row, nil
}

func (r *AdminAttendancePostgresRepository) List(ctx context.Context, filter adminAttendanceQuery) ([]adminAttendanceRow, error) {
	const query = `
		WITH days AS (
			SELECT generate_series($1::date, $2::date, interval '1 day')::date AS business_date
		), scheduled AS (
			SELECT d.business_date,
				u.id AS user_id, u.employee_number, u.name, u.email, u.position,
				ws.id AS schedule_id, ws.name AS schedule_name,
				to_char(ws.start_time, 'HH24:MI') AS start_time,
				to_char(ws.end_time, 'HH24:MI') AS end_time,
				ws.grace_minutes, ws.is_active, ws.created_at AS schedule_created_at, ws.updated_at AS schedule_updated_at,
				ar.id::text AS attendance_id, ar.check_in_at, ar.check_out_at,
				CASE
					WHEN ar.id IS NULL THEN 'NOT_CHECKED_IN'
					WHEN ar.check_out_at IS NULL THEN 'CHECKED_IN'
					ELSE 'CHECKED_OUT'
				END AS attendance_state,
				CASE
					WHEN ar.id IS NULL THEN FALSE
					ELSE (ar.check_in_at AT TIME ZONE $9)::time > ws.start_time + make_interval(mins => ws.grace_minutes)
				END AS is_late,
				COALESCE(cil.id::text, '') AS office_id,
				COALESCE(cil.name, '') AS office_name,
				COALESCE(cil.radius_meters, 0) AS office_radius
			FROM days d
			JOIN employee_schedule_assignments esa
				ON esa.effective_from <= d.business_date
				AND (esa.effective_to IS NULL OR esa.effective_to >= d.business_date)
			JOIN users u ON u.id = esa.user_id
			JOIN work_schedules ws ON ws.id = esa.schedule_id
			LEFT JOIN attendance_records ar
				ON ar.user_id = u.id AND ar.attendance_date = d.business_date
			LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
			WHERE u.role = 'USER'
				AND u.account_status = 'ACTIVE'
				AND ws.is_active = TRUE
				AND ($3 = '' OR u.id = $3::uuid)
				AND ($4 = '' OR u.employee_number ILIKE '%' || $4 || '%' OR u.name ILIKE '%' || $4 || '%' OR u.email ILIKE '%' || $4 || '%')
		)
		SELECT business_date,
			user_id::text, employee_number, name, email, position,
			schedule_id::text, schedule_name, start_time, end_time, grace_minutes, is_active, schedule_created_at, schedule_updated_at,
			attendance_id, check_in_at, check_out_at, attendance_state, is_late,
			office_id, office_name, office_radius
		FROM scheduled
		WHERE ($5 = '' OR attendance_state = $5)
			AND ($6::boolean IS NULL OR is_late = $6)
		ORDER BY business_date DESC, name ASC, user_id ASC
		LIMIT $7 OFFSET $8
	`

	rows, err := r.pool.Query(ctx, query,
		filter.DateFrom,
		filter.DateTo,
		filter.EmployeeID,
		filter.Search,
		string(filter.AttendanceState),
		filter.IsLate,
		filter.PageSize,
		(filter.Page-1)*filter.PageSize,
		filter.Timezone,
	)
	if err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}
	defer rows.Close()

	items := make([]adminAttendanceRow, 0)
	for rows.Next() {
		row, err := scanAdminAttendanceListRow(rows)
		if err != nil {
			return nil, sanitizeAdminPostgresError(err)
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizeAdminPostgresError(err)
	}
	return items, nil
}

func (r *AdminAttendancePostgresRepository) Count(ctx context.Context, filter adminAttendanceQuery) (int, error) {
	const query = `
		WITH days AS (
			SELECT generate_series($1::date, $2::date, interval '1 day')::date AS business_date
		), scheduled AS (
			SELECT d.business_date,
				u.id AS user_id, u.employee_number, u.name, u.email,
				CASE
					WHEN ar.id IS NULL THEN 'NOT_CHECKED_IN'
					WHEN ar.check_out_at IS NULL THEN 'CHECKED_IN'
					ELSE 'CHECKED_OUT'
				END AS attendance_state,
				CASE
					WHEN ar.id IS NULL THEN FALSE
					ELSE (ar.check_in_at AT TIME ZONE $7)::time > ws.start_time + make_interval(mins => ws.grace_minutes)
				END AS is_late
			FROM days d
			JOIN employee_schedule_assignments esa
				ON esa.effective_from <= d.business_date
				AND (esa.effective_to IS NULL OR esa.effective_to >= d.business_date)
			JOIN users u ON u.id = esa.user_id
			JOIN work_schedules ws ON ws.id = esa.schedule_id
			LEFT JOIN attendance_records ar
				ON ar.user_id = u.id AND ar.attendance_date = d.business_date
			WHERE u.role = 'USER'
				AND u.account_status = 'ACTIVE'
				AND ws.is_active = TRUE
				AND ($3 = '' OR u.id = $3::uuid)
				AND ($4 = '' OR u.employee_number ILIKE '%' || $4 || '%' OR u.name ILIKE '%' || $4 || '%' OR u.email ILIKE '%' || $4 || '%')
		)
		SELECT COUNT(*)::int
		FROM scheduled
		WHERE ($5 = '' OR attendance_state = $5)
			AND ($6::boolean IS NULL OR is_late = $6)
	`

	var count int
	if err := r.pool.QueryRow(ctx, query,
		filter.DateFrom,
		filter.DateTo,
		filter.EmployeeID,
		filter.Search,
		string(filter.AttendanceState),
		filter.IsLate,
		filter.Timezone,
	).Scan(&count); err != nil {
		return 0, sanitizeAdminPostgresError(err)
	}
	return count, nil
}

func (r *AdminAttendancePostgresRepository) Detail(ctx context.Context, id string, timezone string) (adminAttendanceRow, error) {
	const query = `
		SELECT ar.id::text, ar.attendance_date,
			u.id::text, u.employee_number, u.name, u.email, u.position,
			ws.id::text, ws.name, to_char(ws.start_time, 'HH24:MI'), to_char(ws.end_time, 'HH24:MI'),
			ws.grace_minutes, ws.is_active, ws.created_at, ws.updated_at,
			ar.check_in_at, ar.check_out_at,
			CASE WHEN ar.check_out_at IS NULL THEN 'CHECKED_IN' ELSE 'CHECKED_OUT' END,
			(ar.check_in_at AT TIME ZONE $2)::time > ws.start_time + make_interval(mins => ws.grace_minutes),
			COALESCE(cil.id::text, ''), COALESCE(cil.name, ''), COALESCE(cil.radius_meters, 0),
			COALESCE(ar.check_in_latitude, 0), COALESCE(ar.check_in_longitude, 0), COALESCE(ar.check_in_accuracy_meters, 0), COALESCE(ar.check_in_distance_meters, 0),
			COALESCE(col.id::text, ''), COALESCE(col.name, ''), COALESCE(col.radius_meters, 0),
			COALESCE(ar.check_out_latitude, 0), COALESCE(ar.check_out_longitude, 0), COALESCE(ar.check_out_accuracy_meters, 0), COALESCE(ar.check_out_distance_meters, 0)
		FROM attendance_records ar
		JOIN users u ON u.id = ar.user_id
		JOIN work_schedules ws ON ws.id = ar.schedule_id
		LEFT JOIN office_locations cil ON cil.id = ar.check_in_location_id
		LEFT JOIN office_locations col ON col.id = ar.check_out_location_id
		WHERE ar.id = $1::uuid
	`

	row, err := scanAdminAttendanceDetailRow(r.pool.QueryRow(ctx, query, id, timezone))
	if errors.Is(err, pgx.ErrNoRows) {
		return adminAttendanceRow{}, ErrAdminAttendanceNotFound
	}
	if err != nil {
		return adminAttendanceRow{}, sanitizeAdminPostgresError(err)
	}
	return row, nil
}

type adminAttendanceScanner interface {
	Scan(dest ...any) error
}

func scanAdminAttendanceListRow(scanner adminAttendanceScanner) (adminAttendanceRow, error) {
	var (
		position        pgtype.Text
		attendanceID    pgtype.Text
		checkInAt       pgtype.Timestamptz
		checkOutAt      pgtype.Timestamptz
		employeeID      string
		employeeNo      string
		employeeName    string
		employeeEmail   string
		scheduleID      string
		scheduleName    string
		startTime       string
		endTime         string
		graceMinutes    int
		scheduleActive  bool
		scheduleCreated time.Time
		scheduleUpdated time.Time
		state           string
		isLate          bool
		officeID        string
		officeName      string
		officeRadius    int
		businessDate    time.Time
	)

	if err := scanner.Scan(
		&businessDate,
		&employeeID, &employeeNo, &employeeName, &employeeEmail, &position,
		&scheduleID, &scheduleName, &startTime, &endTime, &graceMinutes, &scheduleActive, &scheduleCreated, &scheduleUpdated,
		&attendanceID, &checkInAt, &checkOutAt, &state, &isLate,
		&officeID, &officeName, &officeRadius,
	); err != nil {
		return adminAttendanceRow{}, err
	}

	row := adminAttendanceRow{
		AttendanceDate: businessDate,
		Employee: AdminAttendanceEmployee{
			ID: employeeID, EmployeeNumber: employeeNo, Name: employeeName, Email: employeeEmail,
		},
		Schedule: WorkSchedule{
			ID: scheduleID, Name: scheduleName, StartTime: startTime, EndTime: endTime,
			GraceMinutes: graceMinutes, IsActive: scheduleActive, CreatedAt: scheduleCreated, UpdatedAt: scheduleUpdated,
		},
		AttendanceState: AdminAttendanceState(state),
		IsLate:          isLate,
	}
	if position.Valid {
		row.Employee.Position = &position.String
	}
	if attendanceID.Valid && attendanceID.String != "" {
		value := attendanceID.String
		row.ID = &value
	}
	if checkInAt.Valid {
		value := checkInAt.Time
		row.CheckInAt = &value
	}
	if checkOutAt.Valid {
		value := checkOutAt.Time
		row.CheckOutAt = &value
	}
	if officeID != "" {
		row.OfficeLocation = &AdminAttendanceOfficeLocation{ID: officeID, Name: officeName, RadiusMeters: officeRadius}
	}
	return row, nil
}

func scanAdminAttendanceDetailRow(scanner adminAttendanceScanner) (adminAttendanceRow, error) {
	var (
		id, employeeID, employeeNo, employeeName, employeeEmail      string
		position                                                     pgtype.Text
		scheduleID, scheduleName, startTime, endTime                 string
		graceMinutes                                                 int
		scheduleActive                                               bool
		scheduleCreated, scheduleUpdated                             time.Time
		businessDate, checkInAt                                      time.Time
		checkOutAt                                                   pgtype.Timestamptz
		state                                                        string
		isLate                                                       bool
		checkInOfficeID, checkInOfficeName                           string
		checkInRadius                                                int
		checkInLat, checkInLng, checkInAccuracy, checkInDistance     float64
		checkOutOfficeID, checkOutOfficeName                         string
		checkOutRadius                                               int
		checkOutLat, checkOutLng, checkOutAccuracy, checkOutDistance float64
	)
	if err := scanner.Scan(
		&id, &businessDate,
		&employeeID, &employeeNo, &employeeName, &employeeEmail, &position,
		&scheduleID, &scheduleName, &startTime, &endTime, &graceMinutes, &scheduleActive, &scheduleCreated, &scheduleUpdated,
		&checkInAt, &checkOutAt, &state, &isLate,
		&checkInOfficeID, &checkInOfficeName, &checkInRadius,
		&checkInLat, &checkInLng, &checkInAccuracy, &checkInDistance,
		&checkOutOfficeID, &checkOutOfficeName, &checkOutRadius,
		&checkOutLat, &checkOutLng, &checkOutAccuracy, &checkOutDistance,
	); err != nil {
		return adminAttendanceRow{}, err
	}

	row := adminAttendanceRow{
		ID:              &id,
		AttendanceDate:  businessDate,
		Employee:        AdminAttendanceEmployee{ID: employeeID, EmployeeNumber: employeeNo, Name: employeeName, Email: employeeEmail},
		Schedule:        WorkSchedule{ID: scheduleID, Name: scheduleName, StartTime: startTime, EndTime: endTime, GraceMinutes: graceMinutes, IsActive: scheduleActive, CreatedAt: scheduleCreated, UpdatedAt: scheduleUpdated},
		CheckInAt:       &checkInAt,
		AttendanceState: AdminAttendanceState(state),
		IsLate:          isLate,
	}
	if position.Valid {
		row.Employee.Position = &position.String
	}
	if checkOutAt.Valid {
		value := checkOutAt.Time
		row.CheckOutAt = &value
	}
	if checkInOfficeID != "" {
		row.CheckInLocation = &AdminAttendanceLocationEvidence{
			OfficeLocationID: checkInOfficeID, OfficeLocationName: checkInOfficeName, RadiusMeters: checkInRadius,
			Latitude: checkInLat, Longitude: checkInLng, AccuracyMeters: checkInAccuracy, DistanceMeters: checkInDistance, InsideGeofence: checkInDistance <= float64(checkInRadius),
		}
	}
	if checkOutOfficeID != "" {
		row.CheckOutLocation = &AdminAttendanceLocationEvidence{
			OfficeLocationID: checkOutOfficeID, OfficeLocationName: checkOutOfficeName, RadiusMeters: checkOutRadius,
			Latitude: checkOutLat, Longitude: checkOutLng, AccuracyMeters: checkOutAccuracy, DistanceMeters: checkOutDistance, InsideGeofence: checkOutDistance <= float64(checkOutRadius),
		}
	}
	return row, nil
}
