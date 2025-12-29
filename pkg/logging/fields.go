package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// HTTPFields returns OTel-compliant fields for HTTP requests
func HTTPFields(method, route string, statusCode int) []zap.Field {
	return []zap.Field{
		zap.String("http.method", method),
		zap.String("http.route", route),
		zap.Int("http.status_code", statusCode),
	}
}

// ErrorFields returns OTel-compliant fields for errors
func ErrorFields(err error) []zap.Field {
	if err == nil {
		return []zap.Field{}
	}

	return []zap.Field{
		zap.String("error.message", err.Error()),
		zap.String("error.type", "error"),
	}
}

// DatabaseFields returns OTel-compliant fields for database operations
func DatabaseFields(operation, table string, duration time.Duration) []zap.Field {
	return []zap.Field{
		zap.String("db.system", "postgres"),
		zap.String("db.operation", operation),
		zap.String("db.table", table),
		zap.Float64("db.duration_ms", float64(duration.Milliseconds())),
	}
}

// UserFields returns OTel-compliant fields for user context
func UserFields(userID, username string) []zap.Field {
	fields := []zap.Field{}

	if userID != "" {
		fields = append(fields, zap.String("user.id", userID))
	}

	if username != "" {
		fields = append(fields, zap.String("user.username", username))
	}

	return fields
}

// TraceFields returns OTel-compliant fields for trace context
// Currently returns empty fields, will populate when OTel is integrated
func TraceFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{}

	// Extract trace ID (currently stubbed, will use OTel later)
	if traceID := ExtractTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace.id", traceID))
	}

	// Extract span ID (currently stubbed, will use OTel later)
	if spanID := ExtractSpanID(ctx); spanID != "" {
		fields = append(fields, zap.String("span.id", spanID))
	}

	return fields
}
