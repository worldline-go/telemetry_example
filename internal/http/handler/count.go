package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/worldline-go/telemetry_example/pkg/hold"
	"github.com/worldline-go/telemetry_example/pkg/msg"
	"github.com/worldline-go/telemetry_example/pkg/telemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Handlers struct {
	Counter *hold.Counter
}

func (h *Handlers) Register(group *echo.Group, middlewares []echo.MiddlewareFunc) {
	group.GET("/count", h.GetCount, middlewares...)
	group.POST("/count", h.PostCount, middlewares...)
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
	_, span := otel.GetTracerProvider().Tracer(c.Path()).Start(c.Request().Context(), "GetCount")
	defer span.End()

	count := h.Counter.Get()
	// Store n as a string to not overflow an int64.
	span.SetAttributes(attribute.Int64("request.count.get", count))

	telemetry.GlobalMeter.UpDownCounter.Add(c.Request().Context(), 1, telemetry.GlobalAttr...)

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
	_, span := otel.GetTracerProvider().Tracer(c.Path()).Start(c.Request().Context(), "PostCount")
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

	span.SetAttributes(attribute.Key("request.count.set").Int64(countInt))

	telemetry.GlobalMeter.SuccessCounter.Add(c.Request().Context(), 1, telemetry.GlobalAttr...)
	telemetry.GlobalMeter.HistogramCounter.Record(c.Request().Context(), float64(countInt), telemetry.GlobalAttr...)

	newResult := h.Counter.Add(countInt)
	telemetry.WatchValue = newResult

	telemetry.GlobalMeter.UpDownCounter.Add(c.Request().Context(), 1, telemetry.GlobalAttr...)

	return c.JSON(http.StatusOK, msg.API{
		Data: newResult,
	})
}
