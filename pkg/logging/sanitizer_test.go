package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterSensitiveFields(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "filter password",
			input: map[string]interface{}{
				"username": "testuser",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "testuser",
				"password": "***REDACTED***",
			},
		},
		{
			name: "filter token",
			input: map[string]interface{}{
				"user":  "testuser",
				"token": "jwt-token-123",
			},
			expected: map[string]interface{}{
				"user":  "testuser",
				"token": "***REDACTED***",
			},
		},
		{
			name: "filter secret",
			input: map[string]interface{}{
				"data":   "public",
				"secret": "private-key",
			},
			expected: map[string]interface{}{
				"data":   "public",
				"secret": "***REDACTED***",
			},
		},
		{
			name: "filter api_key",
			input: map[string]interface{}{
				"endpoint": "/api/v1",
				"api_key":  "key-12345",
			},
			expected: map[string]interface{}{
				"endpoint": "/api/v1",
				"api_key":  "***REDACTED***",
			},
		},
		{
			name: "filter apikey",
			input: map[string]interface{}{
				"service": "external",
				"apikey":  "apikey-67890",
			},
			expected: map[string]interface{}{
				"service": "external",
				"apikey":  "***REDACTED***",
			},
		},
		{
			name: "multiple sensitive fields",
			input: map[string]interface{}{
				"username": "user",
				"password": "pass123",
				"token":    "token123",
				"email":    "user@test.com",
			},
			expected: map[string]interface{}{
				"username": "user",
				"password": "***REDACTED***",
				"token":    "***REDACTED***",
				"email":    "user@test.com",
			},
		},
		{
			name: "no sensitive fields",
			input: map[string]interface{}{
				"id":    123,
				"name":  "test",
				"email": "test@example.com",
			},
			expected: map[string]interface{}{
				"id":    123,
				"name":  "test",
				"email": "test@example.com",
			},
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterSensitiveFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
