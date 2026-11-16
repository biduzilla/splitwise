package api

import (
	"database/sql"
	"log/slog"
	"ms_auth/internal/core/security"
	"shared/apierror"
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
}

type ServicesDeps struct {
	jwtService *security.JwtService
}

type AppDependencies struct {
	Infra    InfraDeps
	HTTP     HTTPDeps
	Services ServicesDeps
}

func (app *application) buildDependencies(shutdown chan struct{}) (*dependencies, error) {
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
		},
		Services: ServicesDeps{
			jwtService: jwtService,
		},
	}

	modules := NewModules(deps)

	router := NewRouter(
		[]routes{modules.auth, modules.user},
		errHandler,
		mw,
	)

	return &dependencies{
		routers: router,
	}, nil
}
