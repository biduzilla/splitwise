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

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler errorHandler
	AuthMW     authMiddleware
}

type Dependencies struct {
	Infra InfraDeps
	HTTP  HTTPDeps
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)
	kb := cache.NewKeyBuilder("user")

	service := newService(
		repo,
		deps.Infra.Cache,
		kb,
		transaction.NewWriterExecutor(deps.Infra.Tx, deps.Infra.Cache, kb),
	)
	handler := newHandler(service, deps.HTTP.ErrHandler)
	route := newRouter(handler, deps.HTTP.AuthMW)

	return &Module{
		Service: service,
		router:  route,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
