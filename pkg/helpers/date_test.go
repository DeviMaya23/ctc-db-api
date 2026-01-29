package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseDate tests date parsing with various scenarios
func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		format   string
		wantErr  bool
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:    "valid date with standard format",
			dateStr: "02-01-2006",
			format:  "02-01-2006",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2006, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 2, result.Day())
			},
		},
		{
			name:    "valid date with RFC3339 format",
			dateStr: "2006-01-02T15:04:05Z",
			format:  time.RFC3339,
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2006, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 2, result.Day())
				assert.Equal(t, 15, result.Hour())
			},
		},
		{
			name:    "valid date with custom format",
			dateStr: "2024/12/25",
			format:  "2006/01/02",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 25, result.Day())
			},
		},
		{
			name:    "empty date string returns zero time",
			dateStr: "",
			format:  "02-01-2006",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				assert.True(t, result.IsZero())
			},
		},
		{
			name:    "invalid date format returns error",
			dateStr: "invalid-date",
			format:  "02-01-2006",
			wantErr: true,
			validate: func(t *testing.T, result time.Time) {
				assert.True(t, result.IsZero())
			},
		},
		{
			name:    "date doesn't match format returns error",
			dateStr: "2006-01-02",
			format:  "02-01-2006",
			wantErr: true,
			validate: func(t *testing.T, result time.Time) {
				assert.True(t, result.IsZero())
			},
		},
		{
			name:    "valid date with time",
			dateStr: "15-06-2025 14:30:45",
			format:  "02-01-2006 15:04:05",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2025, result.Year())
				assert.Equal(t, time.June, result.Month())
				assert.Equal(t, 15, result.Day())
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
				assert.Equal(t, 45, result.Second())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.dateStr, tt.format)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid date format")
			} else {
				require.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestParseDate_EdgeCases tests edge cases for date parsing
func TestParseDate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		format  string
		wantErr bool
	}{
		{
			name:    "leap year date",
			dateStr: "29-02-2024",
			format:  "02-01-2006",
			wantErr: false,
		},
		{
			name:    "invalid leap year date",
			dateStr: "29-02-2023",
			format:  "02-01-2006",
			wantErr: true,
		},
		{
			name:    "beginning of year",
			dateStr: "01-01-2000",
			format:  "02-01-2006",
			wantErr: false,
		},
		{
			name:    "end of year",
			dateStr: "31-12-2999",
			format:  "02-01-2006",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			dateStr: "   ",
			format:  "02-01-2006",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.dateStr, tt.format)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
