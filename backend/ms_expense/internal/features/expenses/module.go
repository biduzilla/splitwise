package expenses

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	service *expenseService
	router  *expenseRouter
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	tx transaction.Manager,
	groupClient groupClient,
	producer producer,
	c cache.Cache,
	errHandler errHandler,
	authMw authMiddleware,
	idemMw idemMw,
) *Module {
	repo := newRepository(db, logger)

	kb := cache.NewKeyBuilder("expense")

	service := newService(
		repo,
		transaction.NewWriterExecutor(tx, c, kb),
		groupClient,
		producer,
		kb,
		c,
		logger,
	)
	handler := newHandler(service, errHandler)
	router := newRouter(handler, authMw, idemMw)

	return &Module{
		service: service,
		router:  router,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
func (m *Module) Service() Service            { return m.service }
