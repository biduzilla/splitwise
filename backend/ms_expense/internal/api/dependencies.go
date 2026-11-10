package api

import (
	m "ms_expense/internal/core/middleware"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type dependencies struct {
	repo         *repositories
	tx           transaction.Manager
	services     *services
	errorHandler *apierror.ErrorHandler
	handlers     *handlers
	mw           *middleware.Middleware
	routers      *Router
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
	repo := NewRepositories(app.db, app.Logger)
	tx := transaction.NewManager(app.db)
	clients := NewClients(app.config)
	cacheClient, err := cache.NewRedisCache(
		app.config.Base.Cache.Addr,
		app.config.Base.Cache.Password,
		app.config.Base.Cache.Db, nil,
	)

	if err != nil {
		return nil, err
	}

	app.Logger.Info("reddis connection pool established")

	producer := NewProducer(app.kafkaProducer, app.Logger)
	services, err := NewServices(repo, tx, app.config, app.Logger, producer, clients, cacheClient)
	if err != nil {
		return nil, err
	}

	errHandler := apierror.NewErrorHandler(app.Logger)
	handlers := NewHandlers(services, errHandler)
	middleware := middleware.New(
		errHandler,
		app.config.Base,
		services.jwtService,
		app.Logger,
		shutdown,
	)

	idemMw := m.NewIdempotency(repo.store, cacheClient, errHandler)
	router := NewRouter(
		handlers,
		errHandler,
		middleware,
		idemMw,
	)

	return &dependencies{
		repo:         repo,
		tx:           tx,
		services:     services,
		errorHandler: errHandler,
		handlers:     handlers,
		mw:           middleware,
		routers:      router,
	}, nil
}
