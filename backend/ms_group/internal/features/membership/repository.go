package membership

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/db/sqlformat"
	"uuid"
)

type MembershipRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *MembershipRepository {
	return &MembershipRepository{db: db, logger: logger.WithGroup("db")}
}

type repository interface {
	Insert(ctx context.Context, m *Membership) error
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Membership, error)
	FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*Membership, error)
	Remove(ctx context.Context, groupID, userID uuid.UUID) error
	CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error)
}

func (r *MembershipRepository) Insert(
	ctx context.Context,
	m *Membership,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
        INSERT INTO grp_memberships (
			group_id, 
			user_id, 
			role, 
			created_by
		)
        VALUES (:group_id, :user_id, :role, :created_by)
        RETURNING id, version, created_at
    `
	params := map[string]any{
		"group_id":   m.GroupID,
		"user_id":    m.UserID,
		"role":       m.Role,
		"created_by": userAuth.GetID(),
	}
	query, args := sqlformat.NamedQuery(query, params)
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))
	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	return tx.QueryRowContext(ctx, query, args...).
		Scan(&m.ID, &m.Version, &m.CreatedAt)
}

func (r *MembershipRepository) FindByGroupID(
	ctx context.Context,
	groupID uuid.UUID,
) ([]*Membership, error) {
	query := `
        SELECT 
			id, 
			group_id, 
			user_id, 
			role,
            version, 
			created_at, 
			created_by, 
			updated_at, 
			updated_by, 
			deleted
        FROM grp_memberships
        WHERE group_id = $1 AND deleted = false
        ORDER BY created_at ASC
    `

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []*Membership
	for rows.Next() {
		var m Membership
		if err := rows.Scan(
			&m.ID,
			&m.GroupID,
			&m.UserID,
			&m.Role,
			&m.Version,
			&m.CreatedAt,
			&m.CreatedBy,
			&m.UpdatedAt,
			&m.UpdatedBy,
			&m.Deleted,
		); err != nil {
			return nil, err
		}

		members = append(members, &m)
	}
	return members, rows.Err()
}

func (r *MembershipRepository) FindByGroupIDAndUserID(
	ctx context.Context,
	groupID, userID uuid.UUID,
) (*Membership, error) {
	query := `
        SELECT 
			id, 
			group_id, 
			user_id, 
			role,
            version, 
			created_at, 
			created_by, 
			updated_at, 
			updated_by, 
			deleted
        FROM grp_memberships
        WHERE 
			group_id = $1 
			AND user_id = $2 
			AND deleted = false
    `

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	var m Membership
	err := r.db.QueryRowContext(ctx, query, groupID, userID).
		Scan(
			&m.ID,
			&m.GroupID,
			&m.UserID,
			&m.Role,
			&m.Version,
			&m.CreatedAt,
			&m.CreatedBy,
			&m.UpdatedAt,
			&m.UpdatedBy,
			&m.Deleted,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierror.ErrRecordNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *MembershipRepository) Remove(
	ctx context.Context,
	groupID, userID uuid.UUID,
) error {
	userAuth := contexts.GetUser(ctx)

	query := `
        UPDATE grp_memberships
        SET 
			deleted = true, 
			updated_at = now(), 
			updated_by = $3, 
			version = version + 1
        WHERE 
			group_id = $1 
			AND user_id = $2 
			AND deleted = false
    `
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	res, err := tx.ExecContext(ctx, query, groupID, userID, userAuth.GetID())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierror.ErrRecordNotFound
	}
	return nil
}

func (r *MembershipRepository) CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error) {
	var n int
	q := `SELECT 
			count(*) 
		FROM grp_memberships 
		WHERE group_id = $1 AND deleted = false`

	err := r.db.QueryRowContext(ctx,
		q,
		groupID,
	).Scan(&n)

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(q))

	return n, err
}
