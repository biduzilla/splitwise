package filters

import (
	"math"
	"shared/validator"
	"slices"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) Limit() int {
	return f.PageSize
}

func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}

func (f Filters) SortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{
			CurrentPage:  1,
			PageSize:     pageSize,
			FirstPage:    1,
			LastPage:     1,
			TotalRecords: 0,
		}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

type Page[T any] struct {
	Content  []T      `json:"content"`
	Metadata Metadata `json:"metadata"`
}

func NewPage[T any](content []T, totalRecords, page, pageSize int) Page[T] {
	if content == nil {
		content = []T{}
	}
	return Page[T]{
		Content:  content,
		Metadata: CalculateMetadata(totalRecords, page, pageSize),
	}
}

func NewEmptyPage[T any](page, pageSize int) Page[T] {
	return NewPage([]T{}, 0, page, pageSize)
}

func Map[T, U any](p Page[T], fn func(T) U) Page[U] {
	out := make([]U, len(p.Content))
	for i, v := range p.Content {
		out[i] = fn(v)
	}
	return Page[U]{
		Content:  out,
		Metadata: p.Metadata,
	}
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
