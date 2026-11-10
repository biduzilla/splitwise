package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"shared/db/sqlformat"
	"uuid"
)

type postgresIdempotencyStore struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPostgresIdempotencyStore(db *sql.DB, logger *slog.Logger) IdempotencyStore {
	return &postgresIdempotencyStore{
		db:     db,
		logger: logger.WithGroup("idempotency"),
	}
}

func (s *postgresIdempotencyStore) Find(
	ctx context.Context,
	key uuid.UUID,
) (*StoredResponse, error) {
	query := `
        SELECT response_status, response_body
        FROM exp.idempotency_keys
        WHERE key = $1
    `
	s.logger.Info("query executed", "sql", sqlformat.MinifySQL(query), "key", key)

	var resp StoredResponse
	err := s.db.QueryRowContext(ctx, query, key).Scan(&resp.Status, &resp.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *postgresIdempotencyStore) Save(
	ctx context.Context,
	key, userID uuid.UUID,
	endpoint string,
	status int,
	body []byte,
) error {
	query := `
        INSERT INTO exp.idempotency_keys
            (key, user_id, endpoint, response_status, response_body)
        VALUES ($1, $2, $3, $4, $5::jsonb)
        ON CONFLICT (key) DO NOTHING
    `
	s.logger.Info("query executed", "sql", sqlformat.MinifySQL(query), "key", key)

	payload := "null"
	if len(body) > 0 {
		payload = string(body)
	}

	_, err := s.db.ExecContext(ctx, query,
		key, userID, endpoint, status, payload,
	)
	return err
}
