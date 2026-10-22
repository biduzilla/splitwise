package httpclient

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"shared/apierror"
)

func DefaultErrorMapper(status int, body []byte) error {
	var upstream apierror.ErrorResponse
	if err := json.Unmarshal(body, &upstream); err != nil {
		return apierror.NewHTTPError(
			http.StatusText(status),
			status,
			fmt.Errorf("upstream %d: %s", status, string(body)),
		)
	}

	if len(upstream.Errors) > 0 {
		return apierror.NewValidationError(upstream.Errors)
	}

	return apierror.NewHTTPError(
		upstream.Message,
		status,
		errors.New(upstream.Message),
	)
}
