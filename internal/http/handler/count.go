package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/pkg/hold"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/pkg/msg"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	Counter *hold.Counter
	Tracer  trace.TracerProvider
	Meter   metric.MeterProvider

	metrics struct {
		successCounter syncint64.Counter
		counterUpDown  syncint64.UpDownCounter
		sendGauge      asyncint64.Gauge
		valuehistogram syncfloat64.Histogram

		commonAttributes []attribute.KeyValue

		RegisterCallback func(insts []instrument.Asynchronous, function func(context.Context))

		currentValue int64
	}
}

func (h *Handlers) Register(group *echo.Group, middlewares []echo.MiddlewareFunc) {
	group.GET("/count", h.GetCount, middlewares...)
	group.POST("/count", h.PostCount, middlewares...)

	var err error

	h.metrics.successCounter, err = h.Meter.Meter("").
		SyncInt64().Counter("success", instrument.WithDescription("number of success count"))
	if err != nil {
		log.Panic().Msgf("failed to initialize successCounter; %w", err)
	}

	h.metrics.valuehistogram, err = h.Meter.Meter("").
		SyncFloat64().Histogram("histogram", instrument.WithDescription("value histogram"))
	if err != nil {
		log.Panic().Msgf("failed to initialize valuehistogram; %w", err)
	}

	h.metrics.counterUpDown, err = h.Meter.Meter("").SyncInt64().UpDownCounter("updown", instrument.WithDescription("async gauge"))
	if err != nil {
		log.Panic().Msgf("failed to initialize sendGauge; %w", err)
	}

	h.metrics.sendGauge, err = h.Meter.Meter("").AsyncInt64().Gauge("send", instrument.WithDescription("async gauge"))
	if err != nil {
		log.Panic().Msgf("failed to initialize sendGauge; %w", err)
	}

	h.Meter.Meter("").RegisterCallback([]instrument.Asynchronous{h.metrics.sendGauge}, func(ctx context.Context) {
		h.metrics.sendGauge.Observe(ctx, h.metrics.currentValue, h.metrics.commonAttributes...)
	})

	h.metrics.commonAttributes = append(h.metrics.commonAttributes, attribute.Key("special").String("X"))
}

// GetCount
//
// @Summary     Get Count
// @Description Get Count
// @Produce     json
// @Router      /count [get]
// @Security    ApiKeyAuth
// @Success     200 {object} msg.WebApiSuccess{}
// @Failure     400 {object} msg.WebApiError{}
func (h *Handlers) GetCount(c echo.Context) error {
	_, span := h.Tracer.Tracer(c.Path()).Start(c.Request().Context(), "GetCount")
	defer span.End()

	count := h.Counter.Get()
	// Store n as a string to not overflow an int64.
	span.SetAttributes(attribute.Int64("request.count.get", count))

	h.metrics.counterUpDown.Add(c.Request().Context(), 1, h.metrics.commonAttributes...)

	return c.JSON(http.StatusOK, msg.API{
		Data: h.Counter.Get(),
	})
}

// PostCount
//
// @Summary     Add new count
// @Description Add new count
// @Produce     json
// @Router      /count [post]
// @Security    ApiKeyAuth
// @Param       count query int false "Count Value"
// @Success     200 {object} msg.WebApiSuccess{}
// @Failure     400 {object} msg.WebApiError{}
func (h *Handlers) PostCount(c echo.Context) error {
	_, span := h.Tracer.Tracer(c.Path()).Start(c.Request().Context(), "PostCount")
	defer span.End()

	countInt := int64(0)
	count := c.QueryParam("count")

	if count != "" {
		var err error

		countInt, err = strconv.ParseInt(count, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, msg.API{
				Err: err.Error(),
			})
		}
	}

	span.SetAttributes(attribute.Int64("request.count.set", countInt))

	h.metrics.successCounter.Add(c.Request().Context(), 1, h.metrics.commonAttributes...)

	h.metrics.valuehistogram.Record(c.Request().Context(), float64(countInt), h.metrics.commonAttributes...)

	newResult := h.Counter.Add(countInt)

	h.metrics.currentValue = newResult

	h.metrics.counterUpDown.Add(c.Request().Context(), 1, h.metrics.commonAttributes...)

	return c.JSON(http.StatusOK, msg.API{
		Data: newResult,
	})
}
