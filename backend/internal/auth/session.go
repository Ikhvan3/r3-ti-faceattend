package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSessionNotFound = errors.New("auth session not found")

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastUsedAt       *time.Time
	CreatedIP        *string
	UserAgent        *string
	CreatedAt        time.Time
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session Session) error
	FindActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (Session, error)
	FindByID(ctx context.Context, id string) (Session, error)
	RotateRefreshToken(ctx context.Context, sessionID string, oldTokenHash string, newTokenHash string, now time.Time, expiresAt time.Time) error
	RevokeSession(ctx context.Context, tokenHash string, now time.Time) error
}

type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

func (r *PostgresSessionRepository) CreateSession(ctx context.Context, s Session) error {
	const query = `
		INSERT INTO auth_sessions (
			id, user_id, refresh_token_hash, expires_at, created_ip, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err := r.pool.Exec(ctx, query, s.ID, s.UserID, s.RefreshTokenHash, s.ExpiresAt, s.CreatedIP, s.UserAgent)
	if err != nil {
		return sanitizeSessionError(err)
	}

	return nil
}

func (r *PostgresSessionRepository) FindActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (Session, error) {
	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, last_used_at, created_ip, user_agent, created_at
		FROM auth_sessions
		WHERE refresh_token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > $2
	`

	return r.findOne(ctx, query, tokenHash, now)
}

func (r *PostgresSessionRepository) FindByID(ctx context.Context, id string) (Session, error) {
	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, last_used_at, created_ip, user_agent, created_at
		FROM auth_sessions
		WHERE id = $1
	`

	return r.findOne(ctx, query, id)
}

func (r *PostgresSessionRepository) RotateRefreshToken(ctx context.Context, sessionID string, oldTokenHash string, newTokenHash string, now time.Time, expiresAt time.Time) error {
	const query = `
		UPDATE auth_sessions
		SET refresh_token_hash = $1,
			last_used_at = $2,
			expires_at = $3
		WHERE id = $4
			AND refresh_token_hash = $5
			AND revoked_at IS NULL
			AND expires_at > $2
	`

	tag, err := r.pool.Exec(ctx, query, newTokenHash, now, expiresAt, sessionID, oldTokenHash)
	if err != nil {
		return sanitizeSessionError(err)
	}
	if tag.RowsAffected() != 1 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *PostgresSessionRepository) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE refresh_token_hash = $1
	`

	_, err := r.pool.Exec(ctx, query, tokenHash, now)
	if err != nil {
		return sanitizeSessionError(err)
	}

	return nil
}

func (r *PostgresSessionRepository) findOne(ctx context.Context, query string, args ...any) (Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.LastUsedAt,
		&s.CreatedIP,
		&s.UserAgent,
		&s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, sanitizeSessionError(err)
	}

	return s, nil
}

func sanitizeSessionError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return errors.New("auth session references unknown user")
		case "23505":
			return errors.New("auth session already exists")
		case "23514":
			return errors.New("auth session violates schema constraints")
		}
	}

	return fmt.Errorf("auth session repository operation failed")
}
