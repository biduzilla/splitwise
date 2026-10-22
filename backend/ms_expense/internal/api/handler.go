package api

import (
	"shared/apierror"
)

type handlers struct {
}

func NewHandlers(
	services *services,
	errHandler *apierror.ErrorHandler,
) *handlers {
	return &handlers{}
}
