package api

import (
	"database/sql"
	"log/slog"
	m "ms_expense/internal/core/middleware"
	"shared/apierror"
	"shared/auth/security"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type dependencies struct {
	routers *Router
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler *apierror.ErrorHandler
	AuthMW     *middleware.Middleware
	IdemMW     *m.Idempotency
	Clients    *clients
}

type MessagingDeps struct {
	Producer *producer
}

type AppDependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
	cacheClient, err := cache.NewRedisCache(
		app.config.Base.Cache.Addr,
		app.config.Base.Cache.Password,
		app.config.Base.Cache.Db, nil,
	)
	if err != nil {
		return nil, err
	}
	app.Logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(app.config.Base)
	if err != nil {
		return nil, err
	}

	errHandler := apierror.NewErrorHandler(app.Logger)
	mw := middleware.New(
		errHandler,
		app.config.Base,
		jwtService,
		app.Logger,
		shutdown,
	)

	store := m.NewPostgresIdempotencyStore(app.db, app.Logger)
	idemMw := m.NewIdempotency(store, cacheClient, errHandler)

	deps := AppDependencies{
		Infra: InfraDeps{
			DB:     app.db,
			Logger: app.Logger,
			Tx:     transaction.NewManager(app.db),
			Cache:  cacheClient,
		},
		HTTP: HTTPDeps{
			ErrHandler: errHandler,
			AuthMW:     mw,
			IdemMW:     idemMw,
			Clients:    NewClients(app.config),
		},
		Messaging: MessagingDeps{
			Producer: NewProducer(app.kafkaProducer, app.Logger),
		},
	}

	mods := NewModules(deps)

	router := NewRouter(
		[]routes{mods.expense},
		deps.HTTP.ErrHandler,
		deps.HTTP.AuthMW,
	)

	return &dependencies{routers: router}, nil
}
