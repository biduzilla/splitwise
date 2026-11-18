package api

import (
	"ms_balance/internal/core/config"
	"ms_balance/internal/core/http_client/expenses"
	"ms_balance/internal/core/http_client/groups"
)

type clients struct {
	group   *groups.HTTPClient
	expense *expenses.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		group: groups.NewClient(
			cfg.Clients.GroupURL,
			cfg.Base.Server.Timeout,
		),
		expense: expenses.NewClient(
			cfg.Clients.ExpenseURL,
			cfg.Base.Server.Timeout,
		),
	}
}
