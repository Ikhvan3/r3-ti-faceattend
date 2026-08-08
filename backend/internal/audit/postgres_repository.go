package audit

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) InsertTx(ctx context.Context, tx pgx.Tx, event Event) error {
	beforeJSON, err := json.Marshal(nonNilData(event.BeforeData))
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(nonNilData(event.AfterData))
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO audit_logs (
			actor_user_id, actor_email, actor_role, action, entity_type, entity_id,
			target_user_id, target_label, reason, before_data, after_data
		) VALUES (
			NULLIF($1, '')::uuid, $2, $3, $4, $5, NULLIF($6, '')::uuid,
			NULLIF($7, '')::uuid, NULLIF($8, ''), $9, $10::jsonb, $11::jsonb
		)
	`
	_, err = tx.Exec(ctx, query,
		event.ActorUserID,
		event.ActorEmail,
		event.ActorRole,
		string(event.Action),
		string(event.EntityType),
		event.EntityID,
		event.TargetUserID,
		event.TargetLabel,
		event.Reason,
		string(beforeJSON),
		string(afterJSON),
	)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, filter listQuery) ([]Log, error) {
	const query = `
		SELECT id::text, actor_user_id::text, actor_email, actor_role, action,
			entity_type, entity_id::text, target_user_id::text, target_label,
			reason, before_data, after_data, created_at
		FROM audit_logs
		WHERE ($1 = '' OR action = $1)
			AND ($2 = '' OR entity_type = $2)
			AND ($3::timestamptz IS NULL OR created_at >= $3)
			AND ($4::timestamptz IS NULL OR created_at < $4)
		ORDER BY created_at DESC, id DESC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.pool.Query(ctx, query,
		string(filter.Action),
		string(filter.EntityType),
		filter.DateFrom,
		filter.DateTo,
		filter.PageSize,
		(filter.Page-1)*filter.PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Log, 0)
	for rows.Next() {
		item, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) Count(ctx context.Context, filter listQuery) (int, error) {
	const query = `
		SELECT COUNT(*)::int
		FROM audit_logs
		WHERE ($1 = '' OR action = $1)
			AND ($2 = '' OR entity_type = $2)
			AND ($3::timestamptz IS NULL OR created_at >= $3)
			AND ($4::timestamptz IS NULL OR created_at < $4)
	`
	var count int
	if err := r.pool.QueryRow(ctx, query,
		string(filter.Action),
		string(filter.EntityType),
		filter.DateFrom,
		filter.DateTo,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

type logScanner interface {
	Scan(dest ...any) error
}

func scanLog(scanner logScanner) (Log, error) {
	var (
		actorUserID  pgtype.Text
		entityID     pgtype.Text
		targetUserID pgtype.Text
		targetLabel  pgtype.Text
		action       string
		entityType   string
		beforeJSON   []byte
		afterJSON    []byte
		item         Log
	)
	if err := scanner.Scan(
		&item.ID,
		&actorUserID,
		&item.ActorEmail,
		&item.ActorRole,
		&action,
		&entityType,
		&entityID,
		&targetUserID,
		&targetLabel,
		&item.Reason,
		&beforeJSON,
		&afterJSON,
		&item.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Log{}, err
		}
		return Log{}, err
	}
	item.Action = Action(action)
	item.EntityType = EntityType(entityType)
	if actorUserID.Valid {
		value := actorUserID.String
		item.ActorUserID = &value
	}
	if entityID.Valid {
		value := entityID.String
		item.EntityID = &value
	}
	if targetUserID.Valid {
		value := targetUserID.String
		item.TargetUserID = &value
	}
	if targetLabel.Valid {
		value := targetLabel.String
		item.TargetLabel = &value
	}
	if err := json.Unmarshal(beforeJSON, &item.BeforeData); err != nil {
		return Log{}, err
	}
	if err := json.Unmarshal(afterJSON, &item.AfterData); err != nil {
		return Log{}, err
	}
	return item, nil
}

func nonNilData(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}
