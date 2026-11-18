package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"shared/auth/contexts"
	"shared/db/sqlformat"
	"uuid"
)

type idempotencyRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewIdempotencyRepo(db *sql.DB, logger *slog.Logger) IdempotencyStore {
	return &idempotencyRepo{
		db:     db,
		logger: logger.WithGroup("idempotency_repo"),
	}
}

func (r *idempotencyRepo) Find(
	ctx context.Context,
	key uuid.UUID,
) (*StoredResponse, error) {
	q := `
		SELECT response_status, response_body
		FROM idempotency_keys
		WHERE 
			key = $1
  			AND created_at > NOW() - INTERVAL '24 hours'`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	var resp StoredResponse
	var body []byte

	err := r.db.QueryRowContext(ctx, q, key).
		Scan(&resp.Status, &body)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	resp.Body = body
	return &resp, nil
}

func (r *idempotencyRepo) Save(
	ctx context.Context,
	key, userID uuid.UUID,
	endpoint string,
	status int,
	body []byte,
) error {
	q := `
		INSERT INTO bal.idempotency_keys (
    		key,
    		user_id,
    		endpoint,
    		response_status,
    		response_body
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO NOTHING`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}

	_, err := tx.ExecContext(ctx, q, key, userID, endpoint, status, body)
	return err
}
