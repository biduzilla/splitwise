package api

import (
	"database/sql"
	"log/slog"
	"ms_auth/internal/core/security"
	"ms_auth/internal/features/auth"
	"ms_auth/internal/features/user"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type modules struct {
	user *user.Module
	auth *auth.Module
}

func NewModules(
	db *sql.DB,
	logger *slog.Logger,
	errHandler *apierror.ErrorHandler,
	jwtService *security.JwtService,
	tx transaction.Manager,
	c cache.Cache,
	mw *middleware.Middleware,
) *modules {
	user := user.NewModule(db, logger, errHandler, tx, c, mw)
	auth := auth.NewModule(db, logger, jwtService, user.Service, tx, errHandler)

	return &modules{
		user, auth,
	}
}
