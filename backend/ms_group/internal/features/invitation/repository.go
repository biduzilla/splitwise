package invitation

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

type InvitationRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *InvitationRepository {
	return &InvitationRepository{db: db, logger: logger.WithGroup("db")}
}

type repository interface {
	Insert(ctx context.Context, inv *Invitation) error
	FindByJTI(ctx context.Context, jti uuid.UUID) (*Invitation, error)
	MarkUsed(ctx context.Context, id, userID uuid.UUID) error
}

func (r *InvitationRepository) Insert(ctx context.Context, inv *Invitation) error {
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
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))
	tx := contexts.GetTx(ctx)
	if tx == nil {
		return errors.New("transaction required")
	}
	return tx.QueryRowContext(ctx, query, args...).Scan(&inv.ID, &inv.Version, &inv.CreatedAt)
}

func (r *InvitationRepository) FindByJTI(ctx context.Context, jti uuid.UUID) (*Invitation, error) {
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

	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))

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

func (r *InvitationRepository) MarkUsed(ctx context.Context, id, userID uuid.UUID) error {
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
	r.logger.Info("query executed", "sql", sqlformat.MinifySQL(query))
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
