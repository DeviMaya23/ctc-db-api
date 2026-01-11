package helpers

// PaginationParams holds pagination request parameters
type PaginationParams struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 10

// MaxPageSize prevents overly large page requests
const MaxPageSize = 100

// Normalize sets defaults and validates pagination params
func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
}

// Offset calculates the database offset
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// CalculateTotalPages calculates total pages from total count
func CalculateTotalPages(total int64, pageSize int) int {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}

// PaginatedResponse is a generic wrapper for paginated results
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse[T any](data []T, params PaginationParams, total int64) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Data:       data,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Total:      total,
		TotalPages: CalculateTotalPages(total, params.PageSize),
	}
}
