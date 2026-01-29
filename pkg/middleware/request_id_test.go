package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogger() *logging.Logger {
	logger, _ := logging.NewDevelopmentLogger()
	return logger
}

// TestRequestIDMiddleware_HeaderBehavior tests request ID header generation and usage
func TestRequestIDMiddleware_HeaderBehavior(t *testing.T) {
	tests := []struct {
		name            string
		inputRequestID  string
		expectGenerated bool
		checkHandler    bool
	}{
		{
			name:            "generates new request ID when no header provided",
			inputRequestID:  "",
			expectGenerated: true,
			checkHandler:    true,
		},
		{
			name:            "uses existing request ID from header",
			inputRequestID:  "my-custom-request-id",
			expectGenerated: false,
			checkHandler:    false,
		},
		{
			name:            "sets response header",
			inputRequestID:  "",
			expectGenerated: true,
			checkHandler:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLogger()
			defer logger.Sync()

			e := echo.New()
			middleware := RequestIDMiddleware()

			handlerCalled := false
			handler := func(c echo.Context) error {
				handlerCalled = true
				return c.String(http.StatusOK, "OK")
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.inputRequestID != "" {
				req.Header.Set("X-Request-ID", tt.inputRequestID)
			}
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := middleware(handler)(ctx)
			require.NoError(t, err)

			if tt.checkHandler {
				assert.True(t, handlerCalled)
			}

			// Verify response header is set
			requestID := rec.Header().Get("X-Request-ID")
			assert.NotEmpty(t, requestID, "X-Request-ID header should be set")

			if !tt.expectGenerated {
				assert.Equal(t, tt.inputRequestID, requestID)
			}
		})
	}
}

// TestRequestIDMiddleware_InjectsIntoContext tests request ID is injected into context
func TestRequestIDMiddleware_InjectsIntoContext(t *testing.T) {

	e := echo.New()
	middleware := RequestIDMiddleware()

	expectedID := "test-request-123"
	handler := func(c echo.Context) error {
		ctx := c.Request().Context()
		requestID := logging.GetRequestID(ctx)
		assert.Equal(t, expectedID, requestID)
		return c.String(http.StatusOK, "OK")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", expectedID)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := middleware(handler)(ctx)
	require.NoError(t, err)
}

// TestRequestIDMiddleware_CallsNextHandler tests that next handler is called
func TestRequestIDMiddleware_CallsNextHandler(t *testing.T) {

	e := echo.New()
	middleware := RequestIDMiddleware()

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := middleware(handler)(ctx)
	require.NoError(t, err)
	assert.True(t, handlerCalled, "next handler should be called")
}

// TestRequestIDMiddleware_RequestBodyHandling tests various request body scenarios
func TestRequestIDMiddleware_RequestBodyHandling(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		method     string
		nilBody    bool
		verifyBody bool
	}{
		{
			name:       "handles nil request body",
			body:       "",
			method:     http.MethodGet,
			nilBody:    true,
			verifyBody: false,
		},
		{
			name:       "captures body size",
			body:       `{"key": "value"}`,
			method:     http.MethodPost,
			nilBody:    false,
			verifyBody: true,
		},
		{
			name:       "preserves request body for handler",
			body:       `{"test": "data"}`,
			method:     http.MethodPost,
			nilBody:    false,
			verifyBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLogger()
			defer logger.Sync()

			e := echo.New()
			middleware := RequestIDMiddleware()

			handler := func(c echo.Context) error {
				if tt.verifyBody {
					body, err := io.ReadAll(c.Request().Body)
					require.NoError(t, err)
					assert.Equal(t, tt.body, string(body))
				}
				return c.String(http.StatusOK, "OK")
			}

			var req *http.Request
			if tt.nilBody {
				req = httptest.NewRequest(tt.method, "/test", nil)
				req.Body = nil
			} else {
				req = httptest.NewRequest(tt.method, "/test", bytes.NewReader([]byte(tt.body)))
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := middleware(handler)(ctx)
			require.NoError(t, err)
			assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		})
	}
}

// TestRequestIDMiddleware_ErrorFromHandler tests error propagation
func TestRequestIDMiddleware_ErrorFromHandler(t *testing.T) {

	e := echo.New()
	middleware := RequestIDMiddleware()

	handler := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := middleware(handler)(ctx)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

// TestRequestIDMiddleware_DifferentHTTPMethods tests with various HTTP methods
func TestRequestIDMiddleware_DifferentHTTPMethods(t *testing.T) {

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			e := echo.New()
			middleware := RequestIDMiddleware()

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := middleware(handler)(ctx)
			require.NoError(t, err)
			assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
		})
	}
}

// TestRequestIDMiddleware_MultipleRequests tests each request gets unique ID
func TestRequestIDMiddleware_MultipleRequests(t *testing.T) {

	e := echo.New()
	middleware := RequestIDMiddleware()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	requestIDs := make(map[string]bool)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		err := middleware(handler)(ctx)
		require.NoError(t, err)

		id := rec.Header().Get("X-Request-ID")
		assert.NotEmpty(t, id)
		requestIDs[id] = true
	}

	// Each request should have unique ID
	assert.Equal(t, 3, len(requestIDs), "each request should have unique ID")
}

// TestRequestIDMiddleware_LargeRequestBody tests with large request bodies
func TestRequestIDMiddleware_LargeRequestBody(t *testing.T) {

	e := echo.New()
	middleware := RequestIDMiddleware()

	// Create large JSON payload
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeData[string(rune(i))] = "test-value"
	}
	payload, _ := json.Marshal(largeData)

	handler := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		require.NoError(t, err)
		assert.Equal(t, payload, body)
		return c.String(http.StatusOK, "OK")
	}

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := middleware(handler)(ctx)
	require.NoError(t, err)
}
