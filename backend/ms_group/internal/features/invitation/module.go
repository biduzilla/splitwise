package invitation

import (
	"database/sql"
	"log/slog"
	"ms_group/internal/core/config"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Cache  cache.Cache
	Tx     transaction.Manager
}

type HTTPDeps struct {
	ErrHandler     errHandler
	AuthMiddleware authMiddleware
}

type MessagingDeps struct {
	Producer producer
}

type DomainDeps struct {
	MembershipService membershipService
}

type Dependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
	Domain    DomainDeps
	Config    config.Config
}

type Module struct {
	Service *invitationService
	router  *invitationRouter
}

func NewModule(deps Dependencies) (*Module, error) {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)

	kb := cache.NewKeyBuilder("invitation")
	we := transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb)

	tokenService, err := newTokenService(
		deps.Config.Base.Security.PrivateKeyPath,
		deps.Config.Base.Security.PublicKeyPath,
	)
	if err != nil {
		return nil, err
	}

	service := newService(
		repo,
		deps.Infra.Cache,
		we,
		kb,
		deps.Messaging.Producer,
		deps.Infra.Logger,
		deps.Domain.MembershipService,
		tokenService,
		deps.Config.InviteURLBase,
	)

	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler, deps.HTTP.AuthMiddleware)

	return &Module{
		Service: service,
		router:  router,
	}, nil
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
func (m *Module) MountInGroup(r chi.Router)   { m.router.MountInGroup(r) }
