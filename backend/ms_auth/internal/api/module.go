package api

import (
	"ms_auth/internal/features/auth"
	"ms_auth/internal/features/user"
)

type modules struct {
	user *user.Module
	auth *auth.Module
}

func NewModules(deps AppDependencies) *modules {
	userMod := user.NewModule(user.Dependencies{
		Infra: user.InfraDeps{
			DB:     deps.Infra.DB,
			Logger: deps.Infra.Logger,
			Tx:     deps.Infra.Tx,
			Cache:  deps.Infra.Cache,
		},
		HTTP: user.HTTPDeps{
			ErrHandler: deps.HTTP.ErrHandler,
			AuthMW:     deps.HTTP.AuthMW,
		},
	})

	authMod := auth.NewModule(auth.Dependencies{
		Infra: auth.InfraDeps{
			DB:     deps.Infra.DB,
			Logger: deps.Infra.Logger,
			Tx:     deps.Infra.Tx,
			Cache:  deps.Infra.Cache,
		},
		HTTP: auth.HTTPDeps{
			ErrHandler: deps.HTTP.ErrHandler,
		},
		Domain: auth.Domain{
			Jwt:   deps.Services.jwtService,
			Users: userMod.Service,
		},
	})

	return &modules{
		userMod, authMod,
	}
}
