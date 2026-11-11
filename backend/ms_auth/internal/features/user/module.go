package user

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *userService
	router  *userRouter
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	errHandler errorHandler,
	tx transaction.Manager,
	c cache.Cache,
	m authMiddleware,
) *Module {
	repo := newRepository(db, logger)
	kb := cache.NewKeyBuilder("user")
	we := transaction.NewWriterExecutor(tx, c, kb)
	service := newService(repo, c, kb, we)
	handler := newHandler(service, errHandler)
	route := newRouter(handler, m)

	return &Module{
		Service: service,
		router:  route,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
