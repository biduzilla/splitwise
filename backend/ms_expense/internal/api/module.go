package api

import (
	"database/sql"
	"log/slog"
	m "ms_expense/internal/core/middleware"
	"ms_expense/internal/features/expenses"
	"shared/apierror"
	"shared/cache"
	"shared/obs/middleware"
	"shared/transaction"
)

type modules struct {
	expense *expenses.Module
}

func NewModules(
	db *sql.DB,
	logger *slog.Logger,
	tx transaction.Manager,
	clients *clients,
	producer *producer,
	c cache.Cache,
	errHandler *apierror.ErrorHandler,
	mw *middleware.Middleware,
	idemMw *m.Idempotency,
) *modules {
	return &modules{
		expense: expenses.NewModule(
			db,
			logger,
			tx,
			clients.group,
			producer.expenses,
			c,
			errHandler,
			mw,
			idemMw,
		),
	}
}
