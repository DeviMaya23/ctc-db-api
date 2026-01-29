package telemetry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

// TestStartServiceSpan tests service span creation with various configurations
func TestStartServiceSpan(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		operationName string
		attributes    []attribute.KeyValue
	}{
		{
			name:          "returns context and span",
			serviceName:   "test-service",
			operationName: "test-operation",
			attributes:    nil,
		},
		{
			name:          "with attributes",
			serviceName:   "test-service",
			operationName: "test-operation",
			attributes: []attribute.KeyValue{
				attribute.String("user_id", "123"),
				attribute.Int("status_code", 200),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx, span := StartServiceSpan(ctx, tt.serviceName, tt.operationName, tt.attributes...)
			assert.NotNil(t, newCtx)
			assert.NotNil(t, span)
			defer span.End()
		})
	}
}

// TestEndSpanWithError tests span ending with various error conditions
func TestEndSpanWithError(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name:  "with no error",
			error: nil,
		},
		{
			name:  "with error",
			error: errors.New("test error"),
		},
		{
			name:  "with another error",
			error: errors.New("another error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, span := StartServiceSpan(ctx, "test-service", "test-operation")
			// Should not panic and should record error if present
			EndSpanWithError(span, tt.error)
		})
	}
}

// TestStartDBSpan tests database span creation with various configurations
func TestStartDBSpan(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		operationName  string
		operation      string
		tableName      string
		attributes     []attribute.KeyValue
	}{
		{
			name:           "returns context and DB operation",
			repositoryName: "user-repository",
			operationName:  "get-user",
			operation:      "select",
			tableName:      "users",
			attributes:     nil,
		},
		{
			name:           "with attributes",
			repositoryName: "user-repository",
			operationName:  "get-user",
			operation:      "select",
			tableName:      "users",
			attributes: []attribute.KeyValue{
				attribute.Int("row_count", 10),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx, dbOp := StartDBSpan(ctx, tt.repositoryName, tt.operationName, tt.operation, tt.tableName, tt.attributes...)
			assert.NotNil(t, newCtx)
			assert.NotNil(t, dbOp)
			assert.Equal(t, tt.operation, dbOp.Operation())
			dbOp.End(nil)
		})
	}
}

// TestDBOperation_End tests DBOperation end with various error conditions
func TestDBOperation_End(t *testing.T) {
	tests := []struct {
		name      string
		error     error
		expectErr bool
	}{
		{
			name:      "without error",
			error:     nil,
			expectErr: false,
		},
		{
			name:      "with error",
			error:     errors.New("database error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, dbOp := StartDBSpan(ctx, "test-repo", "test-op", "select", "test_table")

			err := dbOp.End(tt.error)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.error, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDBOperation_Duration tests duration measurement with various timings
func TestDBOperation_Duration(t *testing.T) {
	tests := []struct {
		name        string
		sleepTime   time.Duration
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			name:        "calculates elapsed time",
			sleepTime:   10 * time.Millisecond,
			minDuration: 10 * time.Millisecond,
			maxDuration: 500 * time.Millisecond,
		},
		{
			name:        "accurate measurement",
			sleepTime:   50 * time.Millisecond,
			minDuration: 50 * time.Millisecond,
			maxDuration: 200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, dbOp := StartDBSpan(ctx, "test-repo", "test-op", "select", "test_table")

			time.Sleep(tt.sleepTime)
			duration := dbOp.Duration()

			assert.True(t, duration >= tt.minDuration)
			assert.True(t, duration < tt.maxDuration)

			dbOp.End(nil)
		})
	}
}

// TestDBOperation_Context tests context retrieval
func TestDBOperation_Context(t *testing.T) {
	ctx := context.Background()
	newCtx, dbOp := StartDBSpan(ctx, "test-repo", "test-op", "select", "test_table")

	retrievedCtx := dbOp.Context()
	assert.Equal(t, newCtx, retrievedCtx)

	dbOp.End(nil)
}

// TestStartServiceSpan_ContextPropagation tests context is properly propagated
func TestStartServiceSpan_ContextPropagation(t *testing.T) {
	baseCtx := context.Background()

	ctx1, span1 := StartServiceSpan(baseCtx, "service-1", "operation-1")
	defer span1.End()

	// Context should be different from base context
	assert.NotEqual(t, baseCtx, ctx1)

	// Should be able to use the new context
	ctx2, span2 := StartServiceSpan(ctx1, "service-2", "operation-2")
	defer span2.End()

	assert.NotEqual(t, ctx1, ctx2)
}
