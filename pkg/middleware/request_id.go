package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"lizobly/cotc-db-api/pkg/helpers"
	"lizobly/cotc-db-api/pkg/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// responseBodyWriter wraps echo.ResponseWriter to capture response body
type responseBodyWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // Capture body
	return w.ResponseWriter.Write(b)
}

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

			// Capture request body size (OTel standard)
			var requestBodySize int64
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil {
					requestBodySize = int64(len(bodyBytes))
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Log request start
			logger.WithContext(ctx).Info("request started",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.String("http.request_id", requestID),
				zap.Int64("http.request.body.size", requestBodySize),
			)

			// Wrap response writer to capture response body
			blw := &responseBodyWriter{
				ResponseWriter: res.Writer,
				body:           new(bytes.Buffer),
			}
			res.Writer = blw

			// Log request body in development mode only
			env := helpers.EnvWithDefault("ENVIRONMENT", "development")
			logRequestBody := helpers.EnvWithDefaultBool("LOG_REQUEST_BODY", false)
			if env == "development" && logRequestBody {
				logRequestBodyIfEnabled(c, logger, ctx, requestBodySize)
			}

			// Call next handler
			err := next(c)

			// Calculate duration and response body size
			duration := time.Since(start)
			responseBodySize := int64(blw.body.Len())

			// Log request completion
			logger.WithContext(ctx).Info("request completed",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.Int("http.status_code", res.Status),
				zap.String("http.request_id", requestID),
				zap.Float64("http.request.duration", float64(duration.Milliseconds())),
				zap.Int64("http.request.body.size", requestBodySize),
				zap.Int64("http.response.body.size", responseBodySize),
			)

			// Log response body in development mode only
			if env == "development" && logRequestBody {
				logResponseBodyIfEnabled(blw.body.Bytes(), logger, ctx, responseBodySize)
			}

			return err
		}
	}
}

// logRequestBodyIfEnabled logs the request body in development mode
func logRequestBodyIfEnabled(c echo.Context, logger *logging.Logger, ctx context.Context, bodySize int64) {
	req := c.Request()

	// Only log if there's a body
	if req.Body == nil || bodySize == 0 {
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
		logger.WithContext(ctx).Debug("request body captured",
			zap.Any("app.request.body", filteredBody),
		)
	} else {
		// Not JSON, log as string
		logger.WithContext(ctx).Debug("request body captured",
			zap.String("app.request.body", bodyStr),
		)
	}
}

// logResponseBodyIfEnabled logs the response body in development mode
func logResponseBodyIfEnabled(bodyBytes []byte, logger *logging.Logger, ctx context.Context, bodySize int64) {
	// Only log if there's a body
	if bodySize == 0 {
		return
	}

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
		logger.WithContext(ctx).Debug("response body captured",
			zap.Any("app.response.body", filteredBody),
		)
	} else {
		// Not JSON, log as string
		logger.WithContext(ctx).Debug("response body captured",
			zap.String("app.response.body", bodyStr),
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
