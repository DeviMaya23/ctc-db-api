package helpers

import (
	"os"
	"strconv"
	"strings"
)

func EnvWithDefault(key, defaultValue string) string {

	strVal := os.Getenv(key)
	if strVal == "" {
		return defaultValue
	}
	return strVal
}

func EnvWithDefaultBool(key string, defaultValue bool) bool {

	strVal := os.Getenv(key)
	if strVal == "" {
		return defaultValue
	}

	return strings.EqualFold(strVal, "true")
}

// EnvWithDefaultFloat returns float64 from env or default
func EnvWithDefaultFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
