package invitation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type invitationRouter struct {
	handler        handler
	authMiddleware authMiddleware
}

func newRouter(
	handler handler,
	authMiddleware authMiddleware,
) *invitationRouter {
	return &invitationRouter{
		handler:        handler,
		authMiddleware: authMiddleware,
	}
}

func (r *invitationRouter) MountInGroup(router chi.Router) {
	router.Post("/invitations", r.handler.Create)
}

func (r *invitationRouter) Register(router chi.Router) {
	router.Route("/invitations", func(router chi.Router) {
		router.Use(r.authMiddleware.RequireActivatedUser)
		router.Post("/{token}/accept", r.handler.Accept)
	})
}
