package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestInitTracer_Disabled tests tracer initialization when OTEL_ENABLED is false
func TestInitTracer_Disabled(t *testing.T) {
	// Set environment variable to disable tracing
	t.Setenv("OTEL_ENABLED", "false")

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	defer logger.Sync()

	tp, err := InitTracer(logger)
	assert.NoError(t, err)
	assert.NotNil(t, tp)
	assert.False(t, tp.enabled)
	assert.Nil(t, tp.provider)
}

// TestInitTracer_Enabled tests tracer initialization when OTEL_ENABLED is true
func TestInitTracer_Enabled(t *testing.T) {
	// Set environment variables for enabled mode
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "test-service")
	t.Setenv("OTEL_SERVICE_VERSION", "1.0.0")
	t.Setenv("OTEL_ENVIRONMENT", "test")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	defer logger.Sync()

	tp, err := InitTracer(logger)
	assert.NoError(t, err)
	assert.NotNil(t, tp)
	assert.True(t, tp.enabled)
	assert.NotNil(t, tp.provider)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = tp.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestInitTracer_DefaultValues tests tracer uses default values when env vars are not set
func TestInitTracer_DefaultValues(t *testing.T) {
	t.Setenv("OTEL_ENABLED", "true")
	// Unset other variables to test defaults
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_SERVICE_VERSION", "")
	t.Setenv("OTEL_ENVIRONMENT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	defer logger.Sync()

	tp, err := InitTracer(logger)
	assert.NoError(t, err)
	assert.NotNil(t, tp)
	assert.True(t, tp.enabled)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = tp.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestCreateSampler tests sampler creation with various configurations
func TestCreateSampler(t *testing.T) {
	tests := []struct {
		name       string
		sampler    string
		samplerArg string
	}{
		{
			name:    "always on",
			sampler: "always_on",
		},
		{
			name:    "always off",
			sampler: "always_off",
		},
		{
			name:       "trace ID ratio",
			sampler:    "traceidratio",
			samplerArg: "0.5",
		},
		{
			name:    "parent based always on",
			sampler: "parentbased_always_on",
		},
		{
			name:       "parent based trace ID ratio",
			sampler:    "parentbased_traceidratio",
			samplerArg: "0.1",
		},
		{
			name:    "unknown sampler - uses default",
			sampler: "unknown_sampler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_TRACES_SAMPLER", tt.sampler)
			if tt.samplerArg != "" {
				t.Setenv("OTEL_TRACES_SAMPLER_ARG", tt.samplerArg)
			}

			sampler := createSampler()
			assert.NotNil(t, sampler)
		})
	}
}

// TestTracerProvider_Shutdown_Disabled tests Shutdown when tracing is disabled
func TestTracerProvider_Shutdown_Disabled(t *testing.T) {
	tp := &TracerProvider{
		enabled:  false,
		provider: nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := tp.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestTracerProvider_Shutdown_Enabled tests Shutdown when tracing is enabled
func TestTracerProvider_Shutdown_Enabled(t *testing.T) {
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "test-service")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")

	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	defer logger.Sync()

	tp, err := InitTracer(logger)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tp.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestGetTracer tests tracer retrieval
func TestGetTracer(t *testing.T) {
	tracer := GetTracer("test-service")
	assert.NotNil(t, tracer)
}
