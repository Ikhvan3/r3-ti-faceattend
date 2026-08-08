package face

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

	created, err := insertProfile(ctx, tx, profile)
	if err != nil {
		return FaceProfile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}
	return created, nil
}

// CreateUnique serializes the relatively rare enrollment operation, uses the
// pgvector HNSW index to obtain a small candidate set, performs exact cosine
// comparison in Go, and inserts only when no other user crosses the duplicate
// threshold. Attendance verification remains a separate 1:1 path.
func (r *PostgresRepository) CreateUnique(ctx context.Context, profile FaceProfile, duplicateThreshold float64, searchTopK int) (FaceProfile, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}
	defer rollback(ctx, tx)

	// Enrollment is low-frequency compared with attendance. A transaction-level
	// advisory lock prevents two concurrent requests for the same biometric from
	// both passing the duplicate check before either insert becomes visible.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('r3-ti-faceattend:face-enrollment'))`); err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}

	// pgvector applies additional WHERE filters after an approximate index scan.
	// Iterative scans let HNSW continue searching when model/version filtering
	// removes candidates, improving recall without falling back to a full table
	// scan. This setting is transaction-local.
	if _, err := tx.Exec(ctx, `SET LOCAL hnsw.iterative_scan = strict_order`); err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}

	vectorValue := faceVectorLiteral(profile.Embedding)
	const candidateQuery = `
		SELECT embedding
		FROM face_profiles
		WHERE user_id <> $1
			AND status = 'ENROLLED'
			AND embedding_model = $2
			AND embedding_version = $3
		ORDER BY embedding_vector <=> $4::vector
		LIMIT $5
	`
	rows, err := tx.Query(
		ctx,
		candidateQuery,
		profile.UserID,
		profile.EmbeddingModel,
		profile.EmbeddingVersion,
		vectorValue,
		searchTopK,
	)
	if err != nil {
		return FaceProfile{}, sanitizePostgresError(err)
	}

	for rows.Next() {
		var existingEmbedding []float64
		if err := rows.Scan(&existingEmbedding); err != nil {
			rows.Close()
			return FaceProfile{}, sanitizePostgresError(err)
		}
		similarity, err := CosineSimilarity(profile.Embedding, existingEmbedding)
		if err != nil {
			rows.Close()
			return FaceProfile{}, ErrRepositoryFailure
		}
		if similarity >= duplicateThreshold {
			rows.Close()
			return FaceProfile{}, ErrDuplicateBiometric
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return FaceProfile{}, sanitizePostgresError(err)
	}
	rows.Close()

	created, err := insertProfile(ctx, tx, profile)
	if err != nil {
		return FaceProfile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FaceProfile{}, ErrRepositoryFailure
	}
	return created, nil
}

func insertProfile(ctx context.Context, tx pgx.Tx, profile FaceProfile) (FaceProfile, error) {
	const query = `
		INSERT INTO face_profiles (
			id, user_id, embedding, embedding_vector, embedding_model, embedding_version,
			status, enrolled_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4::vector, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, user_id, embedding, embedding_model, embedding_version,
			status, enrolled_at, created_at, updated_at
	`
	created, err := scanProfile(tx.QueryRow(
		ctx,
		query,
		profile.ID,
		profile.UserID,
		profile.Embedding,
		faceVectorLiteral(profile.Embedding),
		profile.EmbeddingModel,
		profile.EmbeddingVersion,
		profile.Status,
		profile.EnrolledAt,
	))
	if err != nil {
		return FaceProfile{}, sanitizePostgresError(err)
	}
	return created, nil
}

func faceVectorLiteral(values []float64) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.FormatFloat(value, 'g', -1, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
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

func (r *PostgresRepository) CreateVerificationGrant(ctx context.Context, grant VerificationGrant) error {
	const query = `
		INSERT INTO face_verification_grants (
			id, user_id, purpose, token_hash, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := r.pool.Exec(ctx, query, grant.ID, grant.UserID, grant.Purpose, grant.TokenHash, grant.ExpiresAt, grant.CreatedAt); err != nil {
		return sanitizePostgresError(err)
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
	if errors.Is(err, ErrProfileNotFound) || errors.Is(err, ErrAlreadyEnrolled) || errors.Is(err, ErrDuplicateBiometric) || errors.Is(err, ErrInvalidInput) {
		return err
	}
	return fmt.Errorf("%w", ErrRepositoryFailure)
}
