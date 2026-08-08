package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) ListEmployeeFaceEnrollments(ctx context.Context, userIDs []string) (map[string]EmployeeFaceEnrollment, error) {
	result := make(map[string]EmployeeFaceEnrollment, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	const query = `
		SELECT user_id::text, status, embedding_model, embedding_version, enrolled_at
		FROM face_profiles
		WHERE user_id::text = ANY($1::text[])
	`
	rows, err := r.pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, sanitizePostgresError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID           string
			status           string
			embeddingModel   string
			embeddingVersion string
			enrolledAt       pgtype.Timestamptz
		)
		if err := rows.Scan(
			&userID,
			&status,
			&embeddingModel,
			&embeddingVersion,
			&enrolledAt,
		); err != nil {
			return nil, sanitizePostgresError(err)
		}
		item := EmployeeFaceEnrollment{
			Enrolled:         status == "ENROLLED",
			FaceStatus:       status,
			EmbeddingModel:   embeddingModel,
			EmbeddingVersion: embeddingVersion,
		}
		if enrolledAt.Valid {
			value := enrolledAt.Time
			item.EnrolledAt = &value
		}
		result[userID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, sanitizePostgresError(err)
	}
	return result, nil
}
