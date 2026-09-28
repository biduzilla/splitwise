package httputils

import (
	"fmt"
	"net/http"
	"strconv"
	"uuid"
)

func PathParamString(r *http.Request, key string) (string, error) {
	s := r.PathValue(key)
	if s == "" {
		return "", fmt.Errorf("missing path parameter: %s", key)
	}
	return s, nil
}

func PathParamInt(r *http.Request, key string) (int, error) {
	s, err := PathParamString(r, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}
	return n, nil
}

func PathParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	s, err := PathParamString(r, key)
	if err != nil {
		return uuid.Nil(), err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil(), fmt.Errorf("%s must be a valid uuid", key)
	}
	return id, nil
}
