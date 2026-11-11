package expenses

import (
	"net/http"
	"shared/apierror"
	"shared/filters"
	"shared/httpx/httpjson"
	"shared/httpx/httputils"
	"shared/validator"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type expenseHandler struct {
	service    Service
	errHandler errHandler
	safeSort   []string
}

type errHandler interface {
	HandlerError(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	)
}

func newHandler(
	s Service,
	errHandler errHandler,
) *expenseHandler {
	return &expenseHandler{
		service:    s,
		errHandler: errHandler,
		safeSort: []string{"id", "-id", "amount", "-amount",
			"expense_date", "-expense_date",
			"created_at", "-created_at"},
	}
}

type handler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func (h *expenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_expense/internal/features/expense")
	ctx, span := tracer.Start(r.Context(), "expenseHandler.Create")
	defer span.End()

	var in CreateExpenseDTO
	if err := httpjson.ReadJSON(w, r, &in); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apierror.NewBadRequestError(err))
		return
	}

	model, err := in.ToModel()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to Model")
		h.errHandler.HandlerError(w, r, apierror.NewBadRequestError(err))
		return
	}

	err = h.service.Create(ctx, model, in.Participants)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	httputils.Respond(w, r, http.StatusCreated, model.ToDTO(), nil, h.errHandler)
}

func (h *expenseHandler) Get(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_expense/internal/features/expense")
	ctx, span := tracer.Start(r.Context(), "expenseHandler.Get")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "Invalid UUID")
		return
	}

	span.SetAttributes(attribute.String("expense.id", id.String()))

	model, err := h.service.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch expense")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Expense fetched")
	httputils.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *expenseHandler) List(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_expense/internal/features/expense")
	ctx, span := tracer.Start(r.Context(), "expenseHandler.List")
	defer span.End()

	groupID, err := httputils.PathParamUUID(r, "group_id")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid UUID")
		h.errHandler.HandlerError(w, r, apierror.NewBadRequestError(err))
	}

	startDate, endDate, f, err := h.getFindAllParams(r)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid query parameters")
		h.errHandler.HandlerError(w, r, err)
	}

	models, metadata, err := h.service.FindAll(
		ctx,
		groupID,
		f,
		startDate,
		endDate,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch expenses")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	page := filters.NewPage(models, metadata.TotalRecords, metadata.CurrentPage, metadata.PageSize)
	dtos := filters.Map(page, func(u *Expense) ExpenseDTO { return u.ToDTO() })

	httputils.Respond(
		w,
		r,
		http.StatusOK,
		dtos,
		nil,
		h.errHandler,
	)
}

func (h *expenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_expense/internal/features/expense")
	ctx, span := tracer.Start(r.Context(), "expenseHandler.Update")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		return
	}
	var in UpdateExpenseDTO
	if err := httpjson.ReadJSON(w, r, &in); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	model, err := in.ToModel()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to Model")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetAttributes(attribute.String("expense.id", id.String()))

	err = h.service.Update(ctx, id, in.Participants, model)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to Update")
		h.errHandler.HandlerError(w, r, err)
		return
	}
	httputils.Respond(w, r, http.StatusOK, model.ToDTO(), nil, nil)
}

func (h *expenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_expense/internal/features/expense")
	ctx, span := tracer.Start(r.Context(), "expenseHandler.Delete")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "Invalid UUID")
		return
	}

	span.SetAttributes(attribute.String("expense.id", id.String()))

	if err := h.service.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to Delete")
		h.errHandler.HandlerError(w, r, err)
		h.errHandler.HandlerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *expenseHandler) getFindAllParams(r *http.Request) (*time.Time, *time.Time, filters.Filters, error) {
	v := validator.New()
	startDate := httputils.ReadDateParam(r, "start_date", v)
	endDate := httputils.ReadDateParam(r, "end_date", v)

	if !v.Valid() {
		return nil, nil, filters.Filters{}, apierror.NewValidationError(v.Errors)
	}

	filterParam, err := httputils.GetFilters(r, h.safeSort)
	if err != nil {
		return nil, nil, filters.Filters{}, apierror.NewBadRequestError(err)
	}

	return startDate, endDate, filterParam, nil
}
