package api

import (
	"log/slog"
	"ms_expense/internal/core/config"
	"ms_expense/internal/features/expenses"

	"shared/auth/security"
	"shared/cache"
	"shared/transaction"
)

type services struct {
	jwtService *security.JwtService
	expenses   *expenses.ExpenseService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
	producer *producer,
	clients *clients,
	cacheClient cache.Cache,
) (*services, error) {
	jwtService, err := security.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	expenseD := newWriteDeps("expense", tx, cacheClient)
	expenseService := expenses.NewService(
		r.expense,
		expenseD.executor,
		clients.group,
		producer.expenses,
		expenseD.keyBuilder,
		expenseD.cache,
		logger,
	)

	return &services{
		jwtService: jwtService,
		expenses:   expenseService,
	}, nil
}

type writeDeps struct {
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	executor   *transaction.WriteExecutor
}

func newWriteDeps(prefix string, tx transaction.Manager, cacheClient cache.Cache) writeDeps {
	kb := cache.NewKeyBuilder(prefix)
	we := transaction.NewWriterExecutor(tx, cacheClient, kb)
	return writeDeps{
		cache:      cacheClient,
		keyBuilder: kb,
		executor:   we,
	}
}
