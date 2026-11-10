package api

import (
	"ms_expense/internal/core/config"
	"ms_expense/internal/core/http_client/groups"
)

type clients struct {
	group *groups.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		group: groups.NewClient(
			cfg.Clients.GroupURL,
			cfg.Base.Server.Timeout,
		),
	}
}
