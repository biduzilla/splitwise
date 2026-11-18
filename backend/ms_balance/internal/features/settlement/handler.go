package settlement

import (
	"context"
	"net/http"

	"shared/apierror"
	"shared/filters"
	"shared/httpx/httpjson"
	"shared/httpx/httputils"
	"shared/validator"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type settlementHandler struct {
	svc        service
	errHandler errorHandler
}

type service interface {
	Create(ctx context.Context, m *Settlement) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Settlement, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Settlement, error)
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func newHandler(svc service, errHandler errorHandler) *settlementHandler {
	return &settlementHandler{svc: svc, errHandler: errHandler}
}

// Create ─────────────────────────────────────────────────────────────────
// POST /v1/settlements
//
// Body: CreateSettlementDTO. Retorna 201 + SettlementDTO.
// O header Idempotency-Key deve ser tratado por middleware dedicado
// (ver router.go) — o handler assume payload já deduplicado.
func (h *settlementHandler) Create(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_balance/internal/features/settlement")
	ctx, span := tracer.Start(r.Context(), "settlementHandler.Create")
	defer span.End()

	var dto CreateSettlementDTO
	if err := httpjson.ReadJSON(w, r, &dto); err != nil {
		h.errHandler.HandlerError(w, r, apierror.NewBadRequestError(err))
		return
	}

	v := validator.New()
	dto.Validate(v)
	if !v.Valid() {
		h.errHandler.HandlerError(w, r, apierror.NewValidationError(v.Errors))
		return
	}

	span.SetAttributes(
		attribute.String("group.id", dto.GroupID.String()),
		attribute.String("settlement.from_user_id", dto.FromUserID.String()),
		attribute.String("settlement.to_user_id", dto.ToUserID.String()),
		attribute.Int64("settlement.amount", dto.Amount),
	)

	model := dto.ToModel()
	if err := h.svc.Create(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create settlement")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Settlement created")
	httputils.Respond(w, r, http.StatusCreated, model.ToDTO(), nil, nil)
}

// List ───────────────────────────────────────────────────────────────────
// GET /v1/settlements?group_id=<uuid>&page=&page_size=&sort=
//
// group_id é obrigatório. Retorna Page[SettlementDTO].
// A paginação é feita em memória porque o repo hoje não suporta
// LIMIT/OFFSET — grupo tende a ter poucos settlements; quando crescer,
// mover o LIMIT/OFFSET para o repo.
func (h *settlementHandler) List(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_balance/internal/features/settlement")
	ctx, span := tracer.Start(r.Context(), "settlementHandler.List")
	defer span.End()

	groupIDStr := r.URL.Query().Get("group_id")
	if groupIDStr == "" {
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			"missing required param: group_id", http.StatusBadRequest, nil,
		))
		return
	}
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			"invalid group_id", http.StatusBadRequest, err,
		))
		return
	}

	f, err := httputils.GetFilters(r, []string{"id", "-id", "settled_at", "-settled_at"})
	if err != nil {
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetAttributes(attribute.String("group.id", groupID.String()))

	items, err := h.svc.FindByGroupID(ctx, groupID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to list settlements")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	total := len(items)
	start := f.Offset()
	if start > total {
		start = total
	}
	end := start + f.Limit()
	if end > total {
		end = total
	}
	paged := items[start:end]

	dtoItems := make([]SettlementDTO, len(paged))
	for i, s := range paged {
		dtoItems[i] = s.ToDTO()
	}

	page := filters.NewPage(dtoItems, total, f.Page, f.PageSize)
	httputils.Respond(w, r, http.StatusOK, page, nil, nil)
}

// Delete ─────────────────────────────────────────────────────────────────
// DELETE /v1/settlements/{id}
//
// Soft delete + publish Kafka settlement.deleted. Retorna 204.
func (h *settlementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_balance/internal/features/settlement")
	ctx, span := tracer.Start(r.Context(), "settlementHandler.Delete")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		return
	}

	span.SetAttributes(attribute.String("settlement.id", id.String()))

	if err := h.svc.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete settlement")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Settlement deleted")
	httputils.Respond(w, r, http.StatusNoContent, nil, nil, nil)
}
