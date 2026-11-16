package membership

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *membershipService
	router  *membershipRouter
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Cache  cache.Cache
	Tx     transaction.Manager
}

type HTTPDeps struct {
	ErrHandler errHandler
	UserClient userClient
}

type MessagingDeps struct {
	Producer producer
}

type DomainDeps struct {
	GroupFinder groupFinder
}

type Dependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
	Domain    DomainDeps
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)

	kb := cache.NewKeyBuilder("membership")
	we := transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb)

	service := newService(
		repo,
		deps.Infra.Cache,
		we,
		kb,
		deps.Messaging.Producer,
		deps.Infra.Logger,
		deps.Domain.GroupFinder,
		deps.HTTP.UserClient,
	)

	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler)

	return &Module{
		Service: service,
		router:  router,
	}
}

func (m *Module) MountInGroup(r chi.Router) { m.router.MountInGroup(r) }
