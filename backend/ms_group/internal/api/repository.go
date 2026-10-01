package api

import (
	"database/sql"
	"log/slog"
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
)

type repositories struct {
	group      *group.GroupRepository
	invitation *invitation.InvitationRepository
	membership *membership.MembershipRepository
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		group:      group.NewRepository(db, logger),
		invitation: invitation.NewRepository(db, logger),
		membership: membership.NewRepository(db, logger),
	}
}
