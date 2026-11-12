package api

import (
	"ms_auth/internal/core/security"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type dependencies struct {
	routers *Router
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
	tx := transaction.NewManager(app.db)
	cacheClient, err := cache.NewRedisCache(app.config.Base.Cache.Addr,
		app.config.Base.Cache.Password, app.config.Base.Cache.Db, nil)
	if err != nil {
		return nil, err
	}

	app.Logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(app.config)
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

	modules := NewModules(app.db, app.Logger, errHandler, jwtService, tx, cacheClient, mw)

	router := NewRouter(
		[]routes{modules.auth, modules.user},
		errHandler,
		mw,
	)

	return &dependencies{
		routers: router,
	}, nil
}
