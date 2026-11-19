package balance

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type balanceRouter struct {
	handler handler
	authMw  authMiddleware
}

type handler interface {
	GetBalances(w http.ResponseWriter, r *http.Request)
	GetSuggested(w http.ResponseWriter, r *http.Request)
}

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

func newRouter(handler handler, authMw authMiddleware) *balanceRouter {
	return &balanceRouter{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *balanceRouter) Register(mux chi.Router) {
	mux.Route("/groups/{id}", func(rt chi.Router) {
		rt.Use(r.authMw.RequireActivatedUser)

		rt.Get("/balances", r.handler.GetBalances)
		rt.Get("/suggested", r.handler.GetSuggested)
	})
}
