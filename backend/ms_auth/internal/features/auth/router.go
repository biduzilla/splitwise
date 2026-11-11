package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type authRouter struct {
	handler handler
}

type handler interface {
	Authenticate(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Session(w http.ResponseWriter, r *http.Request)
}

func newRouter(handler handler) *authRouter {
	return &authRouter{
		handler: handler,
	}
}

func (r *authRouter) Register(router chi.Router) {
	router.Route("/auth", func(router chi.Router) {
		router.Post("/", r.handler.Authenticate)
		router.Post("/refresh", r.handler.RefreshToken)
		router.Post("/logout", r.handler.Logout)
		router.Get("/session", r.handler.Session)
	})
}
