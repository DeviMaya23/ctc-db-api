package middleware

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
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

func TestRequestBodyLoggingMiddleware_WithRequestBody(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		requestBody     string
		expectInLog     string
		checkStatusCode bool
		expectedStatus  int
	}{
		{
			name:            "POST request with JSON body",
			method:          "POST",
			requestBody:     `{"username":"testuser","email":"test@example.com"}`,
			expectInLog:     "request started",
			checkStatusCode: true,
			expectedStatus:  200,
		},
		{
			name:            "PUT request with JSON body",
			method:          "PUT",
			requestBody:     `{"id":1,"name":"updated"}`,
			expectInLog:     "request completed",
			checkStatusCode: true,
			expectedStatus:  200,
		},
		{
			name:            "GET request without body",
			method:          "GET",
			requestBody:     "",
			expectInLog:     "request started",
			checkStatusCode: true,
			expectedStatus:  200,
		},
		{
			name:            "request body with sensitive data",
			method:          "POST",
			requestBody:     `{"username":"user","password":"secret123"}`,
			expectInLog:     "request started",
			checkStatusCode: true,
			expectedStatus:  200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logs := setupLoggerWithObserver()
			middleware := RequestBodyLoggingMiddleware(logger)

			// Create echo context
			e := echo.New()
			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Track if handler was called
			handlerCalled := false
			handler := func(c echo.Context) error {
				handlerCalled = true
				return c.JSON(200, map[string]string{"status": "ok"})
			}

			// Execute middleware
			err := middleware(handler)(c)

			// Assertions
			assert.NoError(t, err)
			assert.True(t, handlerCalled, "handler should be called")
			if tt.checkStatusCode {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}

			// Check logs contain expected message
			assert.Greater(t, logs.Len(), 0, "should have logged")
			logMessages := make([]string, 0)
			for _, entry := range logs.All() {
				logMessages = append(logMessages, entry.Message)
			}
			assert.Contains(t, logMessages, tt.expectInLog)
		})
	}
}

func TestRequestBodyLoggingMiddleware_LogsMetrics(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		expectDuration   bool
		expectBodySize   bool
		expectStatusCode bool
	}{
		{
			name:             "logs duration and body size",
			requestBody:      `{"test":"data"}`,
			expectDuration:   true,
			expectBodySize:   true,
			expectStatusCode: true,
		},
		{
			name:             "empty body still logs metrics",
			requestBody:      "",
			expectDuration:   true,
			expectBodySize:   true,
			expectStatusCode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logs := setupLoggerWithObserver()
			middleware := RequestBodyLoggingMiddleware(logger)

			e := echo.New()
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.JSON(200, map[string]string{"ok": "true"})
			}

			err := middleware(handler)(c)
			assert.NoError(t, err)

			// Check completion log
			completionLog := logs.FilterMessage("request completed")
			assert.Greater(t, completionLog.Len(), 0, "should log request completed")

			logEntry := completionLog.All()[0]
			fields := logEntry.Context

			if tt.expectDuration {
				found := false
				for _, field := range fields {
					if field.Key == "http.request.duration" {
						found = true
						break
					}
				}
				assert.True(t, found, "should log duration")
			}

			if tt.expectBodySize {
				found := false
				for _, field := range fields {
					if field.Key == "http.request.body.size" {
						found = true
						break
					}
				}
				assert.True(t, found, "should log request body size")
			}

			if tt.expectStatusCode {
				found := false
				for _, field := range fields {
					if field.Key == "http.status_code" {
						found = true
						break
					}
				}
				assert.True(t, found, "should log status code")
			}
		})
	}
}

func TestRequestBodyLoggingMiddleware_RestoresRequestBody(t *testing.T) {
	logger, _ := setupLoggerWithObserver()
	middleware := RequestBodyLoggingMiddleware(logger)

	requestBody := `{"test":"data"}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(requestBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		// Handler should be able to read the body
		bodyBytes := make([]byte, len(requestBody))
		n, err := c.Request().Body.Read(bodyBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(requestBody), n, "should read full body")
		assert.Equal(t, requestBody, string(bodyBytes))
		return c.JSON(200, map[string]string{"ok": "true"})
	}

	err := middleware(handler)(c)
	assert.NoError(t, err)
}

func TestRequestBodyLoggingMiddleware_HandlerError(t *testing.T) {
	logger, logs := setupLoggerWithObserver()
	middleware := RequestBodyLoggingMiddleware(logger)

	e := echo.New()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"data"}`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler returns an error
	handler := func(c echo.Context) error {
		return echo.NewHTTPError(400, "bad request")
	}

	err := middleware(handler)(c)
	assert.Error(t, err)

	// Should still log request completion
	assert.Greater(t, logs.Len(), 0, "should still log even on handler error")
	completionLog := logs.FilterMessage("request completed")
	assert.Greater(t, completionLog.Len(), 0, "should log completion")
}

func TestRequestBodyLoggingMiddleware_ResponseBodyCapture(t *testing.T) {
	logger, logs := setupLoggerWithObserver()
	middleware := RequestBodyLoggingMiddleware(logger)

	e := echo.New()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	responseData := map[string]string{"status": "ok", "data": "test"}
	handler := func(c echo.Context) error {
		return c.JSON(200, responseData)
	}

	err := middleware(handler)(c)
	assert.NoError(t, err)

	// Check that response body size is logged
	completionLog := logs.FilterMessage("request completed")
	assert.Greater(t, completionLog.Len(), 0)

	logEntry := completionLog.All()[0]
	found := false
	for _, field := range logEntry.Context {
		if field.Key == "http.response.body.size" && field.Integer > 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "should log non-zero response body size")
}

func TestRequestBodyLoggingMiddleware_HTTPAttributes(t *testing.T) {
	logger, logs := setupLoggerWithObserver()
	middleware := RequestBodyLoggingMiddleware(logger)

	e := echo.New()
	req := httptest.NewRequest("POST", "/api/users", strings.NewReader(`{"name":"test"}`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.JSON(201, map[string]string{"id": "123"})
	}

	err := middleware(handler)(c)
	assert.NoError(t, err)

	// Check start log
	startLog := logs.FilterMessage("request started")
	assert.Greater(t, startLog.Len(), 0)
	startEntry := startLog.All()[0]

	// Check method and route in start log
	assert.NotEmpty(t, startEntry.Context)
	methodFound := false
	for _, field := range startEntry.Context {
		if field.Key == "http.method" && field.String == "POST" {
			methodFound = true
			break
		}
	}
	assert.True(t, methodFound, "should log HTTP method")

	// Check completion log
	completionLog := logs.FilterMessage("request completed")
	assert.Greater(t, completionLog.Len(), 0)
	completionEntry := completionLog.All()[0]

	statusFound := false
	for _, field := range completionEntry.Context {
		if field.Key == "http.status_code" && field.Integer == 201 {
			statusFound = true
			break
		}
	}
	assert.True(t, statusFound, "should log HTTP status code")
}
