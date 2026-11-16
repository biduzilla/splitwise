package auth

import (
	"database/sql"
	"log/slog"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *authService
	router  *authRouter
}

type InfraDeps struct {
	DB     *sql.DB
	Logger *slog.Logger
	Tx     transaction.Manager
	Cache  cache.Cache
}

type HTTPDeps struct {
	ErrHandler errorHandler
}

type Domain struct {
	Jwt   jwtService
	Users userService
}

type Dependencies struct {
	Infra  InfraDeps
	HTTP   HTTPDeps
	Domain Domain
}

func NewModule(deps Dependencies) *Module {
	repo := newRepository(deps.Infra.DB, deps.Infra.Logger)
	service := newService(deps.Domain.Jwt, deps.Domain.Users, repo, deps.Infra.Tx)
	handler := newHandler(service, deps.HTTP.ErrHandler)
	route := newRouter(handler)

	return &Module{
		Service: service,
		router:  route,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
