package middleware

import (
	"context"
	"fmt"
	c "ms_group/internal/core/contexts"
	"net/http"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/httpx/httputils"
	"uuid"
)

type errorHandler interface {
	HandlerError(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	)
}

type groupService interface {
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
}

type Middleware struct {
	errHandler   errorHandler
	groupService groupService
}

func New(
	errHandler errorHandler,
	groupService groupService,
) *Middleware {
	return &Middleware{
		errHandler:   errHandler,
		groupService: groupService,
	}
}

func (m *Middleware) RequireGroupMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := contexts.GetUser(r.Context()).GetID()
		if userID == uuid.Nil() {
			m.errHandler.HandlerError(w, r, apierror.ErrAuthenticationRequired)
			return
		}

		groupID, err := httputils.PathParamUUID(r, "id")
		if err != nil {
			m.errHandler.HandlerError(w, r, apierror.ErrRecordNotFound)
			return
		}

		ok, err := m.groupService.IsMember(r.Context(), groupID, userID)
		if err != nil {
			m.errHandler.HandlerError(w, r, err)
			return
		}

		if !ok {
			m.errHandler.HandlerError(w, r, fmt.Errorf("Not a member"))
			return
		}

		ctx := c.SetGroupID(r.Context(), groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
