package api

import (
	"ms_group/internal/core/config"
	"ms_group/internal/core/httpclient/user"
)

type clients struct {
	user *user.HTTPClient
}

func NewClients(
	cfg config.Config,
) *clients {
	return &clients{
		user: user.NewClient(
			cfg.Clients.AuthURL,
			cfg.Base.Server.Timeout,
		),
	}
}
