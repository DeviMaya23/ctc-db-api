package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartServiceSpan starts a span for a service layer operation
func StartServiceSpan(ctx context.Context, serviceName, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(serviceName)
	ctx, span := tracer.Start(ctx, operationName)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
	return ctx, span
}

// StartDBSpan starts a span for a database operation with common attributes
// operation examples: "select", "insert", "update", "delete"
func StartDBSpan(ctx context.Context, repositoryName, operationName, operation, tableName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(repositoryName)
	ctx, span := tracer.Start(ctx, operationName)

	// Set common database attributes
	span.SetAttributes(
		attribute.String("db.system", "postgres"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", tableName),
	)

	// Add any additional attributes
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return ctx, span
}

// EndSpanWithError ends a span and records an error if present
func EndSpanWithError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
