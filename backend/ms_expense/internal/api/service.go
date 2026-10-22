package api

import (
	"log/slog"
	"ms_auth/internal/core/config"
	"ms_expense/internal/core/security"
	"shared/cache"
	"shared/transaction"
)

type services struct {
	jwtService *security.JwtService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
) (*services, error) {
	cacheClient, err := cache.NewRedisCache(config.Base.Cache.Addr,
		config.Base.Cache.Password, config.Base.Cache.Db, nil)

	if err != nil {
		return nil, err
	}

	logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	return &services{}, nil
}
