package middleware

import (
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingMiddleware returns the OTel tracing middleware if enabled
func TracingMiddleware(logger *logging.Logger) echo.MiddlewareFunc {
	enabled := helpers.EnvWithDefaultBool("OTEL_ENABLED", false)

	if !enabled {
		logger.Info("OTel tracing middleware is disabled")
		// Return a no-op middleware
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	serviceName := helpers.EnvWithDefault("OTEL_SERVICE_NAME", "ctc-db-api")

	logger.Info("OTel tracing middleware enabled",
		zap.String("service.name", serviceName),
	)

	// otelecho.Middleware creates spans for each HTTP request
	baseMiddleware := otelecho.Middleware(serviceName)

	// Wrap otelecho middleware to add request ID to span
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return baseMiddleware(func(c echo.Context) error {
			// At this point, otelecho has created a span
			// Extract request ID from context and add to span
			ctx := c.Request().Context()
			if requestID := logging.GetRequestID(ctx); requestID != "" {
				span := trace.SpanFromContext(ctx)
				span.SetAttributes(attribute.String("http.request_id", requestID))
			}

			return next(c)
		})
	}
}
