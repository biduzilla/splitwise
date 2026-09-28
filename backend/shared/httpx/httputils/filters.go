package httputils

import (
	"net/http"
	"shared/apierror"
	"shared/filters"
	"shared/validator"
)

func GetFilters(r *http.Request, sortSafelist []string) (filters.Filters, error) {
	if len(sortSafelist) == 0 {
		return filters.Filters{}, apierror.NewHTTPError(
			"sort safelist is required",
			http.StatusInternalServerError,
			nil,
		)
	}

	v := validator.New()

	f := filters.Filters{
		Page:         ReadIntParam(r, "page", 1, v),
		PageSize:     ReadIntParam(r, "page_size", 20, v),
		Sort:         ReadStringParam(r, "sort", sortSafelist[0]),
		SortSafelist: sortSafelist,
	}

	if filters.ValidateFilters(v, f); !v.Valid() {
		return filters.Filters{}, apierror.NewValidationError(v.Errors)
	}

	return f, nil
}
