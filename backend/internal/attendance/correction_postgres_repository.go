package attendance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"r3-ti-faceattend/backend/internal/audit"
)

type AdminAttendanceCorrectionPostgresRepository struct {
	pool  *pgxpool.Pool
	audit *audit.PostgresRepository
}

func NewAdminAttendanceCorrectionPostgresRepository(pool *pgxpool.Pool, auditRepo *audit.PostgresRepository) *AdminAttendanceCorrectionPostgresRepository {
	return &AdminAttendanceCorrectionPostgresRepository{pool: pool, audit: auditRepo}
}

func (r *AdminAttendanceCorrectionPostgresRepository) Correct(ctx context.Context, command adminAttendanceCorrectionCommand) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrInternal
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const selectQuery = `
		SELECT ar.attendance_date, ar.user_id::text, ar.check_in_at, ar.check_out_at,
			u.employee_number, u.name
		FROM attendance_records ar
		JOIN users u ON u.id = ar.user_id
		WHERE ar.id = $1::uuid
		FOR UPDATE OF ar
	`
	var (
		attendanceDate time.Time
		targetUserID   string
		oldCheckIn     time.Time
		oldCheckOut    pgtype.Timestamptz
		employeeNumber string
		employeeName   string
	)
	if err := tx.QueryRow(ctx, selectQuery, command.AttendanceID).Scan(
		&attendanceDate,
		&targetUserID,
		&oldCheckIn,
		&oldCheckOut,
		&employeeNumber,
		&employeeName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAdminAttendanceNotFound
		}
		return ErrInternal
	}

	location, err := time.LoadLocation(command.Timezone)
	if err != nil {
		return ErrInternal
	}
	businessDate := attendanceDate.In(location).Format("2006-01-02")
	newCheckIn, err := time.ParseInLocation("2006-01-02 15:04", businessDate+" "+command.CheckInTime, location)
	if err != nil {
		return ErrAttendanceCorrectionInvalid
	}

	var newCheckOut *time.Time
	if command.CheckOutTime != nil {
		value := strings.TrimSpace(*command.CheckOutTime)
		if value != "" {
			parsed, err := time.ParseInLocation("2006-01-02 15:04", businessDate+" "+value, location)
			if err != nil || parsed.Before(newCheckIn) {
				return ErrAttendanceCorrectionInvalid
			}
			newCheckOut = &parsed
		}
	}

	const updateQuery = `
		UPDATE attendance_records
		SET check_in_at = $2, check_out_at = $3, updated_at = NOW()
		WHERE id = $1::uuid
	`
	if _, err := tx.Exec(ctx, updateQuery, command.AttendanceID, newCheckIn.UTC(), newCheckOut); err != nil {
		return ErrInternal
	}

	before := map[string]any{
		"check_in_at": oldCheckIn.UTC().Format(time.RFC3339Nano),
		"check_out_at": nil,
	}
	if oldCheckOut.Valid {
		before["check_out_at"] = oldCheckOut.Time.UTC().Format(time.RFC3339Nano)
	}
	after := map[string]any{
		"check_in_at": newCheckIn.UTC().Format(time.RFC3339Nano),
		"check_out_at": nil,
	}
	if newCheckOut != nil {
		after["check_out_at"] = newCheckOut.UTC().Format(time.RFC3339Nano)
	}

	if err := r.audit.InsertTx(ctx, tx, audit.Event{
		ActorUserID:  command.Actor.Subject,
		ActorEmail:   command.Actor.Email,
		ActorRole:    string(command.Actor.Role),
		Action:       audit.ActionAttendanceCorrected,
		EntityType:   audit.EntityAttendanceRecord,
		EntityID:     command.AttendanceID,
		TargetUserID: targetUserID,
		TargetLabel:  fmt.Sprintf("%s - %s", employeeNumber, employeeName),
		Reason:       command.Reason,
		BeforeData:   before,
		AfterData:    after,
	}); err != nil {
		return ErrInternal
	}

	if err := tx.Commit(ctx); err != nil {
		return ErrInternal
	}
	return nil
}
