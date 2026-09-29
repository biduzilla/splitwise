package group

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/db/sqlformat"
	"shared/filters"
	"uuid"
)

type GroupRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *GroupRepository {
	return &GroupRepository{db: db, logger: logger.WithGroup("db")}
}

type repository interface {
	Insert(ctx context.Context, model *Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*Group, error)
	FindAllByUser(ctx context.Context, f filters.Filters) ([]*Group, filters.Metadata, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

func (r *GroupRepository) Insert(
	ctx context.Context,
	g *Group,
) error {

	userAuth := contexts.GetUser(ctx)

	query := `
        INSERT INTO grp_groups (name, currency, owner_id, created_by)
        VALUES (:name, 'BRL', :owner_id, :created_by)
        RETURNING id, version, created_at
    `
	params := map[string]any{
		"name":       g.Name,
		"owner_id":   g.OwnerID,
		"created_by": userAuth.GetID(),
	}

	query, args := sqlformat.NamedQuery(query, params)

	r.logger.Info("insert group", "sql", query)

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}

	return tx.QueryRowContext(ctx, query, args...).
		Scan(&g.ID, &g.Version, &g.CreatedAt)
}

func (r *GroupRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Group, error) {
	queryGroup := `
	select 
		id,
		name,
		currcency,
		owner_id,
		version, 
		created_at, 
		created_by, 
		updated_at, 
		updated_by, 
		deleted
	from grp_groups
	where id = $1 and deleted = false`

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(queryGroup))

	var g Group
	err := r.db.QueryRowContext(ctx, queryGroup, id).
		Scan(
			&g.ID,
			&g.Name,
			&g.Currency,
			&g.OwnerID,
			&g.Version,
			&g.CreatedAt,
			&g.CreatedBy,
			&g.UpdatedAt,
			&g.UpdatedBy,
			&g.Deleted,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierror.ErrRecordNotFound
	}

	return &g, err
}

func (r *GroupRepository) FindAllByUser(
	ctx context.Context, f filters.Filters,
) ([]*Group, filters.Metadata, error) {
	userAuth := contexts.GetUser(ctx)
	query := `
        SELECT
            count(*) OVER(),
            g.id, 
			g.name,
			g.currency,
			g.owner_id,
            g.version, 
			g.created_at, 
			g.created_by, 
			g.updated_at,
			g.updated_by,
			g.deleted
        FROM grp_groups g
        INNER JOIN grp_memberships m ON m.group_id = g.id
        WHERE g.deleted = false
          AND m.user_id = $1
          AND m.deleted = false
        ORDER BY g.created_at DESC
        LIMIT $2 OFFSET $3
    `

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	rows, err := r.db.QueryContext(ctx, query, userAuth.GetID(), f.Limit(), f.Offset())
	if err != nil {
		return nil, filters.Metadata{}, err
	}
	defer rows.Close()

	var groups []*Group
	total := 0
	for rows.Next() {
		var g Group
		if err := rows.Scan(
			&total,
			&g.ID,
			&g.Name,
			&g.Currency,
			&g.OwnerID,
			&g.Version,
			&g.CreatedAt,
			&g.CreatedBy,
			&g.UpdatedAt,
			&g.UpdatedBy,
			&g.Deleted,
		); err != nil {
			return nil, filters.Metadata{}, err
		}
		groups = append(groups, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, filters.Metadata{}, err
	}

	meta := filters.CalculateMetadata(total, f.Page, f.PageSize)
	return groups, meta, nil
}

func (r *GroupRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	userAuth := contexts.GetUser(ctx)

	query := `
        UPDATE grp_groups
        SET 
			deleted = true, 
			updated_at = now(), 
			updated_by = :user_id, 
			version = version + 1
        WHERE id = :id AND deleted = false
    `
	params := map[string]any{"id": id, "user_id": userAuth.GetID()}
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))
	query, args := sqlformat.NamedQuery(query, params)

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierror.ErrRecordNotFound
	}
	return nil
}
