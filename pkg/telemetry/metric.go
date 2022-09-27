package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

// MetricProvider open grpc connection and send metrics to the collector.
//
// global.SetMeterProvider(pusher)
//
// Shutdown after usage
// meterProvider.Shutdown(ctx)
func (c *Collector) MetricProvider(ctx context.Context) (*metricsdk.MeterProvider, error) {
	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	meterProvider := metricsdk.NewMeterProvider(
		metricsdk.WithReader(
			metricsdk.NewPeriodicReader(
				metricExporter, metricsdk.WithInterval(2*time.Second),
			),
		),
	)

	// metric.WithInstrumentationVersion(),
	// meterProvider.Meter(pc.Service, metric.WithSchemaURL(semconv.SchemaURL)).AsyncFloat64().Counter()

	// Shutdown will flush any remaining spans and shut down the exporter.
	return meterProvider, nil
}
