package telemetry

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	GlobalAttr  []attribute.KeyValue
	GlobalMeter *Meter

	WatchValue int64
)

type Meter struct {
	SuccessCounter   instrument.Int64Counter
	HistogramCounter instrument.Float64Histogram
	UpDownCounter    instrument.Int64UpDownCounter
	SendGaugeCounter instrument.Int64ObservableGauge
}

func AddGlobalAttr(v ...attribute.KeyValue) {
	GlobalAttr = append(GlobalAttr, v...)
}

func SetGlobalMeter() error {
	mp := global.MeterProvider()

	m := &Meter{}

	var err error
	meter := mp.Meter("")

	m.SuccessCounter, err = meter.Int64Counter("count_success", instrument.WithDescription("number of success count"))
	if err != nil {
		return fmt.Errorf("failed to initialize validate_success; %w", err)
	}

	m.HistogramCounter, err = meter.Float64Histogram("count_histogram", instrument.WithDescription("value histogram"))
	if err != nil {
		return fmt.Errorf("failed to initialize valuehistogram; %w", err)
	}

	m.UpDownCounter, err = meter.Int64UpDownCounter("count_updown", instrument.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	m.SendGaugeCounter, err = meter.Int64ObservableGauge("send", instrument.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveInt64(m.SendGaugeCounter, WatchValue, GlobalAttr...)
		return nil
	}, m.SendGaugeCounter)

	if err != nil {
		log.Error().Err(err).Msg("failed to register up gauge metric")
	}

	GlobalMeter = m
	return nil
}
