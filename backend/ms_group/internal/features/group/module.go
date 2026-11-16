package group

import (
	"database/sql"
	"log/slog"
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
	Middleware     middleware
}

type MessagingDeps struct {
	Producer producer
}

type DomainDeps struct {
	MembershipService membershipService
	SubRouters        []GroupSubRouter
}

type Dependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
	Domain    DomainDeps
}

type Module struct {
	Service *groupService
	router  *groupRouter
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)

	kb := cache.NewKeyBuilder("group")
	we := transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb)

	service := newService(
		repo,
		deps.Infra.Cache,
		we,
		kb,
		deps.Messaging.Producer,
		deps.Infra.Logger,
		deps.Domain.MembershipService,
	)

	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler, deps.HTTP.AuthMiddleware, deps.HTTP.Middleware, deps.Domain.SubRouters...)

	return &Module{
		Service: service,
		router:  router,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
