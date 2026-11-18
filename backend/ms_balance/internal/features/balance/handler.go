package balance

import (
	"context"
	"net/http"
	"shared/httpx/httputils"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type balanceHandler struct {
	svc        service
	errHandler errorHandler
}

type service interface {
	GetSuggestedTransfers(
		ctx context.Context,
		groupID uuid.UUID,
	) ([]Transfer, error)

	GetBalances(
		ctx context.Context,
		groupID uuid.UUID,
	) (map[uuid.UUID]int64, error)
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func newHandler(svc service, errHandler errorHandler) *balanceHandler {
	return &balanceHandler{svc: svc, errHandler: errHandler}
}

func (h *balanceHandler) GetBalances(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_balance/internal/features/balance")
	ctx, span := tracer.Start(r.Context(), "balanceHandler.GetBalances")
	defer span.End()

	groupID, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		return
	}

	span.SetAttributes(attribute.String("group.id", groupID.String()))

	balances, err := h.svc.GetBalances(ctx, groupID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get balances")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	// NÃO RETORNA NADA, PROBLEMA
	httputils.Respond(w, r, http.StatusOK, balances, nil, nil)
}

func (h *balanceHandler) GetSuggested(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_balance/internal/features/balance")
	ctx, span := tracer.Start(r.Context(), "balanceHandler.GetSuggested")
	defer span.End()

	groupID, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		return
	}

	span.SetAttributes(attribute.String("group.id", groupID.String()))

	transfers, err := h.svc.GetSuggestedTransfers(ctx, groupID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get balances")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	// NÃO RETORNA NADA, PROBLEMA
	httputils.Respond(w, r, http.StatusOK, map[string]any{"transfers": transfers}, nil, nil)
}
