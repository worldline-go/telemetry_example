package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetricProvider open grpc connection and send metrics to the collector.
//
// global.SetMeterProvider(pusher)
//
// Shutdown after usage
// meterProvider.Shutdown(ctx)
func MetricProvider(ctx context.Context, url string) (*metricsdk.MeterProvider, error) {
	conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	meterProvider := metricsdk.NewMeterProvider(metricsdk.WithReader(metricsdk.NewPeriodicReader(metricExporter)))

	// metric.WithInstrumentationVersion(),
	// meterProvider.Meter(pc.Service, metric.WithSchemaURL(semconv.SchemaURL)).AsyncFloat64().Counter()

	// Shutdown will flush any remaining spans and shut down the exporter.
	return meterProvider, nil
}
