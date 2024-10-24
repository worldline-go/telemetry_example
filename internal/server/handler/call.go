package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Call
//
// @Summary     Call API
// @Description Call an api with name
// @Tags        call
// @Produce     application/json
// @Accept      application/json
// @Param       data body model.Service true "message"
// @Param       service path string true "service name"
// @Router      /call/{service} [POST]
// @Success     200 {object} model.Message{}
// @Failure     400 {object} model.Message{}
func (h *Handler) Call(c echo.Context) error {
	var serviceBody model.Service
	if err := c.Bind(&serviceBody); err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "failed to bind request",
		})
	}

	tracer := otel.Tracer("")
	ctx, span := tracer.Start(c.Request().Context(), "call", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	service := c.Param("service")

	span.SetAttributes(attribute.String("request.call.service", service))
	span.SetAttributes(attribute.Bool("request.call.error", serviceBody.Error))

	if service == "" {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "service name is required",
		})
	}

	if _, ok := h.Clients[service]; !ok {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "service [" + service + "] not found",
		})
	}

	// call service
	messageFromService := model.Service{
		Message: serviceBody.Message,
		Error:   serviceBody.Error,
	}

	messageByte, err := json.Marshal(messageFromService)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: "failed to marshal message",
		})
	}

	request, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "/api/v1/message",
		bytes.NewReader(messageByte),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: "failed to create request",
		})
	}

	ctx, spanCall := tracer.Start(ctx, service, trace.WithSpanKind(trace.SpanKindClient))
	defer spanCall.End()

	// add context propagation
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))

	log.Info().Msgf("headers: %v", request.Header)

	responseMessage := model.Message{}

	if err := h.Clients[service].Do(request, klient.ResponseFuncJSON(&responseMessage)); err != nil {
		spanCall.SetStatus(codes.Error, err.Error())

		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, responseMessage)
}

// @Summary     Message to return
// @Description Message ping/pong
// @Tags        call
// @Accept      application/json
// @Produce     application/json
// @Param       data body model.Service true "message"
// @Router      /message [POST]
// @Success     200 {object} model.Message{}
// @Failure     400 {object} model.Message{}
func (h *Handler) Message(c echo.Context) error {
	var serviceBody model.Service
	if err := c.Bind(&serviceBody); err != nil {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "failed to bind request",
			Data:    err.Error(),
		})
	}

	log.Info().Msgf("headers: %v", c.Request().Header)
	_, span := otel.Tracer("").Start(c.Request().Context(), "message", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(attribute.String("request.message", serviceBody.Message))

	if serviceBody.Error {
		return c.JSON(http.StatusBadRequest, model.Message{
			Message: "error from service " + config.ServiceName,
		})
	}

	return c.JSON(http.StatusOK, model.Message{
		Message: "message from service " + config.ServiceName,
	})
}
