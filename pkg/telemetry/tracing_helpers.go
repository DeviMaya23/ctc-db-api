package telemetry

import (
	"context"
	"time"

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

// DBOperation tracks the lifecycle of a database operation for telemetry
type DBOperation struct {
	ctx       context.Context
	span      trace.Span
	startTime time.Time
	operation string
	table     string
}

// StartDBSpan starts a span and returns a DBOperation for tracking metrics
func StartDBSpan(ctx context.Context, repositoryName, operationName, operation, tableName string, attrs ...attribute.KeyValue) (context.Context, *DBOperation) {
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

	return ctx, &DBOperation{
		ctx:       ctx,
		span:      span,
		startTime: time.Now(),
		operation: operation,
		table:     tableName,
	}
}

// End concludes the operation, records duration metrics and any errors
func (op *DBOperation) End(err error) error {
	defer func() {
		if err != nil {
			op.span.RecordError(err)
			op.span.SetStatus(codes.Error, err.Error())
		} else {
			op.span.SetStatus(codes.Ok, "")
		}
		op.span.End()
	}()

	duration := time.Since(op.startTime)
	op.span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Milliseconds())))

	return err
}

// Duration returns the elapsed time since operation start
func (op *DBOperation) Duration() time.Duration {
	return time.Since(op.startTime)
}

// Context returns the operation context
func (op *DBOperation) Context() context.Context {
	return op.ctx
}

// Operation returns the operation type (select, insert, update, delete)
func (op *DBOperation) Operation() string {
	return op.operation
}

// Table returns the table name
func (op *DBOperation) Table() string {
	return op.table
}
