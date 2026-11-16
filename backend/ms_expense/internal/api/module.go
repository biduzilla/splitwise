package api

import (
	"ms_expense/internal/features/expenses"
)

type modules struct {
	expense *expenses.Module
}

func NewModules(deps AppDependencies) *modules {
	expenseDeps := expenses.Dependencies{
		Infra: expenses.InfraDeps{
			DB:     deps.Infra.DB,
			Logger: deps.Infra.Logger,
			Tx:     deps.Infra.Tx,
			Cache:  deps.Infra.Cache,
		},
		HTTP: expenses.HTTPDeps{
			ErrHandler:  deps.HTTP.ErrHandler,
			AuthMW:      deps.HTTP.AuthMW,
			IdemMW:      deps.HTTP.IdemMW,
			GroupClient: deps.HTTP.Clients.group,
		},
		Messaging: expenses.MessagingDeps{
			Producer: deps.Messaging.Producer.expenses,
		},
	}

	return &modules{
		expense: expenses.NewModule(expenseDeps),
	}
}
