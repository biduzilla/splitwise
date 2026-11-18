package balance

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *balanceService
	router  *balanceRouter
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type DomainDeps struct {
	SettlementService settlementService
}

type HTTPDeps struct {
	ErrHandler    errorHandler
	AuthMW        authMiddleware
	GroupClient   groupClient
	ExpenseClient expenseClient
}

type Dependencies struct {
	Infra  InfraDeps
	Domain DomainDeps
	HTTP   HTTPDeps
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)
	kb := cache.NewKeyBuilder("balance")
	service := newService(
		repo,
		deps.Domain.SettlementService,
		deps.HTTP.GroupClient,
		deps.HTTP.ExpenseClient,
		kb,
		deps.Infra.Cache,
		transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb),
	)
	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler, deps.HTTP.AuthMW)

	return &Module{
		router:  router,
		Service: service,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
