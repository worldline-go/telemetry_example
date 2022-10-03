package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MetricProvider open grpc connection and send metrics to the collector.
//
// global.SetMeterProvider(pusher)
//
// Shutdown after usage
// meterProvider.Shutdown(ctx)
func (c *Collector) MetricCollector(ctx context.Context) (metricsdk.Reader, error) {
	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	reader := metricsdk.NewPeriodicReader(
		metricExporter, metricsdk.WithInterval(2*time.Second),
	)

	return reader, nil
}

func (c *Collector) MetricPrometheus() prometheus.Exporter {
	exporter := prometheus.New()

	return exporter
}

func (c *Collector) MetricProvider(mReaders ...metricsdk.Reader) *metricsdk.MeterProvider {
	// Set resource for auto show some attributes about this service
	// you can use resource.Default()
	// Set OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES
	options := []metricsdk.Option{metricsdk.WithResource(resource.Default())}

	for _, mReader := range mReaders {
		options = append(options, metricsdk.WithReader(mReader))
	}

	meterProvider := metricsdk.NewMeterProvider(
		options...,
	)

	// Shutdown will flush any remaining spans and shut down the exporter.
	return meterProvider
}
