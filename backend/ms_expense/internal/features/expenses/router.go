package expenses

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type expenseRouter struct {
	handler handler
	authMw  authMiddleware
	idemMw  idemMw
}

func newRouter(
	handler handler,
	authMw authMiddleware,
	idemMw idemMw,
) *expenseRouter {
	return &expenseRouter{
		handler: handler,
		authMw:  authMw,
		idemMw:  idemMw,
	}
}

type idemMw interface {
	Wrap(next http.Handler) http.Handler
}

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

func (r *expenseRouter) Register(mux chi.Router) {
	mux.Route("/expenses", func(rt chi.Router) {
		rt.Use(r.authMw.RequireActivatedUser)

		rt.Get("/", r.handler.List)
		rt.With(r.idemMw.Wrap).Post("/", r.handler.Create)
		rt.Get("/{id}", r.handler.Get)
		rt.Put("/{id}", r.handler.Update)
		rt.Delete("/{id}", r.handler.Delete)
	})
}
