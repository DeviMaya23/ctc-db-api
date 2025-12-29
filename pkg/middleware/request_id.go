package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"lizobly/cotc-db-api/pkg/helpers"
	"lizobly/cotc-db-api/pkg/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestIDMiddleware generates or extracts request IDs and logs HTTP requests
func RequestIDMiddleware(logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Extract or generate request ID
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Inject request ID into context
			ctx := logging.WithRequestID(req.Context(), requestID)
			c.SetRequest(req.WithContext(ctx))

			// Set response header
			res.Header().Set("X-Request-ID", requestID)

			// Log request start
			logger.WithContext(ctx).Info("request started",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.String("http.request_id", requestID),
			)

			// Log request body in development mode only
			env := helpers.EnvWithDefault("ENVIRONMENT", "development")
			logRequestBody := helpers.EnvWithDefaultBool("LOG_REQUEST_BODY", false)
			if env == "development" && logRequestBody {
				logRequestBodyIfEnabled(c, logger, ctx)
			}

			// Call next handler
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Log request completion
			logger.WithContext(ctx).Info("request completed",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.Int("http.status_code", res.Status),
				zap.String("http.request_id", requestID),
				zap.Float64("duration_ms", float64(duration.Milliseconds())),
			)

			return err
		}
	}
}

// logRequestBodyIfEnabled logs the request body in development mode
func logRequestBodyIfEnabled(c echo.Context, logger *logging.Logger, ctx context.Context) {
	req := c.Request()

	// Only log if there's a body
	if req.Body == nil {
		return
	}

	// Read body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		logger.WithContext(ctx).Warn("failed to read request body for logging",
			zap.Error(err),
		)
		return
	}

	// Restore body for actual handler
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Truncate large bodies
	maxBodySize := 1024 // 1KB
	bodyStr := string(bodyBytes)
	if len(bodyStr) > maxBodySize {
		bodyStr = bodyStr[:maxBodySize] + "... (truncated)"
	}

	// Try to parse as JSON for better formatting
	var bodyJSON map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &bodyJSON); err == nil {
		// Filter sensitive fields
		filteredBody := filterSensitiveFields(bodyJSON)
		logger.WithContext(ctx).Debug("request body",
			zap.Any("body", filteredBody),
		)
	} else {
		// Not JSON, log as string
		logger.WithContext(ctx).Debug("request body",
			zap.String("body", bodyStr),
		)
	}
}

// filterSensitiveFields removes sensitive data from logs
func filterSensitiveFields(body map[string]interface{}) map[string]interface{} {
	sensitiveFields := []string{"password", "token", "secret", "api_key", "apikey"}

	filtered := make(map[string]interface{})
	for key, value := range body {
		// Check if field is sensitive
		isSensitive := false
		for _, sensitive := range sensitiveFields {
			if key == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			filtered[key] = "***REDACTED***"
		} else {
			filtered[key] = value
		}
	}

	return filtered
}
