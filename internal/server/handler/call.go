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
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Call
//
// @Summary     Call API
// @Description Call an api with name
// @Produce     application/json
// @Param       service path string true "service name"
// @Router      /call/{service} [POST]
// @Success     200 {object} model.Message{}
// @Failure     400 {object} model.Message{}
func (h *Handler) Call(c echo.Context) error {
	tracer := otel.Tracer(c.Path())
	ctx, span := tracer.Start(c.Request().Context(), "call")
	defer span.End()

	service := c.Param("service")

	span.SetAttributes(attribute.String("request.call.service", service))

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
	messageFromService := model.Message{
		Data: "message from service " + config.ServiceName,
	}

	messageByte, err := json.Marshal(messageFromService)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: "failed to marshal message",
		})
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost, "/api/v1/message",
		bytes.NewReader(messageByte),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: "failed to create request",
		})
	}

	ctx, spanCall := tracer.Start(ctx, "call-"+service, trace.WithSpanKind(trace.SpanKindClient))
	defer spanCall.End()

	// add context propagation
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))

	log.Info().Msgf("headers: %v", request.Header)

	responseMessage := model.Message{}

	if err := h.Clients[service].Do(request, func(r *http.Response) error {
		if err := klient.UnexpectedResponse(r); err != nil {
			return err
		}

		if err := json.NewDecoder(r.Body).Decode(&responseMessage); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, model.Message{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, responseMessage)
}

// @Summary     Message to return
// @Description Message ping/pong
// @Accept      application/json
// @Produce     application/json
// @Router      /message [POST]
// @Success     200 {object} model.Message{}
func (h *Handler) Message(c echo.Context) error {
	log.Info().Msgf("headers: %v", c.Request().Header)
	_, span := otel.Tracer(c.Path()).Start(c.Request().Context(), "message")
	defer span.End()

	return c.Stream(http.StatusOK, "application/json", c.Request().Body)
}
