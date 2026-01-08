package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional context-aware methods
type Logger struct {
	*zap.Logger
}

// NewLogger creates a logger based on environment
func NewLogger(env string) (*Logger, error) {
	if env == "production" {
		return NewProductionLogger()
	}
	return NewDevelopmentLogger()
}

// NewDevelopmentLogger creates a development logger with debug level and console output
func NewDevelopmentLogger() (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.WarnLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewProductionLogger creates a production logger with info level, JSON output, and sampling
func NewProductionLogger() (*Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: zapLogger}, nil
}

// WithContext enriches the logger with context information (request ID, user ID, trace IDs)
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}

	// Extract request ID
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("http.request_id", requestID))
	}

	// Extract user ID
	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user.id", userID))
	}

	// Extract trace context (OTel-ready, currently returns empty)
	if traceID := ExtractTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace.id", traceID))
	}

	if spanID := ExtractSpanID(ctx); spanID != "" {
		fields = append(fields, zap.String("span.id", spanID))
	}

	if len(fields) > 0 {
		return &Logger{Logger: l.Logger.With(fields...)}
	}

	return l
}

// Named creates a named logger (useful for sub-components)
func (l *Logger) Named(name string) *Logger {
	return &Logger{Logger: l.Logger.Named(name)}
}

// ExtractTraceID extracts trace ID from context using OpenTelemetry
func ExtractTraceID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// ExtractSpanID extracts span ID from context using OpenTelemetry
func ExtractSpanID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}
