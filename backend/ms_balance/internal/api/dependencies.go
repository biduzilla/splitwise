package api

import (
	"database/sql"
	"log/slog"
	"ms_balance/internal/core/middleware"
	"shared/apierror"
	"shared/auth/security"
	"shared/cache"
	obsMiddleware "shared/obs/middleware"
	"shared/transaction"
)

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler  *apierror.ErrorHandler
	AuthMW      *obsMiddleware.Middleware
	Clients     *clients
	Idempotency *middleware.Idempotency
}

type MessagingDeps struct {
	Producer *producer
}

type AppDependencies struct {
	Infra     InfraDeps
	HTTP      HTTPDeps
	Messaging MessagingDeps
}

type dependencies struct {
	routers *Router
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, func(), error) {
	cacheClient, err := cache.NewRedisCache(
		app.config.Base.Cache.Addr,
		app.config.Base.Cache.Password,
		app.config.Base.Cache.Db, nil,
	)
	if err != nil {
		return nil, nil, err
	}
	app.Logger.Info("redis connection pool established")

	jwtService, err := security.NewService(app.config.Base)
	if err != nil {
		return nil, nil, err
	}

	tx := transaction.NewManager(app.db)
	clients := NewClients(app.config)
	errHandler := apierror.NewErrorHandler(app.Logger)
	mw := obsMiddleware.New(errHandler, app.config.Base, jwtService, app.Logger, shutdown)
	producers := newProducer(app.kafkaProducer, app.Logger)

	idempotencyRepo := middleware.NewIdempotencyRepo(app.db, app.Logger)
	idempotencyMW := middleware.NewIdempotency(idempotencyRepo, cacheClient, errHandler)

	appDeps := AppDependencies{
		Infra: InfraDeps{
			DB:     app.db,
			Logger: app.Logger,
			Cache:  cacheClient,
			Tx:     tx,
		},
		HTTP: HTTPDeps{
			ErrHandler:  errHandler,
			AuthMW:      mw,
			Clients:     clients,
			Idempotency: idempotencyMW,
		},
		Messaging: MessagingDeps{
			Producer: producers,
		},
	}

	mods := newModules(appDeps)
	consumers := NewConsumer(app.kafkaConsumer, mods, cacheClient, app.Logger)
	shutdownConsumers := consumers.Start(app.Logger)
	router := NewRouter(
		errHandler,
		mw,
		mods.balance,
		mods.settlement,
	)

	return &dependencies{
		routers: router,
	}, shutdownConsumers, nil
}
