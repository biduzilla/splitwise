package api

import (
	"database/sql"
	"expvar"
	"fmt"
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
	"net/http"
	"shared/apierror"
	"shared/obs/metrics"

	"github.com/go-chi/chi/v5"
)

type mw interface {
	Metrics(next http.Handler) http.Handler
	EnableCORS(next http.Handler) http.Handler
	RequireAuthenticatedUser(next http.Handler) http.Handler
	RequireActivatedUser(next http.Handler) http.Handler
	Authenticate(next http.Handler) http.Handler
	RateLimit(next http.Handler) http.Handler
	RecoverPanic(next http.Handler) http.Handler
	Logging(next http.Handler) http.Handler
	TimeoutMiddleWare(next http.Handler) http.Handler
	RequestID(next http.Handler) http.Handler
}

type customMiddleware interface {
	RequireGroupMember(next http.Handler) http.Handler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type Router struct {
	errHandler       errorHandler
	m                mw
	group            *group.GroupRouter
	invitation       *invitation.InvitationRouter
	customMiddleware customMiddleware
}

func NewRouter(
	handlers *handlers,
	errHandler errorHandler,
	m mw,
	customMiddleware customMiddleware,
) *Router {
	membershipRouter := membership.NewRouter(handlers.membership)
	invitationRouter := invitation.NewRouter(handlers.invitation, m)
	return &Router{
		m:          m,
		errHandler: errHandler,
		group: group.NewRouter(
			handlers.group,
			m,
			customMiddleware,
			membershipRouter,
			invitationRouter,
		),
		invitation: invitationRouter,
	}
}

func (router *Router) RegisterRoutes(db *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(router.m.RecoverPanic)
	r.Use(router.m.TimeoutMiddleWare)
	r.Use(router.m.RequestID)
	r.Use(router.m.Metrics)
	r.Use(router.m.Logging)

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		router.errHandler.HandlerError(w, req, apierror.ErrRecordNotFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		message := fmt.Sprintf("the %s method is not supported for this resource", req.Method)
		router.errHandler.HandlerError(w, req, apierror.NewHTTPError(message, http.StatusMethodNotAllowed, nil))
	})

	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Handle("/metrics", metrics.MetricsHandler(db))
	r.Mount("/debug/vars", expvar.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(router.m.RateLimit)
		r.Use(router.m.EnableCORS)
		r.Use(router.m.Authenticate)

		router.group.Routes(r)
		router.invitation.Routes(r)
	})

	return r
}
