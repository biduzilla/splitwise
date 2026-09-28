package httputils

import (
	"net/http"
	"shared/apierror"
	"shared/httpx/httpjson"
	"uuid"
)

type ErrorRespond interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

func ParseUUID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp ErrorRespond,
) (uuid.UUID, bool) {
	id, err := PathParamUUID(r, "id")
	if err != nil {
		errRsp.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return uuid.Nil(), false
	}
	return id, true
}

func ParseIntID(
	w http.ResponseWriter,
	r *http.Request,
	errRsp ErrorRespond,
) (int, bool) {
	id, err := PathParamInt(r, "id")
	if err != nil {
		errRsp.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return 0, false
	}
	return id, true
}

func ParseStringField(
	w http.ResponseWriter,
	r *http.Request,
	errRsp ErrorRespond,
	field string,
) (string, bool) {
	value, err := PathParamString(r, field)
	if err != nil {
		errRsp.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return "", false
	}
	return value, true
}

func Respond(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
	headers http.Header,
	errRsp ErrorRespond,
) {
	if err := httpjson.WriteJSON(w, status, data, headers); err != nil {
		errRsp.HandlerError(w, r, err)
	}
}
