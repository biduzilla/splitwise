package api

import (
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

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
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
	modules := NewModules(
		app.db,
		app.Logger,
		tx,
		clients,
		producer,
		cacheClient,
		errHandler,
		mw,
		idemMw,
	)

	router := NewRouter(
		[]routes{modules.expense},
		errHandler,
		mw,
	)

	return &dependencies{
		routers: router,
	}, nil
}
