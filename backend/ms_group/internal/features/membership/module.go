package membership

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *membershipService
	router  *membershipRouter
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	c cache.Cache,
	tx transaction.Manager,
	producer producer,
	errHandler errHandler,
	groupFinder groupFinder,
	userClient userClient,
) *Module {
	repo := newRepository(db, logger)

	kb := cache.NewKeyBuilder("membership")
	we := transaction.NewWriterExecutor(tx, c, kb)
	service := newService(repo, c, we, kb, producer, logger, groupFinder, userClient)
	handler := newHandler(service, errHandler)
	router := newRouter(handler)

	return &Module{
		Service: service,
		router:  router,
	}
}

func (m *Module) MountInGroup(r chi.Router) { m.router.MountInGroup(r) }
