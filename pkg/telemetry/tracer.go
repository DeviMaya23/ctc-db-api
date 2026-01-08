package telemetry

import (
	"context"
	"fmt"
	"time"

	"lizobly/ctc-db-api/pkg/helpers"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TracerProvider struct {
	provider *sdktrace.TracerProvider
	enabled  bool
}

// InitTracer initializes the OpenTelemetry tracer provider
func InitTracer(logger *zap.Logger) (*TracerProvider, error) {
	enabled := helpers.EnvWithDefaultBool("OTEL_ENABLED", false)

	if !enabled {
		logger.Info("OpenTelemetry tracing is disabled")
		return &TracerProvider{enabled: false}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(helpers.EnvWithDefault("OTEL_SERVICE_NAME", "ctc-db-api")),
			semconv.ServiceVersion(helpers.EnvWithDefault("OTEL_SERVICE_VERSION", "1.0.0")),
			semconv.DeploymentEnvironmentName(helpers.EnvWithDefault("OTEL_ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	endpoint := helpers.EnvWithDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")

	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use WithTLSClientConfig() in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create sampler based on configuration
	sampler := createSampler()

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator (for distributed tracing across services)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OpenTelemetry tracer initialized",
		zap.String("endpoint", endpoint),
		zap.String("service", helpers.EnvWithDefault("OTEL_SERVICE_NAME", "ctc-db-api")),
		zap.String("environment", helpers.EnvWithDefault("OTEL_ENVIRONMENT", "development")),
	)

	return &TracerProvider{
		provider: tp,
		enabled:  true,
	}, nil
}

// createSampler creates a sampler based on environment configuration
func createSampler() sdktrace.Sampler {
	samplerType := helpers.EnvWithDefault("OTEL_TRACES_SAMPLER", "always_on")

	switch samplerType {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		ratio := helpers.EnvWithDefaultFloat("OTEL_TRACES_SAMPLER_ARG", 1.0)
		return sdktrace.TraceIDRatioBased(ratio)
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	case "parentbased_traceidratio":
		ratio := helpers.EnvWithDefaultFloat("OTEL_TRACES_SAMPLER_ARG", 0.1)
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	default:
		return sdktrace.AlwaysSample()
	}
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if !tp.enabled || tp.provider == nil {
		return nil
	}

	return tp.provider.Shutdown(ctx)
}

// GetTracer returns a tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
