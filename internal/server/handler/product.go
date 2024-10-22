package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/worldline-go/telemetry_example/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// @Summary     Product to record
// @Description Message ping/pong
// @Accept      application/json
// @Produce     application/json
// @Param       product body model.Product true "Product to record"
// @Router      /products [POST]
// @Success     200 {object} model.Message{}
func (h *Handler) AddProduct(c echo.Context) error {
	ctx := context.WithoutCancel(c.Request().Context())

	var product model.Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	ctx, span := otel.Tracer("add_product").Start(ctx, "kafka", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	id, err := h.DB.AddNewProduct(ctx, product.Name, product.Description)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.Message{
		Message: "product added",
		Data:    id,
	})
}

// @Summary     Product to record
// @Description Message ping/pong
// @Accept      application/json
// @Produce     application/json
// @Param       name path string true "Product name"
// @Router      /products-push/{name} [POST]
// @Success     200 {object} model.Message{}
func (h *Handler) SendProduct(c echo.Context) error {
	ctx := context.WithoutCancel(c.Request().Context())

	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "name is required",
		})
	}

	product, err := h.DB.GetProduct(ctx, name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	ctx, span := otel.Tracer("kafka_product").Start(ctx, "request", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	if err := h.KafkaProducer.Produce(ctx, product); err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.Message{
		Message: "product sent",
		Data:    product,
	})
}
