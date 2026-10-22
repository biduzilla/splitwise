package user

import (
	"net/http"
	"shared/apierror"
	"shared/filters"
	"shared/httpx/httpjson"
	"shared/httpx/httputils"
	"shared/validator"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type UserHandler struct {
	userService userService
	errHandler  errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type userHandler interface {
	SignUp(
		w http.ResponseWriter,
		r *http.Request,
	)

	FindAllByIds(
		w http.ResponseWriter,
		r *http.Request,
	)

	FindAll(
		w http.ResponseWriter,
		r *http.Request,
	)

	Update(
		w http.ResponseWriter,
		r *http.Request,
	)

	DeleteById(
		w http.ResponseWriter,
		r *http.Request,
	)
}

func NewHandler(
	userService userService,
	errorHandler errorHandler,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		errHandler:  errorHandler,
	}
}

func (h *UserHandler) FindAllByIds(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "Userhttputils.FindAllByIds")

	defer span.End()

	v := validator.New()
	ids := httputils.ReadUUIDListParam(r, "ids", 100, v)
	if !v.Valid() {
		span.SetStatus(codes.Error, "invalid ids")
		h.errHandler.HandlerError(w, r, apierror.NewValidationError(v.Errors))
		return
	}

	if len(ids) == 0 {
		span.SetStatus(codes.Error, "missing ids")
		h.errHandler.HandlerError(w, r, apierror.NewValidationError(
			map[string]string{"ids": "must be provided"},
		))
		return
	}

	span.SetAttributes(attribute.Int("ids.count", len(ids)))

	f, err := httputils.GetFilters(r, []string{"id", "name", "-id", "-name"})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid filters")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	models, metadata, err := h.userService.FindAllByIds(
		ctx,
		ids,
		f,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch users")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	page := filters.NewPage(models, metadata.TotalRecords, metadata.CurrentPage, metadata.PageSize)
	dtos := filters.Map(page, func(u *User) UserDTO { return u.ToDTO() })

	httputils.Respond(
		w,
		r,
		http.StatusOK,
		dtos,
		nil,
		h.errHandler,
	)
}

func (h *UserHandler) SignUp(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "Userhttputils.SignUp")

	defer span.End()

	var dto CreateUserDTO
	if err := httpjson.ReadJSON(w, r, &dto); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	span.SetAttributes(attribute.String("user.Email", dto.Email))
	span.SetAttributes(attribute.String("user.Name", dto.Name))

	user, err := h.userService.SignUp(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	httputils.Respond(w, r, http.StatusCreated, user.ToDTO(), nil, h.errHandler)
}

func (h *UserHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "Userhttputils.FindAll")

	defer span.End()

	var input struct {
		search string
		filters.Filters
	}

	v := validator.New()

	input.search = httputils.ReadStringParam(r, "search", "")
	input.Filters.Page = httputils.ReadIntParam(r, "page", 1, v)
	input.Filters.PageSize = httputils.ReadIntParam(r, "page_size", 20, v)
	input.Filters.Sort = httputils.ReadStringParam(r, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "-id", "-name"}

	if filters.ValidateFilters(v, input.Filters); !v.Valid() {
		h.errHandler.HandlerError(
			w,
			r,
			apierror.NewValidationError(v.Errors))
		return
	}

	models, metadata, err := h.userService.FindAll(ctx,
		input.search,
		input.Filters,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch users")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	dtos := make([]UserDTO, len(models))

	for i, m := range models {
		dtos[i] = m.ToDTO()
	}

	httputils.Respond(
		w,
		r,
		http.StatusOK,
		httpjson.Envelope{
			"content":  dtos,
			"metadata": metadata,
		},
		nil,
		h.errHandler,
	)
}

func (h *UserHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "Userhttputils.Update")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	var dto UserDTO
	if err := httpjson.ReadJSON(w, r, &dto); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	model := dto.ToModel()
	model.ID = id

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := h.userService.Update(ctx, model); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update user")
		h.errHandler.HandlerError(w, r, err)
		return
	}
	httputils.Respond(w, r, http.StatusOK, model.ToDTO(), nil, h.errHandler)
}

func (h *UserHandler) DeleteById(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracer := otel.Tracer("ms_auth/internal/features/user")
	ctx, span := tracer.Start(r.Context(), "Userhttputils.DeleteById")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "invalid id")
		return
	}

	span.SetAttributes(attribute.String("user.id", id.String()))

	if err := h.userService.DeleteById(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete user")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	httputils.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
