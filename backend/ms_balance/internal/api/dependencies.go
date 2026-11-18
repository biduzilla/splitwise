package api

import (
	"database/sql"
	"log/slog"
	"shared/apierror"
	"shared/auth/security"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler *apierror.ErrorHandler
	AuthMW     *middleware.Middleware
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

type dependencies struct {
	routers *Router
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
	app.Logger.Info("redis connection pool established")

	jwtService, err := security.NewService(app.config.Base)
	if err != nil {
		return nil, err
	}

	tx := transaction.NewManager(app.db)
	clients := NewClients(app.config)
	errHandler := apierror.NewErrorHandler(app.Logger)
	mw := middleware.New(errHandler, app.config.Base, jwtService, app.Logger, shutdown)
	producers := newProducer(app.kafkaProducer, app.Logger)

	appDeps := AppDependencies{
		Infra: InfraDeps{
			DB:     app.db,
			Logger: app.Logger,
			Cache:  cacheClient,
			Tx:     tx,
		},
		HTTP: HTTPDeps{
			ErrHandler: errHandler,
			AuthMW:     mw,
			Clients:    clients,
		},
		Messaging: MessagingDeps{
			Producer: producers,
		},
	}

	mods := newModules(appDeps)
	router := NewRouter(
		errHandler,
		mw,
		mods.balance,
	)

	return &dependencies{
		routers: router,
	}, nil
}
