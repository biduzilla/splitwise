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

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger.WithGroup("db")}
}

type repository interface {
	Insert(ctx context.Context, model *Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*GroupDetail, error)
	FindAllByUser(ctx context.Context, userID uuid.UUID, f filters.Filters) ([]*Group, filters.Metadata, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error

	InsertMembership(ctx context.Context, m *Membership) error
	FindMembership(ctx context.Context, groupID, userID uuid.UUID) (*Membership, error)
	RemoveMembership(ctx context.Context, groupID, userID uuid.UUID) error
	CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error)

	InsertInvitation(ctx context.Context, inv *Invitation) error
	FindInvitationByJTI(ctx context.Context, jti uuid.UUID) (*Invitation, error)
	MarkInvitationUsed(ctx context.Context, id, userID uuid.UUID) error
}

func (r *Repository) Insert(
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

func (r *Repository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*GroupDetail, error) {
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
	if err != nil {
		return nil, err
	}

	queryMembers := `
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

	rows, err := r.db.QueryContext(ctx, queryMembers, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []Membership
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

		members = append(members, m)
	}

	return &GroupDetail{Group: g, Members: members}, rows.Err()
}

func (r *Repository) FindAllByUser(
	ctx context.Context, userID uuid.UUID, f filters.Filters,
) ([]*Group, filters.Metadata, error) {
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

	rows, err := r.db.QueryContext(ctx, query, userID, f.Limit(), f.Offset())
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

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
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

func (r *Repository) InsertMembership(
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

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	return tx.QueryRowContext(ctx, query, args...).
		Scan(&m.ID, &m.Version, &m.CreatedAt)
}

func (r *Repository) FindMembership(
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) RemoveMembership(
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

func (r *Repository) CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT 
			count(*) 
		FROM grp_memberships 
		WHERE group_id = $1 AND deleted = false`,
		groupID,
	).Scan(&n)
	return n, err
}

func (r *Repository) InsertInvitation(ctx context.Context, inv *Invitation) error {
	userAuth := contexts.GetUser(ctx)

	query := `
        INSERT INTO grp_invitations (
			group_id, 
			inviter_id, 
			jti, 
			expires_at, 
			created_by
		)
        VALUES (
			:group_id, 
			:inviter_id, 
			:jti, 
			:expires_at, 
			:created_by
		)
        RETURNING id, version, created_at
    `
	params := map[string]any{
		"group_id":   inv.GroupID,
		"inviter_id": inv.InviterID,
		"jti":        inv.JTI,
		"expires_at": inv.ExpiresAt,
		"created_by": userAuth.GetID(),
	}
	query, args := sqlformat.NamedQuery(query, params)

	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	return tx.QueryRowContext(ctx, query, args...).Scan(&inv.ID, &inv.Version, &inv.CreatedAt)
}

func (r *Repository) FindInvitationByJTI(ctx context.Context, jti uuid.UUID) (*Invitation, error) {
	query := `
        SELECT 
			id, 
			group_id, 
			inviter_id, 
			jti, 
			expires_at, 
			used_at, 
			used_by,
            version, 
			created_at, 
			created_by, 
			updated_at, 
			updated_by, 
			deleted
        FROM grp_invitations
        WHERE jti = $1 AND deleted = false
    `
	var inv Invitation
	err := r.db.QueryRowContext(ctx, query, jti).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.JTI,
		&inv.ExpiresAt,
		&inv.UsedAt,
		&inv.UsedBy,
		&inv.Version,
		&inv.CreatedAt,
		&inv.CreatedBy,
		&inv.UpdatedAt,
		&inv.UpdatedBy,
		&inv.Deleted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierror.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *Repository) MarkInvitationUsed(ctx context.Context, id, userID uuid.UUID) error {
	query := `
        UPDATE grp_invitations
        SET 
			used_at = now(), 
			used_by = $2,
            updated_at = now(), 
			updated_by = $2,
			version = version + 1
        WHERE id = $1 AND used_at IS NULL
    `
	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	res, err := tx.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierror.NewHTTPError("invitation already used", 409, nil)
	}
	return nil
}
