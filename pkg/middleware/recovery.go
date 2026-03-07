package middleware

import (
	"fmt"
	"net/http"

	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from non-handler panics and records them in the span with minimal logging
func RecoveryMiddleware(logger *logging.Logger) echo.MiddlewareFunc {
	showPanicDetails := helpers.EnvWithDefaultBool("SHOW_PANIC_DETAILS", false)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					ctx := c.Request().Context()
					span := trace.SpanFromContext(ctx)

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

					// Record error in span with attributes
					span.RecordError(panicErr)
					span.SetStatus(codes.Error, "panic recovered")
					span.SetAttributes(
						attribute.String("http.route", c.Path()),
						attribute.String("panic.type", panicType),
					)

					// Minimal logging at Error level (zap will add stacktrace)
					logger.WithContext(ctx).Error("panic recovered",
						zap.String("panic", fmt.Sprintf("%v", r)),
						zap.String("panic.type", panicType),
						zap.String("http.method", c.Request().Method),
						zap.String("http.route", c.Path()),
					)

					// Return error response with conditional detail
					if showPanicDetails {
						// Development: show panic details
						c.JSON(http.StatusInternalServerError, map[string]interface{}{
							"message": "internal server error",
							"error":   fmt.Sprintf("%v", r),
							"type":    panicType,
						})
					} else {
						// Production: generic message
						controller.InternalError(c, "internal server error")
					}
				}
			}()

			return next(c)
		}
	}
}
