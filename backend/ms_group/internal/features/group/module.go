package group

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *groupService
	router  *groupRouter
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	c cache.Cache,
	tx transaction.Manager,
	producer producer,
	membershipService membershipService,
	errHandler errHandler,
	authMiddleware authMiddleware,
	m middleware,
	subs ...GroupSubRouter,
) *Module {
	repo := newRepository(db, logger)

	kb := cache.NewKeyBuilder("group")
	we := transaction.NewWriterExecutor(tx, c, kb)
	service := newService(repo, c, we, kb, producer, logger, membershipService)
	handler := newHandler(service, errHandler)
	router := newRouter(handler, authMiddleware, m, subs...)

	return &Module{
		Service: service,
		router:  router,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
