package api

import (
	"ms_expense/internal/features/expenses"
	"shared/apierror"
)

type handlers struct {
	expenses *expenses.ExpenseHandler
}

func NewHandlers(
	services *services,
	errHandler *apierror.ErrorHandler,
) *handlers {
	return &handlers{
		expenses: expenses.NewHandler(services.expenses, errHandler),
	}
}
