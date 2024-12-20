package telemetry

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	GlobalAttr  []attribute.KeyValue
	GlobalMeter *Meter

	WatchValue int64
)

type Meter struct {
	SuccessCounter   metric.Int64Counter
	HistogramCounter metric.Float64Histogram
	UpDownCounter    metric.Int64UpDownCounter
	SendGaugeCounter metric.Int64ObservableGauge
}

func SetGlobalMeter() error {
	mp := otel.GetMeterProvider()

	m := &Meter{}

	var err error
	meter := mp.Meter("")

	m.SuccessCounter, err = meter.Int64Counter("count_success", metric.WithDescription("number of success count"))
	if err != nil {
		return fmt.Errorf("failed to initialize validate_success; %w", err)
	}

	m.HistogramCounter, err = meter.Float64Histogram("count_histogram", metric.WithDescription("value histogram"))
	if err != nil {
		return fmt.Errorf("failed to initialize valuehistogram; %w", err)
	}

	m.UpDownCounter, err = meter.Int64UpDownCounter("count_updown", metric.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	m.SendGaugeCounter, err = meter.Int64ObservableGauge("send", metric.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveInt64(m.SendGaugeCounter, WatchValue, metric.WithAttributes(GlobalAttr...))
		return nil
	}, m.SendGaugeCounter)

	if err != nil {
		log.Error().Err(err).Msg("failed to register up gauge metric")
	}

	GlobalMeter = m

	return nil
}
