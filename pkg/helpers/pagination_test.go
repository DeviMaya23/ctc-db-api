package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPaginationParams_Normalize tests pagination parameter normalization
func TestPaginationParams_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
	}{
		{
			name:         "invalid page zero",
			page:         0,
			pageSize:     10,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "negative page",
			page:         -5,
			pageSize:     10,
			expectedPage: 1,
			expectedSize: 10,
		},
		{
			name:         "invalid page size zero",
			page:         1,
			pageSize:     0,
			expectedPage: 1,
			expectedSize: DefaultPageSize,
		},
		{
			name:         "negative page size",
			page:         1,
			pageSize:     -10,
			expectedPage: 1,
			expectedSize: DefaultPageSize,
		},
		{
			name:         "exceeding max page size",
			page:         1,
			pageSize:     500,
			expectedPage: 1,
			expectedSize: MaxPageSize,
		},
		{
			name:         "valid params remain unchanged",
			page:         2,
			pageSize:     20,
			expectedPage: 2,
			expectedSize: 20,
		},
		{
			name:         "multiple invalid fields",
			page:         -1,
			pageSize:     -1,
			expectedPage: 1,
			expectedSize: DefaultPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := PaginationParams{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}

			params.Normalize()

			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedSize, params.PageSize)
		})
	}
}

// TestPaginationParams_Offset tests offset calculation
func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedOffset int
	}{
		{
			name:           "first page",
			page:           1,
			pageSize:       10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			pageSize:       10,
			expectedOffset: 10,
		},
		{
			name:           "third page",
			page:           3,
			pageSize:       25,
			expectedOffset: 50,
		},
		{
			name:           "large page number",
			page:           100,
			pageSize:       50,
			expectedOffset: 4950,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := PaginationParams{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}

			offset := params.Offset()

			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

// TestCalculateTotalPages tests page calculation with various scenarios
func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		pageSize int
		expected int
	}{
		{
			name:     "exact divisor",
			total:    100,
			pageSize: 10,
			expected: 10,
		},
		{
			name:     "with remainder",
			total:    105,
			pageSize: 10,
			expected: 11,
		},
		{
			name:     "single page",
			total:    5,
			pageSize: 10,
			expected: 1,
		},
		{
			name:     "zero total",
			total:    0,
			pageSize: 10,
			expected: 0,
		},
		{
			name:     "large total",
			total:    10000,
			pageSize: 25,
			expected: 400,
		},
		{
			name:     "50 items, 10 per page",
			total:    50,
			pageSize: 10,
			expected: 5,
		},
		{
			name:     "51 items, 10 per page",
			total:    51,
			pageSize: 10,
			expected: 6,
		},
		{
			name:     "100 items, 50 per page",
			total:    100,
			pageSize: 50,
			expected: 2,
		},
		{
			name:     "101 items, 50 per page",
			total:    101,
			pageSize: 50,
			expected: 3,
		},
		{
			name:     "1000 items, 100 per page",
			total:    1000,
			pageSize: 100,
			expected: 10,
		},
		{
			name:     "1001 items, 100 per page",
			total:    1001,
			pageSize: 100,
			expected: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotalPages(tt.total, tt.pageSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewPaginatedResponse_CreatesValidResponse tests response creation
func TestNewPaginatedResponse_CreatesValidResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	params := PaginationParams{Page: 1, PageSize: 10}
	total := int64(3)

	response := NewPaginatedResponse(data, params, total)

	assert.Equal(t, data, response.Data)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, int64(3), response.Total)
	assert.Equal(t, 1, response.TotalPages)
}
