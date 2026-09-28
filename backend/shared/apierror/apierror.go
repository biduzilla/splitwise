package apierror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"shared/auth/contexts"
	"shared/httpx/httpjson"
)

type StatusError interface {
	error
	StatusCode() int
	PublicMessage() string
}

type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) StatusCode() int {
	return e.Code
}

func (e *HTTPError) PublicMessage() string {
	return e.Message
}

type ValidationError struct {
	FieldErrors map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

func (e *ValidationError) StatusCode() int { return http.StatusUnprocessableEntity }

func (e *ValidationError) PublicMessage() string { return "validation failed" }

func NewHTTPError(message string, code int, err error) *HTTPError {
	return &HTTPError{Message: message, Code: code, Err: err}
}

func NewBadRequestError(err error) *HTTPError {
	return NewHTTPError(err.Error(), http.StatusBadRequest, err)
}

func NewValidationError(fieldErrors map[string]string) *ValidationError {
	return &ValidationError{FieldErrors: fieldErrors}
}

func ValidationAlreadyExists(field string) error {
	return &ValidationError{
		FieldErrors: map[string]string{
			field: fmt.Sprintf("a record with this %s already exists", field),
		},
	}
}

var (
	ErrRecordNotFound = &HTTPError{
		Code:    http.StatusNotFound,
		Message: "the requested resource could not be found",
		Err:     errors.New("record not found"),
	}

	ErrEditConflict = &HTTPError{
		Code:    http.StatusConflict,
		Message: "unable to update the record due to an edit conflict, please try again",
		Err:     errors.New("edit conflict"),
	}

	ErrInvalidCredentials = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "invalid authentication credentials",
		Err:     errors.New("invalid authentication credentials"),
	}

	ErrInactiveAccount = &HTTPError{
		Code:    http.StatusForbidden,
		Message: "your user account must be activated to access this resource",
		Err:     errors.New("inactive account"),
	}

	ErrTokenExpired = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "expired token",
		Err:     errors.New("token has expired"),
	}

	ErrInvalidTokenType = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "invalid token type",
		Err:     errors.New("invalid token type for this operation"),
	}

	ErrInvalidTokenClaims = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "token claims are invalid or malformed",
		Err:     errors.New("token claims are invalid or malformed"),
	}

	ErrRollbackFailed = &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "transaction rollback failed",
		Err:     errors.New("transaction rollback failed"),
	}

	ErrTransactionNotFound = &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "transaction not found",
		Err:     errors.New("transaction not found"),
	}

	ErrAuthenticationRequired = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "you must be authenticated to access this resource",
		Err:     errors.New("authentication required"),
	}

	ErrRequestTimeout = &HTTPError{
		Code:    http.StatusGatewayTimeout,
		Message: "request timeout, please try again",
		Err:     errors.New("request timeout"),
	}

	ErrInvalidAuthenticationToken = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "invalid or missing authentication token",
		Err:     errors.New("invalid or missing authentication token"),
	}

	ErrMalFormedToken = &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: "malformed token",
		Err:     errors.New("malformed token"),
	}

	ErrNotFoundPermission = &HTTPError{
		Code:    http.StatusForbidden,
		Message: "Not found permission",
		Err:     errors.New("your user account doesn't have the necessary permissions to access this resource"),
	}

	ErrRateLimitExceeded = &HTTPError{
		Code:    http.StatusTooManyRequests,
		Message: "rate limit exceed",
		Err:     errors.New("rate limit exceed"),
	}
)

type ErrorResponse struct {
	Path    string            `json:"path"`
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

type ErrorHandler struct {
	logger *slog.Logger
}

func NewErrorHandler(logger *slog.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

func (e *ErrorHandler) HandlerError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if valErr, ok := errors.AsType[*ValidationError](err); ok {
		e.write(
			w, r, http.StatusUnprocessableEntity,
			"validation failed",
			valErr.FieldErrors,
			err,
		)
		return
	}

	if statusErr, ok := errors.AsType[StatusError](err); ok {
		e.write(w, r,
			statusErr.StatusCode(),
			statusErr.PublicMessage(),
			nil,
			err,
		)
		return
	}

	e.write(
		w, r,
		http.StatusInternalServerError,
		"the server encountered a problem and could not process your request",
		nil,
		err,
	)
}

func (e *ErrorHandler) write(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
	fieldErrors map[string]string,
	internalErr error,
) {
	if statusCode >= 500 {
		e.logError(r, internalErr)
	}

	resp := ErrorResponse{
		Path:    r.URL.Path,
		Status:  http.StatusText(statusCode),
		Message: message,
		Errors:  fieldErrors,
	}

	if err := httpjson.WriteJSON(w, statusCode, resp, nil); err != nil {
		e.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (e *ErrorHandler) logError(r *http.Request, err error) {
	e.logger.Error("request error",
		"error", err,
		"request_method", r.Method,
		"request_url", r.URL.String(),
		"request_id", contexts.GetRequestID(r.Context()),
	)
}
