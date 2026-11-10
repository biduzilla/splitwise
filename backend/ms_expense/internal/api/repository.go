package api

import (
	"database/sql"
	"log/slog"
	"ms_expense/internal/core/middleware"
	"ms_expense/internal/features/expenses"
)

type repositories struct {
	expense *expenses.ExpenseRepository
	store   middleware.IdempotencyStore
}

func NewRepositories(
	db *sql.DB,
	logger *slog.Logger,
) *repositories {
	return &repositories{
		expense: expenses.NewRepository(db, logger),
		store:   middleware.NewPostgresIdempotencyStore(db, logger),
	}
}
