package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Get provider and don't forget to shutdown after usage, it will help to flush last messages.
// Use OTPL provider, it is more general and most of tools supporting and it is going to be standard.
//
// Also you can set globally this providor
// otel.SetTracerProvider(tracerProvider)
// set global propagator to tracecontext (the default is no-op).
// otel.SetTextMapPropagator(propagation.TraceContext{})

type ProviderConfig struct {
	Service     string
	Environment string
	ID          string
}

func TracerProvider(ctx context.Context, url string, pc ProviderConfig) (*tracesdk.TracerProvider, error) {
	conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)

	// to much resources not need all of them this is just for example
	res, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithContainerID(),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithOSDescription(),
		resource.WithOSType(),
		resource.WithProcess(),
		resource.WithProcessCommandArgs(),
		resource.WithProcessExecutableName(),
		resource.WithProcessExecutablePath(),
		resource.WithProcessOwner(),
		resource.WithProcessPID(),
		resource.WithProcessRuntimeDescription(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithTelemetrySDK(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(pc.Service),
			attribute.String("environment", pc.Environment),
			attribute.String("ID", pc.ID),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed resource new; %w", err)
	}

	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(res),
		tracesdk.WithSpanProcessor(bsp),
	)

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider, nil
}

// TracerProviderJaeger to export directly jaeger http://localhost:14268/api/traces
// Don't use this one.
func TracerProviderJaeger(ctx context.Context, url string, pc ProviderConfig) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(pc.Service),
			attribute.String("environment", pc.Environment),
			attribute.String("ID", pc.ID),
		)),
	)

	return tp, nil
}
