package group

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type middleware interface {
	RequireGroupMember(next http.Handler) http.Handler
}

type GroupSubRouter interface {
	MountInGroup(router chi.Router)
}

type groupRouter struct {
	handler        handler
	authMiddleware authMiddleware
	m              middleware
	subs           []GroupSubRouter
}

func newRouter(
	handler handler,
	authMiddleware authMiddleware,
	m middleware,
	subs ...GroupSubRouter,
) *groupRouter {
	return &groupRouter{
		handler:        handler,
		authMiddleware: authMiddleware,
		m:              m,
		subs:           subs,
	}
}

func (r *groupRouter) Register(router chi.Router) {
	router.Route("/groups", func(router chi.Router) {
		router.Use(r.authMiddleware.RequireActivatedUser)

		router.Post("/", r.handler.Create)
		router.Get("/", r.handler.FindAll)

		router.Route("/{id}", func(router chi.Router) {
			router.Use(r.m.RequireGroupMember)
			router.Get("/", r.handler.FindByID)
			router.Delete("/", r.handler.Delete)

			for _, sub := range r.subs {
				sub.MountInGroup(router)
			}
		})
	})
}
