package api

import (
	"ms_balance/internal/features/balance"
	"ms_balance/internal/features/settlement"
)

type modules struct {
	settlement *settlement.Module
	balance    *balance.Module
}

func newModules(deps AppDependencies) *modules {
	settlement := settlement.NewModule(
		settlement.Dependencies{
			Infra: settlement.InfraDeps{
				DB:     deps.Infra.DB,
				Logger: deps.Infra.Logger,
				Tx:     deps.Infra.Tx,
				Cache:  deps.Infra.Cache,
			},
			Messaging: settlement.MessagingDeps{
				Producer: deps.Messaging.Producer.settlement,
			},
		},
	)

	balance := balance.NewModule(
		balance.Dependencies{
			Infra: balance.InfraDeps{
				DB:     deps.Infra.DB,
				Logger: deps.Infra.Logger,
				Tx:     deps.Infra.Tx,
				Cache:  deps.Infra.Cache,
			},
			Domain: balance.DomainDeps{
				SettlementService: settlement.Service,
			},
			HTTP: balance.HTTPDeps{
				ErrHandler:    deps.HTTP.ErrHandler,
				AuthMW:        deps.HTTP.AuthMW,
				GroupClient:   deps.HTTP.Clients.group,
				ExpenseClient: deps.HTTP.Clients.expense,
			},
		},
	)
	return &modules{
		settlement: settlement,
		balance:    balance,
	}
}
