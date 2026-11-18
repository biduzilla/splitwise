package settlement

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"
)

type Module struct {
	Service *settlementService
	router  *settlementRouter
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type MessagingDeps struct {
	Producer producer
}

type HTTPDeps struct {
	ErrHandler    errorHandler
	AuthMW        authMiddleware
	IdempotencyMW idempotencyMiddleware
}

type Dependencies struct {
	Infra     InfraDeps
	Messaging MessagingDeps
	HTTP      HTTPDeps
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)
	kb := cache.NewKeyBuilder("settlement")
	service := newService(
		repo,
		transaction.NewWriterExecutor(
			deps.Infra.Tx,
			deps.Infra.Cache,
			kb,
		),
		deps.Messaging.Producer,
		deps.Infra.Cache,
		kb,
	)
	handler := newHandler(service, deps.HTTP.ErrHandler)
	router := newRouter(handler, deps.HTTP.AuthMW, deps.HTTP.IdempotencyMW)

	return &Module{
		Service: service,
		router:  router,
	}
}
