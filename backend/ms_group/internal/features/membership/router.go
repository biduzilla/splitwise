package membership

import (
	"github.com/go-chi/chi/v5"
)

type MembershipRouter struct {
	handler handler
}

func NewRouter(handler handler) *MembershipRouter {
	return &MembershipRouter{handler: handler}
}

func (r *MembershipRouter) MountInGroup(router chi.Router) {
	router.Delete("/members/{userId}", r.handler.Remove)
}
