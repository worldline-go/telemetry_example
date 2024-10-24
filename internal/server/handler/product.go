package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/worldline-go/telemetry_example/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// @Summary     Add new product
// @Description Add new product
// @Tags        products
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

	ctx, span := otel.Tracer("").Start(ctx,
		"add_product",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.name", "postgres|products")),
	)
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

// @Summary     Get product
// @Description Get product with name
// @Tags        products
// @Produce     application/json
// @Param       name path string true "Product name"
// @Router      /products/{name} [GET]
// @Success     200 {object} model.Message{}
func (h *Handler) GetProduct(c echo.Context) error {
	productName := c.Param("name")
	if productName == "" {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "name is required",
		})
	}

	ctx := context.WithoutCancel(c.Request().Context())

	ctx, span := otel.Tracer("").Start(ctx,
		"get_product",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.name", "postgres|products")),
	)
	defer span.End()

	product, err := h.DB.GetProduct(ctx, productName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.Message{
		Data: product,
	})
}

// @Summary     Product to record kafka
// @Tags        products
// @Description Send product to kafka
// @Accept      application/json
// @Produce     application/json
// @Param       name path string true "Product name"
// @Router      /products-send/{name} [POST]
// @Success     200 {object} model.Message{}
func (h *Handler) SendProduct(c echo.Context) error {
	ctx := context.WithoutCancel(c.Request().Context())

	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "name is required",
		})
	}

	ctxDB, spanDB := otel.Tracer("").Start(ctx,
		"get_product",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.name", "postgres|products")),
	)
	defer spanDB.End()

	product, err := h.DB.GetProduct(ctxDB, name)
	if err != nil {
		spanDB.SetStatus(codes.Error, err.Error())

		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	spanDB.End()

	ctx, spanKafka := otel.Tracer("").Start(ctx, "produce_message", trace.WithSpanKind(trace.SpanKindProducer))
	defer spanKafka.End()

	time.Sleep(5 * time.Millisecond)

	if err := h.KafkaProducer.Produce(ctx, product); err != nil {
		spanKafka.SetStatus(codes.Error, err.Error())

		return c.JSON(http.StatusBadRequest, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.Message{
		Message: "product sent",
		Data:    product,
	})
}
