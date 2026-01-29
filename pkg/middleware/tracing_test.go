package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoggerForTracing() *logging.Logger {
	logger, _ := logging.NewDevelopmentLogger()
	return logger
}

// TestTracingMiddleware_Configuration tests middleware configuration with various settings
func TestTracingMiddleware_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		otelEnabled    string
		serviceName    string
		testHandler    bool
		expectNonNil   bool
		expectCallNext bool
	}{
		{
			name:           "disabled - no-op middleware",
			otelEnabled:    "false",
			serviceName:    "",
			testHandler:    true,
			expectNonNil:   true,
			expectCallNext: true,
		},
		{
			name:         "enabled with service name",
			otelEnabled:  "true",
			serviceName:  "test-service",
			testHandler:  false,
			expectNonNil: true,
		},
		{
			name:         "enabled with custom service name",
			otelEnabled:  "true",
			serviceName:  "custom-service",
			testHandler:  true,
			expectNonNil: true,
		},
		{
			name:         "enabled with empty service name - uses default",
			otelEnabled:  "true",
			serviceName:  "",
			testHandler:  true,
			expectNonNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLoggerForTracing()
			defer logger.Sync()

			t.Setenv("OTEL_ENABLED", tt.otelEnabled)
			t.Setenv("OTEL_SERVICE_NAME", tt.serviceName)

			middleware := TracingMiddleware(logger)
			if tt.expectNonNil {
				assert.NotNil(t, middleware)
			}

			if tt.testHandler {
				e := echo.New()
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

				if tt.expectCallNext {
					assert.True(t, handlerCalled)
					assert.Equal(t, http.StatusOK, rec.Code)
				}
			}
		})
	}
}

// TestTracingMiddleware_DisabledBehaviors tests various behaviors when tracing is disabled
func TestTracingMiddleware_DisabledBehaviors(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func() *http.Request
		handler     func(c echo.Context) error
		expectError bool
		assertions  func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "calls next handler",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/test", nil)
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			},
			expectError: false,
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "OK", rec.Body.String())
			},
		},
		{
			name: "works with request ID",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req.WithContext(logging.WithRequestID(req.Context(), "test-request-123"))
			},
			handler: func(c echo.Context) error {
				// Request ID should be available in context
				ctx := c.Request().Context()
				_ = logging.GetRequestID(ctx)
				return c.String(http.StatusOK, "OK")
			},
			expectError: false,
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "doesn't interfere with custom status and response",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/test", nil)
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusCreated, "custom response")
			},
			expectError: false,
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Equal(t, "custom response", rec.Body.String())
			},
		},
		{
			name: "propagates errors",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/test", nil)
			},
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusBadRequest, "test error")
			},
			expectError: true,
			assertions: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLoggerForTracing()
			defer logger.Sync()

			t.Setenv("OTEL_ENABLED", "false")
			middleware := TracingMiddleware(logger)

			e := echo.New()
			req := tt.setupReq()
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := middleware(tt.handler)(ctx)
			tt.assertions(t, rec, err)
		})
	}
}

// TestTracingMiddleware_DifferentRoutes tests middleware works with various routes
func TestTracingMiddleware_DifferentRoutes(t *testing.T) {
	tests := []struct {
		name   string
		route  string
		method string
	}{
		{
			name:   "API users endpoint",
			route:  "/api/v1/users",
			method: http.MethodGet,
		},
		{
			name:   "API travelers endpoint with ID",
			route:  "/api/v1/travelers/123",
			method: http.MethodGet,
		},
		{
			name:   "API accessories endpoint",
			route:  "/api/v1/accessories",
			method: http.MethodGet,
		},
		{
			name:   "health check endpoint",
			route:  "/health",
			method: http.MethodGet,
		},
	}

	logger := setupLoggerForTracing()
	defer logger.Sync()

	t.Setenv("OTEL_ENABLED", "false")
	middleware := TracingMiddleware(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			req := httptest.NewRequest(tt.method, tt.route, nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := middleware(handler)(ctx)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
