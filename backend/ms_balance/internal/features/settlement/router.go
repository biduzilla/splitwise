package settlement

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type settlementRouter struct {
	handler     handler
	authMw      authMiddleware
	idempotency idempotencyMiddleware
}

type handler interface {
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type idempotencyMiddleware interface {
	Wrap(next http.Handler) http.Handler
}

func newRouter(
	handler handler,
	authMw authMiddleware,
	idempotency idempotencyMiddleware,
) *settlementRouter {
	return &settlementRouter{
		handler:     handler,
		authMw:      authMw,
		idempotency: idempotency,
	}
}

func (r *settlementRouter) Register(mux chi.Router) {
	mux.Route("/settlements", func(rt chi.Router) {
		rt.Use(r.authMw.RequireActivatedUser)

		rt.With(r.idempotency.Wrap).Post("/", r.handler.Create)
		rt.Get("/", r.handler.List)
		rt.Delete("/{id}", r.handler.Delete)
	})
}

func (m *Module) RegisterRoutes(r chi.Router) { m.router.Register(r) }
