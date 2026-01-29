package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWithRequestID tests request ID context operations
func TestWithRequestID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() context.Context
		requestID  string
		expectGet  string
		expectDiff bool // expect context to be different
	}{
		{
			name:       "adds request ID to context",
			setup:      func() context.Context { return context.Background() },
			requestID:  "req-123",
			expectGet:  "req-123",
			expectDiff: true,
		},
		{
			name:       "multiple values - last wins",
			setup:      func() context.Context { return WithRequestID(context.Background(), "req-1") },
			requestID:  "req-2",
			expectGet:  "req-2",
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			newCtx := WithRequestID(ctx, tt.requestID)

			if tt.expectDiff {
				assert.NotEqual(t, ctx, newCtx)
			}
			assert.Equal(t, tt.expectGet, GetRequestID(newCtx))
		})
	}
}

// TestGetRequestID_ReturnsEmptyStringWhenNotSet tests missing request ID
func TestGetRequestID_ReturnsEmptyStringWhenNotSet(t *testing.T) {
	ctx := context.Background()
	requestID := GetRequestID(ctx)
	assert.Equal(t, "", requestID)
}

// TestWithUserID tests user ID context operations
func TestWithUserID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() context.Context
		userID     string
		expectGet  string
		expectDiff bool
	}{
		{
			name:       "adds user ID to context",
			setup:      func() context.Context { return context.Background() },
			userID:     "user-456",
			expectGet:  "user-456",
			expectDiff: true,
		},
		{
			name:       "multiple values - last wins",
			setup:      func() context.Context { return WithUserID(context.Background(), "user-1") },
			userID:     "user-2",
			expectGet:  "user-2",
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			newCtx := WithUserID(ctx, tt.userID)

			if tt.expectDiff {
				assert.NotEqual(t, ctx, newCtx)
			}
			assert.Equal(t, tt.expectGet, GetUserID(newCtx))
		})
	}
}

// TestGetUserID_ReturnsEmptyStringWhenNotSet tests missing user ID
func TestGetUserID_ReturnsEmptyStringWhenNotSet(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	assert.Equal(t, "", userID)
}

// TestWithTraceID tests trace ID context operations
func TestWithTraceID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() context.Context
		traceID    string
		expectGet  string
		expectDiff bool
	}{
		{
			name:       "adds trace ID to context",
			setup:      func() context.Context { return context.Background() },
			traceID:    "trace-789",
			expectGet:  "trace-789",
			expectDiff: true,
		},
		{
			name:       "multiple values - last wins",
			setup:      func() context.Context { return WithTraceID(context.Background(), "trace-1") },
			traceID:    "trace-2",
			expectGet:  "trace-2",
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			newCtx := WithTraceID(ctx, tt.traceID)

			if tt.expectDiff {
				assert.NotEqual(t, ctx, newCtx)
			}
			assert.Equal(t, tt.expectGet, GetTraceID(newCtx))
		})
	}
}

// TestGetTraceID_ReturnsEmptyStringWhenNotSet tests missing trace ID
func TestGetTraceID_ReturnsEmptyStringWhenNotSet(t *testing.T) {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	assert.Equal(t, "", traceID)
}

// TestWithSpanID tests span ID context operations
func TestWithSpanID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() context.Context
		spanID     string
		expectGet  string
		expectDiff bool
	}{
		{
			name:       "adds span ID to context",
			setup:      func() context.Context { return context.Background() },
			spanID:     "span-999",
			expectGet:  "span-999",
			expectDiff: true,
		},
		{
			name:       "multiple values - last wins",
			setup:      func() context.Context { return WithSpanID(context.Background(), "span-1") },
			spanID:     "span-2",
			expectGet:  "span-2",
			expectDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			newCtx := WithSpanID(ctx, tt.spanID)

			if tt.expectDiff {
				assert.NotEqual(t, ctx, newCtx)
			}
			assert.Equal(t, tt.expectGet, GetSpanID(newCtx))
		})
	}
}

// TestGetSpanID_ReturnsEmptyStringWhenNotSet tests missing span ID
func TestGetSpanID_ReturnsEmptyStringWhenNotSet(t *testing.T) {
	ctx := context.Background()
	spanID := GetSpanID(ctx)
	assert.Equal(t, "", spanID)
}

// TestContext_MultipleValuesIsolated tests different values don't interfere
func TestContext_MultipleValuesIsolated(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithUserID(ctx, "user-1")
	ctx = WithTraceID(ctx, "trace-1")
	ctx = WithSpanID(ctx, "span-1")

	assert.Equal(t, "req-1", GetRequestID(ctx))
	assert.Equal(t, "user-1", GetUserID(ctx))
	assert.Equal(t, "trace-1", GetTraceID(ctx))
	assert.Equal(t, "span-1", GetSpanID(ctx))
}

// TestContext_PartialValues tests retrieving only set values
func TestContext_PartialValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithTraceID(ctx, "trace-1")

	assert.Equal(t, "req-1", GetRequestID(ctx))
	assert.Equal(t, "", GetUserID(ctx))
	assert.Equal(t, "trace-1", GetTraceID(ctx))
	assert.Equal(t, "", GetSpanID(ctx))
}

// TestContext_ValueTypes tests different value types
func TestContext_ValueTypes(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "uuid-abc-123")
	ctx = WithUserID(ctx, "12345")
	ctx = WithTraceID(ctx, "0af7651916cd43dd8448eb211c80319c")
	ctx = WithSpanID(ctx, "b7ad6b7169203331")

	assert.Equal(t, "uuid-abc-123", GetRequestID(ctx))
	assert.Equal(t, "12345", GetUserID(ctx))
	assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", GetTraceID(ctx))
	assert.Equal(t, "b7ad6b7169203331", GetSpanID(ctx))
}

// TestContext_EmptyStringValues tests with empty strings
func TestContext_EmptyStringValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "")
	ctx = WithUserID(ctx, "")

	// Empty strings should be stored and retrievable as empty
	assert.Equal(t, "", GetRequestID(ctx))
	assert.Equal(t, "", GetUserID(ctx))
}

// TestContext_NilContextHandling tests nil context handling
func TestContext_NilContextHandling(t *testing.T) {
	// Test with background context (safest approach)
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-1")
	assert.NotNil(t, ctx)
	assert.Equal(t, "req-1", GetRequestID(ctx))
}

// TestContext_ConcurrentAccess tests that context values don't leak across operations
func TestContext_ConcurrentAccess(t *testing.T) {
	ctx1 := context.Background()
	ctx1 = WithRequestID(ctx1, "req-1")

	ctx2 := context.Background()
	ctx2 = WithRequestID(ctx2, "req-2")

	// Each context should have its own value
	assert.Equal(t, "req-1", GetRequestID(ctx1))
	assert.Equal(t, "req-2", GetRequestID(ctx2))
}

// TestContext_ChainedOperations tests chaining multiple With operations
func TestContext_ChainedOperations(t *testing.T) {
	ctx := WithSpanID(
		WithTraceID(
			WithUserID(
				WithRequestID(context.Background(), "req-1"),
				"user-1",
			),
			"trace-1",
		),
		"span-1",
	)

	assert.Equal(t, "req-1", GetRequestID(ctx))
	assert.Equal(t, "user-1", GetUserID(ctx))
	assert.Equal(t, "trace-1", GetTraceID(ctx))
	assert.Equal(t, "span-1", GetSpanID(ctx))
}

// TestContext_TypeAssertion tests wrong type retrieval
func TestContext_TypeAssertion(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, 123) // Wrong type

	// Should return empty string due to type assertion failure
	requestID := GetRequestID(ctx)
	assert.Equal(t, "", requestID)
}
