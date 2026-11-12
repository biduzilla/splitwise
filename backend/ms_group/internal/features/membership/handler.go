package membership

import (
	"net/http"
	"shared/apierror"
	"shared/httpx/httputils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type errHandler interface {
	HandlerError(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	)
}

type membershipHandler struct {
	svc        service
	errHandler errHandler
}

type handler interface {
	Remove(w http.ResponseWriter, r *http.Request)
}

func newHandler(svc service, errHandler errHandler) *membershipHandler {
	return &membershipHandler{svc: svc, errHandler: errHandler}
}

// DELETE /v1/groups/{id}/members/{userId}
func (h *membershipHandler) Remove(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/membership")
	ctx, span := tracer.Start(r.Context(), "membershipHandler.Remove")
	defer span.End()

	groupID, err := httputils.PathParamUUID(r, "id")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid group id")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(err.Error(), http.StatusBadRequest, err))
		return
	}

	userID, err := httputils.PathParamUUID(r, "userId")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid user id")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(err.Error(), http.StatusBadRequest, err))
		return
	}

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
	)

	if err := h.svc.Remove(ctx, groupID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to remove member")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Member removed")
	httputils.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
