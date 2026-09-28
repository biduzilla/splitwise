package httputils

import (
	"net/http"
	"strconv"
	"time"
	"uuid"
)

type FieldErrorAdder interface {
	AddError(field, message string)
}

func ReadStringParam(r *http.Request, key, defaultValue string) string {
	if v := getValueQuery(r, key); v != "" {
		return v
	}
	return defaultValue
}

func ReadIntParam(
	r *http.Request,
	key string,
	defaultValue int,
	v FieldErrorAdder,
) int {
	s := getValueQuery(r, key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

func ReadDateParam(
	r *http.Request,
	key string,
	v FieldErrorAdder,
) *time.Time {
	s := getValueQuery(r, key)
	if s == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		v.AddError(key, "must be a valid date (YYYY-MM-DD)")
		return nil
	}
	return &t
}

func ReadUUIDParam(r *http.Request, key string, v FieldErrorAdder) uuid.UUID {
	s := getValueQuery(r, key)
	if s == "" {
		return uuid.Nil()
	}
	id, err := uuid.Parse(s)
	if err != nil {
		v.AddError(key, "must be a valid uuid")
		return uuid.Nil()
	}
	return id
}

func getValueQuery(r *http.Request, key string) string {
	if key == "" {
		return ""
	}
	return r.URL.Query().Get(key)
}
