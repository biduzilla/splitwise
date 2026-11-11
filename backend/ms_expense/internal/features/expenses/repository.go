package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/db/sqlformat"
	"shared/filters"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/lib/pq"
)

type repository interface {
	Insert(ctx context.Context, e *Expense) error
	InsertShares(ctx context.Context, expenseID uuid.UUID, shares []Share) error
	FindByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	FindByGroupID(
		ctx context.Context,
		groupID uuid.UUID,
		startDate, endDate *time.Time,
		f filters.Filters,
	) ([]*Expense, filters.Metadata, error)
	Update(ctx context.Context, e *Expense) error
	ReplaceShares(ctx context.Context, expenseID uuid.UUID, shares []Share) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type expenseRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func newRepository(db *sql.DB, logger *slog.Logger) *expenseRepository {
	return &expenseRepository{
		db:     db,
		logger: logger.WithGroup("db"),
	}
}

func (r *expenseRepository) Insert(ctx context.Context, m *Expense) error {
	q := `
        INSERT INTO expenses(
			group_id, 
			paid_by, 
			amount, 
			description, 
			category, 
			split_mode,
            expense_date, 
			created_by
		)
        VALUES (?)
        RETURNING id, version, created_at`

	userAuth := contexts.GetUser(ctx)

	row, err := queryRow(ctx, q, r.logger,
		m.GroupID, m.PaidBy, m.Amount, m.Description,
		m.Category, m.SplitMode, m.ExpenseDate, userAuth.GetID(),
	)
	if err != nil {
		return err
	}
	return row.Scan(&m.ID, &m.Version, &m.CreatedAt)
}

func (r *expenseRepository) InsertShares(
	ctx context.Context,
	expenseID uuid.UUID,
	shares []Share,
) error {
	if len(shares) == 0 {
		return errors.New("expense must have at least one share")
	}

	q := `
	insert into expense_shares (expense_id, user_id, amount)
	select 
		$1, 
		u, 
		a
	FROM unnest($2::uuid[], $3::bigint[]) AS t(u, a)
	`
	userIDs := make([]string, len(shares))
	amounts := make([]int64, len(shares))
	for i, s := range shares {
		userIDs[i] = s.UserID.String()
		amounts[i] = s.Amount
	}

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}

	_, err := tx.ExecContext(ctx, q, expenseID, pq.Array(userIDs), pq.Array(amounts))
	return err
}

func (r *expenseRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Expense, error) {
	q := `
        SELECT 
			id, 
			group_id, 
			paid_by, 
			amount, 
			description, 
			category, 
			split_mode,
            expense_date, 
			version, 
			created_at, 
			created_by,
            updated_at, 
			updated_by, 
			deleted
        FROM expenses
        WHERE id = $1 AND deleted = false`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	var e Expense
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&e.ID,
		&e.GroupID,
		&e.PaidBy,
		&e.Amount,
		&e.Description,
		&e.Category,
		&e.SplitMode,
		&e.ExpenseDate,
		&e.Version,
		&e.CreatedAt,
		&e.CreatedBy,
		&e.UpdatedAt,
		&e.UpdatedBy,
		&e.Deleted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	shares, err := r.loadShares(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Shares = shares
	return &e, nil
}

func (r *expenseRepository) loadShares(
	ctx context.Context,
	expenseID uuid.UUID,
) ([]Share, error) {
	q := `
	select 
		user_id,
		amount
	from expense_shared
	where expense_id = $1
	`
	rows, err := r.db.QueryContext(ctx, q, expenseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []Share
	for rows.Next() {
		var s Share
		if err := rows.Scan(&s.UserID, &s.Amount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *expenseRepository) FindByGroupID(
	ctx context.Context,
	groupID uuid.UUID,
	startDate, endDate *time.Time,
	f filters.Filters,
) ([]*Expense, filters.Metadata, error) {
	var total int
	q_count := `
	select 
		count(*)
	from expenses 
	WHERE 
		group_id = $1 
		AND deleted = false
		AND ($2::timestamptz IS NULL OR created_at >= $2)
        AND ($3::timestamptz IS NULL OR created_at <= $3)
	`
	if err := r.db.QueryRowContext(ctx, q_count, groupID).Scan(&total); err != nil {
		return nil, filters.Metadata{}, err
	}

	q := fmt.Sprintf(`
        SELECT id, group_id, paid_by, amount, description, category, split_mode,
               expense_date, version, created_at, created_by,
               updated_at, updated_by, deleted
        FROM exp.expenses
        WHERE 
			group_id = $1 
			AND deleted = false
			AND ($2::timestamptz IS NULL OR created_at >= $2)
        	AND ($3::timestamptz IS NULL OR created_at <= $3)
        ORDER BY %s %s
        LIMIT $2 OFFSET $3`,
		f.SortColumn(), f.SortDirection())

	rows, err := r.db.QueryContext(ctx, q, groupID, f.Limit(), f.Offset())
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	var out []*Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.PaidBy,
			&e.Amount,
			&e.Description,
			&e.Category,
			&e.SplitMode,
			&e.ExpenseDate,
			&e.Version,
			&e.CreatedAt,
			&e.CreatedBy,
			&e.UpdatedAt,
			&e.UpdatedBy,
			&e.Deleted,
		); err != nil {
			return nil, filters.Metadata{}, err
		}
		out = append(out, &e)
	}

	metadata := filters.CalculateMetadata(total, f.Page, f.PageSize)
	return out, metadata, rows.Err()
}

func (r *expenseRepository) Update(
	ctx context.Context,
	e *Expense,
) error {
	userAuth := contexts.GetUser(ctx)
	q := `
	update expenses
	SET 
		paid_by = $1, 
		amount = $2, 
		description = $3, 
		category = $4,
        split_mode = $5, 
		expense_date = $6, 
		updated_by = $7,
        updated_at = now(), 
		version = version + 1
	WHERE 
		id = $8 
		AND version = $9 
		AND deleted = false
    RETURNING version
	`
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	err := r.db.QueryRowContext(
		ctx,
		q,
		e.PaidBy,
		e.Amount,
		e.Description,
		e.Category,
		e.SplitMode,
		e.ExpenseDate,
		userAuth.GetID(),
		e.ID,
		e.Version,
	).Scan(&e.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return apierror.ErrEditConflict
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *expenseRepository) ReplaceShares(
	ctx context.Context,
	expenseID uuid.UUID,
	shares []Share,
) error {
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM expense_shares WHERE expense_id = $1`, expenseID); err != nil {
		return err
	}
	return r.InsertShares(ctx, expenseID, shares)
}

func (r *expenseRepository) SoftDelete(
	ctx context.Context,
	id uuid.UUID,
) error {
	userID := contexts.GetUser(ctx).GetID()
	q := `
		UPDATE expenses
        SET 
		  	deleted = true, 
			updated_by = $1, 
			updated_at = now(), 
			version = version + 1
        WHERE 
			id = $2 
			AND deleted = false`
	res, err := r.db.ExecContext(ctx, q, userID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierror.ErrEditConflict
	}
	return nil
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
