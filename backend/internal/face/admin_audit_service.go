package face

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"r3-ti-faceattend/backend/internal/audit"
	"r3-ti-faceattend/backend/internal/auth"
	"r3-ti-faceattend/backend/internal/user"
)

type AuditedAdminService struct {
	pool  *pgxpool.Pool
	audit *audit.PostgresRepository
}

func NewAuditedAdminService(pool *pgxpool.Pool, auditRepo *audit.PostgresRepository) AuditedAdminService {
	return AuditedAdminService{pool: pool, audit: auditRepo}
}

func (s AuditedAdminService) AdminReset(ctx context.Context, claims auth.Claims, targetUserID string, reasons ...string) error {
	if strings.TrimSpace(claims.Subject) == "" || claims.Role != user.RoleAdmin {
		return ErrForbidden
	}

	targetUserID = strings.TrimSpace(targetUserID)
	var parsedID pgtype.UUID
	if err := parsedID.Scan(targetUserID); err != nil || !parsedID.Valid {
		return ErrInvalidInput
	}

	reason := ""
	if len(reasons) > 0 {
		reason = strings.TrimSpace(reasons[0])
	}
	if len(reason) < 5 || len(reason) > 1000 {
		return ErrInvalidInput
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrRepositoryFailure
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const userQuery = `
		SELECT role, employee_number, name
		FROM users
		WHERE id = $1::uuid
		FOR UPDATE
	`
	var role user.Role
	var employeeNumber, employeeName string
	if err := tx.QueryRow(ctx, userQuery, targetUserID).Scan(&role, &employeeNumber, &employeeName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		return ErrRepositoryFailure
	}
	if role != user.RoleUser {
		return ErrProfileNotFound
	}

	const deleteQuery = `DELETE FROM face_profiles WHERE user_id = $1::uuid RETURNING id::text`
	var profileID string
	if err := tx.QueryRow(ctx, deleteQuery, targetUserID).Scan(&profileID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProfileNotFound
		}
		return ErrRepositoryFailure
	}

	if err := s.audit.InsertTx(ctx, tx, audit.Event{
		ActorUserID:  claims.Subject,
		ActorEmail:   claims.Email,
		ActorRole:    string(claims.Role),
		Action:       audit.ActionFaceEnrollmentReset,
		EntityType:   audit.EntityFaceProfile,
		EntityID:     profileID,
		TargetUserID: targetUserID,
		TargetLabel:  fmt.Sprintf("%s - %s", employeeNumber, employeeName),
		Reason:       reason,
		BeforeData:   map[string]any{"status": FaceStatusEnrolled},
		AfterData:    map[string]any{"status": FaceStatusNotEnrolled},
	}); err != nil {
		return ErrRepositoryFailure
	}

	if err := tx.Commit(ctx); err != nil {
		return ErrRepositoryFailure
	}
	return nil
}
