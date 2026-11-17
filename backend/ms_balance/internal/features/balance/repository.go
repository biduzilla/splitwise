package balance

import (
	"context"
	"database/sql"
	"log/slog"
	"shared/db/sqlformat"
	"uuid"
)

type balanceRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func (r *balanceRepository) GetByGroupID(
	ctx context.Context,
	groupID uuid.UUID,
) (map[uuid.UUID]int64, error) {
	q := `SELECT 
		user_id, 
		balance 
	FROM group_balances 
	WHERE group_id = $1`

	r.logger.Info(sqlformat.MinifySQL(q))

	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[uuid.UUID]int64)
	for rows.Next() {
		var userID uuid.UUID
		var balance int64
		if err := rows.Scan(&userID, &balance); err != nil {
			return nil, err
		}
		res[userID] = balance
	}
	return res, rows.Err()
}

func (r *balanceRepository) UpsertBatch(
	ctx context.Context,
	groupID uuid.UUID,
	balances map[uuid.UUID]int64,
) error {
	q := `
		INSERT INTO group_balances (
			group_id, 
			user_id, 
			balance, 
			updated_at
		)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (group_id, user_id) 
		DO UPDATE SET 
			balance = EXCLUDED.balance, 
			updated_at = now()`

	r.logger.Info(sqlformat.MinifySQL(q))

	for userID, balance := range balances {
		if _, err := r.db.ExecContext(
			ctx,
			q,
			groupID,
			userID,
			balance,
		); err != nil {
			return err
		}
	}
	return nil
}
