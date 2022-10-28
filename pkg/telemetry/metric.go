package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

var (
	GlobalAttr  []attribute.KeyValue
	GlobalMeter *Meter

	WatchValue int64
)

type Meter struct {
	SuccessCounter   syncint64.Counter
	HistogramCounter syncfloat64.Histogram
	UpDownCounter    syncint64.Counter
	SendGaugeCounter asyncint64.Gauge
}

func AddGlobalAttr(v ...attribute.KeyValue) {
	GlobalAttr = append(GlobalAttr, v...)
}

func SetGlobalMeter() error {
	mp := global.MeterProvider()

	m := &Meter{}

	var err error

	m.SuccessCounter, err = mp.Meter("").SyncInt64().Counter("count_success", instrument.WithDescription("number of success count"))
	if err != nil {
		return fmt.Errorf("failed to initialize validate_success; %w", err)
	}

	m.HistogramCounter, err = mp.Meter("").
		SyncFloat64().Histogram("count_histogram", instrument.WithDescription("value histogram"))
	if err != nil {
		return fmt.Errorf("failed to initialize valuehistogram; %w", err)
	}

	m.UpDownCounter, err = mp.Meter("").SyncInt64().UpDownCounter("count_updown", instrument.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	m.SendGaugeCounter, err = mp.Meter("").AsyncInt64().Gauge("send", instrument.WithDescription("async gauge"))
	if err != nil {
		return fmt.Errorf("failed to initialize sendGauge; %w", err)
	}

	mp.Meter("").RegisterCallback([]instrument.Asynchronous{m.SendGaugeCounter}, func(ctx context.Context) {
		m.SendGaugeCounter.Observe(ctx, WatchValue, GlobalAttr...)
	})

	GlobalMeter = m
	return nil
}
