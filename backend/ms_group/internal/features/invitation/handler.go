package invitation

import (
	"net/http"
	"shared/apierror"
	"shared/httpx/httputils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type invitationHandler struct {
	svc        service
	errHandler errHandler
}

type errHandler interface {
	HandlerError(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	)
}

func newHandler(svc service, errHandler errHandler) *invitationHandler {
	return &invitationHandler{svc: svc, errHandler: errHandler}
}

type handler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Accept(w http.ResponseWriter, r *http.Request)
}

// POST /v1/groups/{id}/invitations
func (h *invitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/invitation")
	ctx, span := tracer.Start(r.Context(), "invitationHandler.Create")
	defer span.End()

	groupID, err := httputils.PathParamUUID(r, "id")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid group id")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(err.Error(), http.StatusBadRequest, err))
		return
	}

	span.SetAttributes(attribute.String("group.id", groupID.String()))

	inv, err := h.svc.Create(ctx, groupID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create invitation")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Invitation created")
	httputils.Respond(w, r, http.StatusCreated, inv, nil, h.errHandler)
}

// POST /v1/invitations/{token}/accept
func (h *invitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/invitation")
	ctx, span := tracer.Start(r.Context(), "invitationHandler.Accept")
	defer span.End()

	token, err := httputils.PathParamString(r, "token")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Missing token")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(err.Error(), http.StatusBadRequest, err))
		return
	}

	// Nunca logar o token inteiro — só o começo, pra debug.
	span.SetAttributes(attribute.String("invitation.token_prefix", token[:min(8, len(token))]))

	m, err := h.svc.Accept(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to accept invitation")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetAttributes(
		attribute.String("group.id", m.GroupID.String()),
		attribute.String("user.id", m.UserID.String()),
	)
	span.SetStatus(codes.Ok, "Invitation accepted")

	httputils.Respond(w, r, http.StatusOK, m.ToDTO(), nil, h.errHandler)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
