package membership

import (
	"github.com/go-chi/chi/v5"
)

type membershipRouter struct {
	handler handler
}

func newRouter(handler handler) *membershipRouter {
	return &membershipRouter{handler: handler}
}

func (r *membershipRouter) MountInGroup(router chi.Router) {
	router.Delete("/members/{userId}", r.handler.Remove)
}
