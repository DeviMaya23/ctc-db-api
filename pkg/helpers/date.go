package helpers

import (
	"fmt"
	"time"
)

// ParseDate parses a date string using the provided format.
// Returns a zero time.Time if the input string is empty.
// Returns an error if the date format is invalid.
func ParseDate(dateStr string, format string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	parsedDate, err := time.Parse(format, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}

	return parsedDate, nil
}
