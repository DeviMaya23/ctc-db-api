package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"

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

// RequestBodyLoggingMiddleware logs HTTP request/response bodies and metadata.
func RequestBodyLoggingMiddleware(logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Capture request body size (OTel standard)
			var requestBodySize int64
			var requestBodyBytes []byte
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err == nil {
					requestBodyBytes = bodyBytes
					requestBodySize = int64(len(bodyBytes))
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Update request on Echo context with restored body
			c.SetRequest(req.WithContext(req.Context()))

			// Log request start
			logger.WithContext(req.Context()).Info("request started",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.Int64("http.request.body.size", requestBodySize),
			)

			// Wrap response writer to capture response body
			blw := &responseBodyWriter{
				ResponseWriter: res.Writer,
				body:           new(bytes.Buffer),
			}
			res.Writer = blw

			// Log request body when enabled
			logRequestBody := helpers.EnvWithDefaultBool("LOG_REQUEST_BODY", false)
			if logRequestBody {
				logRequestBodyContent(requestBodyBytes, logger, req.Context())
			}

			// Call next handler
			err := next(c)

			// Calculate duration and response body size
			duration := time.Since(start)
			responseBodySize := int64(blw.body.Len())

			// Log request completion
			logger.WithContext(req.Context()).Info("request completed",
				zap.String("http.method", req.Method),
				zap.String("http.route", c.Path()),
				zap.Int("http.status_code", res.Status),
				zap.Float64("http.request.duration", float64(duration.Milliseconds())),
				zap.Int64("http.request.body.size", requestBodySize),
				zap.Int64("http.response.body.size", responseBodySize),
			)

			// Log response body when enabled
			if logRequestBody {
				logResponseBodyIfEnabled(blw.body.Bytes(), logger, req.Context(), responseBodySize)
			}

			return err
		}
	}
}

// logRequestBodyContent logs the request body content when enabled
func logRequestBodyContent(bodyBytes []byte, logger *logging.Logger, ctx context.Context) {
	// Only log if there's a body
	if len(bodyBytes) == 0 {
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
		filteredBody := logging.FilterSensitiveFields(bodyJSON)
		logger.WithContext(ctx).Info("request body captured",
			zap.Any("app.request.body", filteredBody),
		)
	} else {
		// Not JSON, log as string
		logger.WithContext(ctx).Info("request body captured",
			zap.String("app.request.body", bodyStr),
		)
	}
}

// logResponseBodyIfEnabled logs the response body when enabled
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
		filteredBody := logging.FilterSensitiveFields(bodyJSON)
		logger.WithContext(ctx).Info("response body captured",
			zap.Any("app.response.body", filteredBody),
		)
	} else {
		// Not JSON, log as string
		logger.WithContext(ctx).Info("response body captured",
			zap.String("app.response.body", bodyStr),
		)
	}
}
