package invitation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authMiddleware interface {
	RequireActivatedUser(next http.Handler) http.Handler
}

type InvitationRouter struct {
	handler        handler
	authMiddleware authMiddleware
}

func NewRouter(
	handler handler,
	authMiddleware authMiddleware,
) *InvitationRouter {
	return &InvitationRouter{
		handler:        handler,
		authMiddleware: authMiddleware,
	}
}

func (r *InvitationRouter) MountInGroup(router chi.Router) {
	router.Post("/invitations", r.handler.Create)
}

func (r *InvitationRouter) Routes(router chi.Router) {
	router.Route("/invitations", func(router chi.Router) {
		router.Use(r.authMiddleware.RequireActivatedUser)
		router.Post("/{token}/accept", r.handler.Accept)
	})
}
