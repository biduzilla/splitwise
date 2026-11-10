package expenses

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ExpenseRouter struct {
	handler handler
	authMw  authMiddleware
	idemMw  idemMw
}

func NewRouter(
	handler handler,
	authMw authMiddleware,
	idemMw idemMw,
) *ExpenseRouter {
	return &ExpenseRouter{
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

func (r *ExpenseRouter) Register(mux chi.Router) {
	mux.Route("/v1/expenses", func(rt chi.Router) {
		rt.Use(r.authMw.RequireActivatedUser)

		rt.Get("/", r.handler.List)
		rt.With(r.idemMw.Wrap).Post("/", r.handler.Create)
		rt.Get("/{id}", r.handler.Get)
		rt.Put("/{id}", r.handler.Update)
		rt.Delete("/{id}", r.handler.Delete)
	})
}
