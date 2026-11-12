package api

import (
	"database/sql"
	"log/slog"
	"messaging"
	"ms_group/internal/core/config"
	m "ms_group/internal/core/middleware"
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"

	"github.com/IBM/sarama"
)

type modules struct {
	group      *group.Module
	invitation *invitation.Module
	membership *membership.Module
}

func NewModule(
	db *sql.DB,
	logger *slog.Logger,
	c cache.Cache,
	tx transaction.Manager,
	producer sarama.SyncProducer,
	errHandler *apierror.ErrorHandler,
	clients clients,
	mw *middleware.Middleware,
	config config.Config,
	customMiddleware *m.Middleware,
) (*modules, error) {
	producersSvc := messaging.NewProducer(producer, logger)

	membership := membership.NewModule(
		db,
		logger,
		c,
		tx,
		producersSvc,
		errHandler,
		nil,
		clients.user,
	)

	invitation, err := invitation.NewModule(
		db,
		logger,
		c,
		tx,
		producersSvc,
		membership.Service,
		errHandler,
		mw,
		config,
	)
	if err != nil {
		return nil, err
	}

	group := group.NewModule(
		db,
		logger,
		c,
		tx,
		producersSvc,
		membership.Service,
		errHandler,
		mw,
		customMiddleware,
		membership, invitation,
	)

	return &modules{
		group:      group,
		invitation: invitation,
		membership: membership,
	}, nil
}
