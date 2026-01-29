package middleware

import (
	"bytes"
	"context"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func setupLoggerWithObserver() (*logging.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	return &logging.Logger{Logger: zapLogger}, logs
}

func TestLogRequestBodyContent(t *testing.T) {
	tests := []struct {
		name            string
		bodyBytes       []byte
		expectLog       bool
		expectedField   string
		checkTruncation bool
	}{
		{
			name:      "empty body - no log",
			bodyBytes: []byte{},
			expectLog: false,
		},
		{
			name:          "small JSON body",
			bodyBytes:     []byte(`{"username":"testuser","email":"test@example.com"}`),
			expectLog:     true,
			expectedField: "app.request.body",
		},
		{
			name:          "JSON with sensitive data",
			bodyBytes:     []byte(`{"username":"testuser","password":"secret123"}`),
			expectLog:     true,
			expectedField: "app.request.body",
		},
		{
			name:          "non-JSON body",
			bodyBytes:     []byte("plain text body"),
			expectLog:     true,
			expectedField: "app.request.body",
		},
		{
			name:            "large body - truncation",
			bodyBytes:       bytes.Repeat([]byte("a"), 2000),
			expectLog:       true,
			expectedField:   "app.request.body",
			checkTruncation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logs := setupLoggerWithObserver()
			ctx := context.Background()

			logRequestBodyContent(tt.bodyBytes, logger, ctx)

			if tt.expectLog {
				assert.Greater(t, logs.Len(), 0, "should have logged")

				logEntry := logs.All()[0]
				assert.Equal(t, "request body captured", logEntry.Message)

				// Check for field
				found := false
				for _, field := range logEntry.Context {
					if field.Key == tt.expectedField {
						found = true
						break
					}
				}
				assert.True(t, found, "should contain expected field")
			} else {
				assert.Equal(t, 0, logs.Len(), "should not have logged")
			}
		})
	}
}

func TestLogResponseBodyIfEnabled(t *testing.T) {
	tests := []struct {
		name          string
		bodyBytes     []byte
		bodySize      int64
		expectLog     bool
		expectedField string
	}{
		{
			name:      "empty body - no log",
			bodyBytes: []byte{},
			bodySize:  0,
			expectLog: false,
		},
		{
			name:          "small JSON response",
			bodyBytes:     []byte(`{"id":123,"name":"test"}`),
			bodySize:      27,
			expectLog:     true,
			expectedField: "app.response.body",
		},
		{
			name:          "JSON with token",
			bodyBytes:     []byte(`{"token":"secret-token-123","user":"testuser"}`),
			bodySize:      48,
			expectLog:     true,
			expectedField: "app.response.body",
		},
		{
			name:          "non-JSON response",
			bodyBytes:     []byte("OK"),
			bodySize:      2,
			expectLog:     true,
			expectedField: "app.response.body",
		},
		{
			name:          "large response body",
			bodyBytes:     bytes.Repeat([]byte("b"), 2000),
			bodySize:      2000,
			expectLog:     true,
			expectedField: "app.response.body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logs := setupLoggerWithObserver()
			ctx := context.Background()

			logResponseBodyIfEnabled(tt.bodyBytes, logger, ctx, tt.bodySize)

			if tt.expectLog {
				assert.Greater(t, logs.Len(), 0, "should have logged")

				logEntry := logs.All()[0]
				assert.Equal(t, "response body captured", logEntry.Message)
			} else {
				assert.Equal(t, 0, logs.Len(), "should not have logged")
			}
		})
	}
}

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
			result := filterSensitiveFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
