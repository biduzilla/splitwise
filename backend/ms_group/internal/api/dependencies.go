package api

import (
	m "ms_group/internal/core/middleware"
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
	cacheClient, err := cache.NewRedisCache(app.config.Base.Cache.Addr,
		app.config.Base.Cache.Password, app.config.Base.Cache.Db, nil)

	if err != nil {
		return nil, err
	}

	app.Logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(app.config.Base)
	if err != nil {
		return nil, err
	}
	tx := transaction.NewManager(app.db)
	clients := NewClients(app.config)
	errHandler := apierror.NewErrorHandler(app.Logger)
	mw := middleware.New(
		errHandler,
		app.config.Base,
		jwtService,
		app.Logger,
		shutdown,
	)

	customMiddleware := m.New(errHandler, nil)

	modules, err := NewModule(
		app.db,
		app.Logger,
		cacheClient,
		tx,
		app.kafkaProducer,
		errHandler,
		*clients,
		mw,
		app.config,
		customMiddleware,
	)
	if err != nil {
		return nil, err
	}

	customMiddleware.SetGroupService(modules.group.Service)

	router := NewRouter(
		errHandler,
		mw,
		modules.group,
		modules.invitation,
	)

	return &dependencies{
		routers: router,
	}, nil
}
