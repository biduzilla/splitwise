package api

import (
	"ms_group/internal/core/config"
	"ms_group/internal/core/httpclient/auth"
)

type clients struct {
	auth *auth.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		auth: auth.NewClient(
			cfg.Clients.AuthURL,
			cfg.Base.Server.Timeout,
		),
	}
}
