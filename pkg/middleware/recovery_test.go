package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLoggerForRecovery() *logging.Logger {
	logger, _ := logging.NewDevelopmentLogger()
	return logger
}

// TestRecoveryMiddleware_NormalFlow tests middleware does not interfere with normal requests
func TestRecoveryMiddleware_NormalFlow(t *testing.T) {
	logger := setupLoggerForRecovery()
	defer logger.Sync()

	e := echo.New()
	middleware := RecoveryMiddleware(logger)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := middleware(handler)(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

// TestRecoveryMiddleware_PanicRecovery tests recovery from various panic types and scenarios
func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	tests := []struct {
		name        string
		setupPanic  func(c echo.Context) error
		method      string
		path        string
		showDetails string
	}{
		{
			name: "string panic",
			setupPanic: func(c echo.Context) error {
				panic("something went wrong")
			},
			method:      http.MethodGet,
			path:        "/test",
			showDetails: "false",
		},
		{
			name: "error panic",
			setupPanic: func(c echo.Context) error {
				panic(errors.New("database connection failed"))
			},
			method:      http.MethodPost,
			path:        "/api/v1/users",
			showDetails: "false",
		},
		{
			name: "nested function panic",
			setupPanic: func(c echo.Context) error {
				nestedFunc := func() {
					panic("nested panic")
				}
				nestedFunc()
				return c.String(http.StatusOK, "should not reach here")
			},
			method:      http.MethodGet,
			path:        "/nested",
			showDetails: "false",
		},
		{
			name: "int panic",
			setupPanic: func(c echo.Context) error {
				panic(42)
			},
			method:      http.MethodGet,
			path:        "/test",
			showDetails: "false",
		},
		{
			name: "struct panic",
			setupPanic: func(c echo.Context) error {
				panic(struct{ msg string }{msg: "struct panic"})
			},
			method:      http.MethodPut,
			path:        "/api/v1/items/1",
			showDetails: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLoggerForRecovery()
			defer logger.Sync()

			t.Setenv("SHOW_PANIC_DETAILS", tt.showDetails)

			e := echo.New()
			middleware := RecoveryMiddleware(logger)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.SetPath(tt.path)

			err := middleware(tt.setupPanic)(ctx)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)

			var response controller.ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "internal server error", response.Message)
		})
	}
}

// TestRecoveryMiddleware_ContextPropagation tests request ID is in logs
func TestRecoveryMiddleware_ContextPropagation(t *testing.T) {
	logger := setupLoggerForRecovery()
	defer logger.Sync()

	t.Setenv("SHOW_PANIC_DETAILS", "false")

	e := echo.New()

	// Chain with request ID middleware
	requestIDMiddleware := RequestIDMiddleware()
	recoveryMiddleware := RecoveryMiddleware(logger)

	handler := func(c echo.Context) error {
		panic("test panic with request ID")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/test")

	// Apply both middlewares
	err := requestIDMiddleware(recoveryMiddleware(handler))(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Verify request ID header is set
	requestID := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}

// TestRecoveryMiddleware_ResponseFormat tests JSON response format with and without panic details
func TestRecoveryMiddleware_ResponseFormat(t *testing.T) {
	tests := []struct {
		name               string
		panicValue         interface{}
		showPanicDetails   string
		expectDetailFields bool
		validateDetails    func(t *testing.T, response map[string]interface{})
	}{
		{
			name:               "production mode - no details",
			panicValue:         "test panic",
			showPanicDetails:   "false",
			expectDetailFields: false,
			validateDetails: func(t *testing.T, response map[string]interface{}) {
				assert.NotContains(t, response, "error")
				assert.NotContains(t, response, "type")
			},
		},
		{
			name:               "development mode - with details for string panic",
			panicValue:         "detailed error message",
			showPanicDetails:   "true",
			expectDetailFields: true,
			validateDetails: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
				assert.Contains(t, response, "type")
				assert.Contains(t, response["error"], "detailed error message")
				assert.NotEmpty(t, response["type"])
			},
		},
		{
			name:               "development mode - with details for error panic",
			panicValue:         errors.New("error with details"),
			showPanicDetails:   "true",
			expectDetailFields: true,
			validateDetails: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
				assert.Contains(t, response, "type")
				assert.Contains(t, response["error"], "error with details")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := setupLoggerForRecovery()
			defer logger.Sync()

			t.Setenv("SHOW_PANIC_DETAILS", tt.showPanicDetails)

			e := echo.New()
			middleware := RecoveryMiddleware(logger)

			handler := func(c echo.Context) error {
				panic(tt.panicValue)
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.SetPath("/test")

			err := middleware(handler)(ctx)
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "internal server error", response["message"])

			tt.validateDetails(t, response)
		})
	}
}
