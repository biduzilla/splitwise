package group

import (
	"net/http"
	"shared/apierror"
	"shared/filters"
	"shared/httpx/httpjson"
	"shared/httpx/httputils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GroupHandler struct {
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

func NewHandler(svc service, errHandler errHandler) *GroupHandler {
	return &GroupHandler{svc: svc, errHandler: errHandler}
}

type groupHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// POST /v1/groups
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/group")
	ctx, span := tracer.Start(r.Context(), "GroupHandler.Create")
	defer span.End()

	var dto CreateGroupDTO
	if err := httpjson.ReadJSON(w, r, &dto); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(err.Error(), http.StatusBadRequest, err))
		return
	}

	span.SetAttributes(attribute.String("group.name", dto.Name))

	g, err := h.svc.Create(ctx, dto)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create group")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetAttributes(attribute.String("group.id", g.ID.String()))
	span.SetStatus(codes.Ok, "Group created")

	httputils.Respond(w, r, http.StatusCreated, g.ToDTO(), nil, h.errHandler)
}

// GET /v1/groups
func (h *GroupHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/group")
	ctx, span := tracer.Start(r.Context(), "GroupHandler.FindAll")
	defer span.End()

	f, err := httputils.GetFilters(r, []string{"id", "-id", "name", "-name"})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid filters")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	groups, meta, err := h.svc.ListByUser(ctx, f)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch groups")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	page := filters.NewPage(groups, meta.TotalRecords, meta.CurrentPage, meta.PageSize)
	dtos := filters.Map(page, func(g *Group) GroupDTO { return g.ToDTO() })

	span.SetAttributes(attribute.Int("groups.count", len(dtos.Content)))
	span.SetStatus(codes.Ok, "Groups fetched")

	httputils.Respond(w, r, http.StatusOK, dtos, nil, h.errHandler)
}

// GET /v1/groups/{id}
func (h *GroupHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/group")
	ctx, span := tracer.Start(r.Context(), "GroupHandler.FindByID")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "Invalid UUID")
		return
	}

	span.SetAttributes(attribute.String("group.id", id.String()))

	g, err := h.svc.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to fetch group")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Group fetched")
	httputils.Respond(w, r, http.StatusOK, g.ToDTO(), nil, h.errHandler)
}

// DELETE /v1/groups/{id}
func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_group/internal/features/group")
	ctx, span := tracer.Start(r.Context(), "GroupHandler.Delete")
	defer span.End()

	id, ok := httputils.ParseUUID(w, r, h.errHandler)
	if !ok {
		span.SetStatus(codes.Error, "Invalid UUID")
		return
	}

	span.SetAttributes(attribute.String("group.id", id.String()))

	if err := h.svc.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete group")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	span.SetStatus(codes.Ok, "Group deleted")
	httputils.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}
