package api

import (
	m "ms_group/internal/core/middleware"
	"shared/apierror"
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
	services, err := NewServices(repo, tx, app.config, app.Logger)
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

	customMiddleware := m.New(errHandler, services.group)

	router := NewRouter(
		handlers,
		errHandler,
		middleware,
		customMiddleware,
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
