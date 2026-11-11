package auth

import (
	"database/sql"
	"log/slog"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *authService
	router  *authRouter
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	jwt jwtService,
	users userService,
	tx transaction.Manager,
	errHandler errorHandler,
) *Module {
	repo := newRepository(db, logger)
	service := newService(jwt, users, repo, tx)
	handler := newHandler(service, errHandler)
	route := newRouter(handler)

	return &Module{
		Service: service,
		router:  route,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
