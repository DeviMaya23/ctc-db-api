package middleware

import (
	"context"
	"errors"
	"fmt"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/logging"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TimeoutMiddleware wraps request context with timeout and logs timeout events
// It also recovers from panics in the handler, records them in the span with stacktrace, and logs them
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

			// Run handler in goroutine with panic recovery
			go func() {
				defer func() {
					if r := recover(); r != nil {
						ctx := c.Request().Context()
						span := trace.SpanFromContext(ctx)

						// Capture stacktrace
						stacktrace := string(debug.Stack())

						// Convert panic value to error
						var panicErr error
						var panicType string
						if e, ok := r.(error); ok {
							panicErr = e
							panicType = fmt.Sprintf("%T", e)
						} else {
							panicErr = fmt.Errorf("%v", r)
							panicType = fmt.Sprintf("%T", r)
						}

						// Record error in span with stacktrace
						span.RecordError(panicErr, trace.WithAttributes(
							attribute.String("exception.stacktrace", stacktrace),
						))
						span.SetStatus(codes.Error, "panic recovered in timeout handler")
						span.SetAttributes(
							attribute.String("http.route", c.Path()),
							attribute.String("panic.type", panicType),
						)

						// Log panic with stacktrace
						logger.WithContext(ctx).Error("panic recovered in timeout handler",
							zap.String("panic", fmt.Sprintf("%v", r)),
							zap.String("panic.type", panicType),
							zap.String("http.method", c.Request().Method),
							zap.String("http.route", c.Path()),
						)

						// Send error through channel
						done <- panicErr
					}
				}()
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
