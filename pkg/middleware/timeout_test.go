package middleware

import (
	"context"
	"encoding/json"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutMiddleware_CompletesBeforeTimeout(t *testing.T) {
	e := echo.New()
	logger, _ := logging.NewDevelopmentLogger()

	// Fast handler that completes before timeout
	fastHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middleware := TimeoutMiddleware(1*time.Second, logger)
	handler := middleware(fastHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestTimeoutMiddleware_TimesOut(t *testing.T) {
	e := echo.New()
	logger, _ := logging.NewDevelopmentLogger()

	// Slow handler that exceeds timeout
	slowHandler := func(c echo.Context) error {
		select {
		case <-time.After(2 * time.Second):
			return c.String(http.StatusOK, "completed")
		case <-c.Request().Context().Done():
			// Handler should respect context cancellation
			return c.Request().Context().Err()
		}
	}

	middleware := TimeoutMiddleware(100*time.Millisecond, logger)
	handler := middleware(slowHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)

	// Parse response
	var response controller.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "request timeout", response.Message)
}

func TestTimeoutMiddleware_ContextPropagation(t *testing.T) {
	e := echo.New()
	logger, _ := logging.NewDevelopmentLogger()

	// Handler that checks if context has timeout
	handlerWithContextCheck := func(c echo.Context) error {
		ctx := c.Request().Context()

		// Verify context has deadline
		deadline, ok := ctx.Deadline()
		assert.True(t, ok, "context should have deadline")
		assert.True(t, time.Until(deadline) > 0, "deadline should be in the future")

		return c.String(http.StatusOK, "context propagated")
	}

	middleware := TimeoutMiddleware(5*time.Second, logger)
	handler := middleware(handlerWithContextCheck)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTimeoutMiddleware_HandlerRespectsContextCancellation(t *testing.T) {
	e := echo.New()
	logger, _ := logging.NewDevelopmentLogger()

	cancelled := false

	// Handler that respects context cancellation
	handlerThatChecksContext := func(c echo.Context) error {
		ctx := c.Request().Context()

		select {
		case <-time.After(1 * time.Second):
			return c.String(http.StatusOK, "completed")
		case <-ctx.Done():
			cancelled = true
			return ctx.Err()
		}
	}

	middleware := TimeoutMiddleware(50*time.Millisecond, logger)
	handler := middleware(handlerThatChecksContext)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	require.NoError(t, err)

	// Should return timeout response
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)

	// Give goroutine time to finish
	time.Sleep(100 * time.Millisecond)

	// Verify context was cancelled
	assert.True(t, cancelled, "handler should have detected context cancellation")
}

func TestTimeoutMiddleware_DifferentTimeoutValues(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handlerDelay   time.Duration
		expectedStatus int
	}{
		{
			name:           "completes well before timeout",
			timeout:        1 * time.Second,
			handlerDelay:   10 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exceeds short timeout",
			timeout:        50 * time.Millisecond,
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name:           "completes just before timeout",
			timeout:        200 * time.Millisecond,
			handlerDelay:   150 * time.Millisecond,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			logger, _ := logging.NewDevelopmentLogger()

			handler := func(c echo.Context) error {
				select {
				case <-time.After(tt.handlerDelay):
					return c.String(http.StatusOK, "completed")
				case <-c.Request().Context().Done():
					return c.Request().Context().Err()
				}
			}

			middleware := TimeoutMiddleware(tt.timeout, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := wrappedHandler(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestTimeoutMiddleware_ContextDeadlineExceeded(t *testing.T) {
	e := echo.New()
	logger, _ := logging.NewDevelopmentLogger()

	handlerThatChecksError := func(c echo.Context) error {
		// Simulate work
		time.Sleep(200 * time.Millisecond)

		// Check if context is done
		if c.Request().Context().Err() == context.DeadlineExceeded {
			return c.Request().Context().Err()
		}

		return c.String(http.StatusOK, "completed")
	}

	middleware := TimeoutMiddleware(100*time.Millisecond, logger)
	handler := middleware(handlerThatChecksError)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := handler(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, rec.Code)
}
