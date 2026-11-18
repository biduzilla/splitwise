package settlement

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/db/sqlformat"
	"strconv"
	"strings"
	"uuid"
)

type settlementRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func newRepository(db *sql.DB, logger *slog.Logger) *settlementRepo {
	return &settlementRepo{db: db, logger: logger.WithGroup("db")}
}

func (r *settlementRepo) Insert(
	ctx context.Context,
	m *Settlement,
) error {
	q := `
		INSERT INTO settlements (
			group_id, 
			from_user_id, 
			to_user_id, 
			amount, 
			created_by
		)
		VALUES (?)
		RETURNING id, version, settled_at`

	userID := contexts.GetUser(ctx).GetID()
	row, err := queryRow(ctx, q, r.logger,
		m.GroupID, m.FromUserID, m.ToUserID, m.Amount, userID,
	)
	if err != nil {
		return err
	}
	return row.Scan(&m.ID, &m.Version, &m.SettledAt)
}

func (r *settlementRepo) SoftDelete(
	ctx context.Context,
	id uuid.UUID,
) error {
	q := `
		UPDATE settlements 
		SET 
			deleted = true, 
			version = version + 1, 
			updated_by = $1 
		WHERE 
			id = $2 
			AND deleted = false`
	userID := contexts.GetUser(ctx).GetID()

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}

	res, err := tx.ExecContext(ctx, q, userID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierror.ErrRecordNotFound
	}
	return nil
}

func (r *settlementRepo) FindByGroupID(
	ctx context.Context,
	groupID uuid.UUID,
) ([]*Settlement, error) {
	q := `
		SELECT 
			id, 
			group_id, 
			from_user_id, 
			to_user_id, 
			amount, 
			settled_at, 
			version
		FROM settlements 
		WHERE 
			group_id = $1 
			AND deleted = false`
	rows, err := r.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Settlement
	for rows.Next() {
		var s Settlement
		if err := rows.Scan(&s.ID, &s.GroupID, &s.FromUserID, &s.ToUserID, &s.Amount, &s.SettledAt, &s.Version); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *settlementRepo) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Settlement, error) {
	q := `
		SELECT
    		id,
    		group_id,
    		from_user_id,
    		to_user_id,
    		amount,
    		settled_at,
    		version,
		FROM settlements
		WHERE
    		id = $1
    		AND deleted = false`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	var s Settlement
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&s.ID, &s.GroupID, &s.FromUserID, &s.ToUserID,
		&s.Amount, &s.SettledAt, &s.Version, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierror.ErrRecordNotFound
		}
		return nil, err
	}
	return &s, nil
}

func queryRow(ctx context.Context, tpl string, logger *slog.Logger, params ...any) (*sql.Row, error) {
	parts := make([]string, len(params))
	for i := range params {
		parts[i] = "$" + strconv.Itoa(i+1)
	}

	query := strings.Replace(tpl, "?", strings.Join(parts, ", "), 1)

	logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	return tx.QueryRowContext(ctx, query, params...), nil
}
