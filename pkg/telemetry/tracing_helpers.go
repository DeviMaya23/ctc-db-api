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

// StartRepositorySpan starts a span for a repository layer operation
func StartRepositorySpan(ctx context.Context, repositoryName, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(repositoryName)
	ctx, span := tracer.Start(ctx, operationName)
	span.SetAttributes(attribute.String("db.operation", operationName))
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
