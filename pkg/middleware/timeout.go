package middleware

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/logging"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TimeoutMiddleware wraps request context with timeout and logs timeout events
func TimeoutMiddleware(timeout time.Duration, logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			// Replace request context with timeout context
			c.SetRequest(c.Request().WithContext(ctx))

			// Channel to capture handler result
			done := make(chan error, 1)

			// Run handler in goroutine
			go func() {
				done <- next(c)
			}()

			// Wait for handler completion or timeout
			select {
			case err := <-done:
				// Handler completed normally
				return err
			case <-ctx.Done():
				// Timeout occurred
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					logger.WithContext(ctx).Warn("request timeout",
						zap.String("method", c.Request().Method),
						zap.String("path", c.Request().URL.Path),
						zap.Duration("timeout", timeout),
					)
					return controller.RequestTimeout(c, "request timeout")
				}
				// Context was cancelled for other reasons
				return ctx.Err()
			}
		}
	}
}
