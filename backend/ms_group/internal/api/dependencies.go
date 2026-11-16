package api

import (
	"messaging"
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
	customMiddleware := m.New(errHandler, nil)
	producersSvc := messaging.NewProducer(app.kafkaProducer, app.Logger)

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
			CustomMW:   customMiddleware,
			Clients:    *clients,
		},
		Messaging: MessagingDeps{
			Producer: producersSvc,
		},
		Config: app.config,
	}

	mods, err := NewModules(appDeps, app.kafkaProducer)
	if err != nil {
		return nil, err
	}

	customMiddleware.SetGroupService(mods.group.Service)

	router := NewRouter(
		errHandler,
		mw,
		mods.group,
		mods.invitation,
	)

	return &dependencies{
		routers: router,
	}, nil
}
