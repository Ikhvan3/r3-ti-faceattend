package face

import (
	"context"
	"errors"
	"fmt"

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

func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string) (FaceProfile, error) {
	const query = `
		SELECT id, user_id, embedding, embedding_model, embedding_version,
			status, enrolled_at, created_at, updated_at
		FROM face_profiles
		WHERE user_id = $1
	`
	return scanProfile(r.pool.QueryRow(ctx, query, userID))
}

func (r *PostgresRepository) Create(ctx context.Context, profile FaceProfile) (FaceProfile, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}
	defer rollback(ctx, tx)

	const query = `
		INSERT INTO face_profiles (
			id, user_id, embedding, embedding_model, embedding_version,
			status, enrolled_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, user_id, embedding, embedding_model, embedding_version,
			status, enrolled_at, created_at, updated_at
	`
	created, err := scanProfile(tx.QueryRow(
		ctx,
		query,
		profile.ID,
		profile.UserID,
		profile.Embedding,
		profile.EmbeddingModel,
		profile.EmbeddingVersion,
		profile.Status,
		profile.EnrolledAt,
	))
	if err != nil {
		return FaceProfile{}, sanitizePostgresError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}
	return created, nil
}

func (r *PostgresRepository) DeleteByUserID(ctx context.Context, userID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ErrRepositoryFailure
	}
	defer rollback(ctx, tx)

	tag, err := tx.Exec(ctx, `DELETE FROM face_profiles WHERE user_id = $1`, userID)
	if err != nil {
		return sanitizePostgresError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProfileNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return ErrRepositoryFailure
	}
	return nil
}

func scanProfile(row pgx.Row) (FaceProfile, error) {
	var profile FaceProfile
	err := row.Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Embedding,
		&profile.EmbeddingModel,
		&profile.EmbeddingVersion,
		&profile.Status,
		&profile.EnrolledAt,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return FaceProfile{}, ErrProfileNotFound
	}
	if err != nil {
		return FaceProfile{}, sanitizePostgresError(err)
	}
	return profile, nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func sanitizePostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if pgErr.ConstraintName == "face_profiles_user_unique" {
				return ErrAlreadyEnrolled
			}
			return ErrAlreadyEnrolled
		case "23514", "23503":
			return ErrInvalidInput
		}
	}
	if errors.Is(err, ErrProfileNotFound) || errors.Is(err, ErrAlreadyEnrolled) || errors.Is(err, ErrInvalidInput) {
		return err
	}
	return fmt.Errorf("%w", ErrRepositoryFailure)
}
