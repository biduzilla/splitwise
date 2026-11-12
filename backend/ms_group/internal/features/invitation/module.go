package invitation

import (
	"database/sql"
	"log/slog"
	"ms_group/internal/core/config"
	"shared/cache"
	"shared/transaction"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service *invitationService
	router  *invitationRouter
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
	config config.Config,
) (*Module, error) {
	repo := newRepository(db, logger)

	kb := cache.NewKeyBuilder("invitation")
	we := transaction.NewWriterExecutor(tx, c, kb)
	tokenService, err := newTokenService(config.Base.Security.PrivateKeyPath, config.Base.Security.PublicKeyPath)
	if err != nil {
		return nil, err
	}
	service := newService(repo, c, we, kb, producer, logger, membershipService, tokenService, config.InviteURLBase)
	handler := newHandler(service, errHandler)
	router := newRouter(handler, authMiddleware)

	return &Module{
		Service: service,
		router:  router,
	}, nil
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
func (m *Module) MountInGroup(r chi.Router)   { m.router.MountInGroup(r) }
