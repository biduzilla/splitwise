package api

import (
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
	"shared/apierror"
)

type handlers struct {
	group      *group.GroupHandler
	invitation *invitation.InvitationHandler
	membership *membership.MembershipHandler
}

func NewHandlers(
	services *services,
	errHandler *apierror.ErrorHandler,
) *handlers {
	return &handlers{
		group:      group.NewHandler(services.group, errHandler),
		invitation: invitation.NewHandler(services.invitation, errHandler),
		membership: membership.NewHandler(services.membership, errHandler),
	}
}
