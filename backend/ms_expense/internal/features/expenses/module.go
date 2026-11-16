package expenses

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *expenseService
	router  *expenseRouter
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler  errHandler
	AuthMW      authMiddleware
	IdemMW      idemMw
	GroupClient groupClient
}

type MessagingDeps struct {
	Producer producer
}

type Dependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)
	kb := cache.NewKeyBuilder("expense")

	service := newService(
		repo,
		transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb),
		deps.HTTP.GroupClient,
		deps.Messaging.Producer,
		deps.Infra.Cache,
		deps.Infra.Logger,
	)

	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler, deps.HTTP.AuthMW, deps.HTTP.IdemMW)

	return &Module{
		Service: service,
		router:  router,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
