package api

import (
	"database/sql"
	"log/slog"
	"messaging"
	"ms_group/internal/core/config"
	m "ms_group/internal/core/middleware"
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"

	"github.com/IBM/sarama"
)

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Cache  cache.Cache
	Tx     transaction.Manager
}

type HTTPDeps struct {
	ErrHandler *apierror.ErrorHandler
	AuthMW     *middleware.Middleware
	CustomMW   *m.Middleware
	Clients    clients
}

type MessagingDeps struct {
	Producer *messaging.Producer
}

type AppDependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
	Config    config.Config
}

type modules struct {
	group      *group.Module
	invitation *invitation.Module
	membership *membership.Module
}

func NewModules(deps AppDependencies, rawProducer sarama.SyncProducer) (*modules, error) {

	membershipMod := membership.NewModule(membership.Dependencies{
		Infra: membership.InfraDeps{
			DB: deps.Infra.DB, Logger: deps.Infra.Logger, Cache: deps.Infra.Cache, Tx: deps.Infra.Tx,
		},
		HTTP: membership.HTTPDeps{
			ErrHandler: deps.HTTP.ErrHandler, UserClient: deps.HTTP.Clients.user,
		},
		Messaging: membership.MessagingDeps{
			Producer: deps.Messaging.Producer,
		},
		Domain: membership.DomainDeps{
			GroupFinder: nil,
		},
	})

	invitationMod, err := invitation.NewModule(invitation.Dependencies{
		Infra: invitation.InfraDeps{
			DB: deps.Infra.DB, Logger: deps.Infra.Logger, Cache: deps.Infra.Cache, Tx: deps.Infra.Tx,
		},
		HTTP: invitation.HTTPDeps{
			ErrHandler: deps.HTTP.ErrHandler, AuthMiddleware: deps.HTTP.AuthMW,
		},
		Messaging: invitation.MessagingDeps{
			Producer: deps.Messaging.Producer,
		},
		Domain: invitation.DomainDeps{
			MembershipService: membershipMod.Service,
		},
		Config: deps.Config,
	})
	if err != nil {
		return nil, err
	}

	groupMod := group.NewModule(group.Dependencies{
		Infra: group.InfraDeps{
			DB: deps.Infra.DB, Logger: deps.Infra.Logger, Cache: deps.Infra.Cache, Tx: deps.Infra.Tx,
		},
		HTTP: group.HTTPDeps{
			ErrHandler: deps.HTTP.ErrHandler, AuthMiddleware: deps.HTTP.AuthMW, Middleware: deps.HTTP.CustomMW,
		},
		Messaging: group.MessagingDeps{
			Producer: deps.Messaging.Producer,
		},
		Domain: group.DomainDeps{
			MembershipService: membershipMod.Service,
			SubRouters:        []group.GroupSubRouter{membershipMod, invitationMod},
		},
	})

	membershipMod.Service.SetGroupFinder(groupMod.Service)

	return &modules{
		group:      groupMod,
		invitation: invitationMod,
		membership: membershipMod,
	}, nil
}
