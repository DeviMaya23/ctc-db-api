package middleware

import (
	"lizobly/ctc-db-api/pkg/logging"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDMiddleware generates or extracts request IDs and injects them into the request context.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Extract or generate request ID
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Inject request ID into context
			ctx := logging.WithRequestID(req.Context(), requestID)

			// Set response header
			res.Header().Set("X-Request-ID", requestID)

			// Update request on Echo context with new context
			c.SetRequest(req.WithContext(ctx))

			// Call next handler
			return next(c)
		}
	}
}
