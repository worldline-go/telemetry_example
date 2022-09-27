package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/pkg/hold"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/pkg/msg"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	Counter *hold.Counter
	Tracer  trace.TracerProvider
	Meter   metric.MeterProvider

	successCounter syncint64.Counter
}

func (h *Handlers) Register(group *echo.Group, middlewares []echo.MiddlewareFunc) {
	group.GET("/count", h.GetCount, middlewares...)

	var err error

	h.successCounter, err = h.Meter.Meter("").
		SyncInt64().Counter("success", instrument.WithDescription("number of success count"))
	if err != nil {
		log.Panic().Msgf("failed to initialize successCounter; %w", err)
	}
}

// GetCount
//
// @Summary     Get Count
// @Description Get Count
// @Produce     json
// @Router      /count [get]
// @Security    ApiKeyAuth
// @Param       count query int false "Count Value"
// @Success     200 {object} msg.WebApiSuccess{}
// @Failure     400 {object} msg.WebApiError{}
func (h *Handlers) GetCount(c echo.Context) error {
	_, span := h.Tracer.Tracer(c.Path()).Start(c.Request().Context(), "GetCount")
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
	// Store n as a string to not overflow an int64.
	span.SetAttributes(attribute.String("request.count", count))

	h.successCounter.Add(c.Request().Context(), 1)

	newResult := h.Counter.Add(countInt)

	return c.JSON(http.StatusOK, msg.API{
		Data: newResult,
	})
}
