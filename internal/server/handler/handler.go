package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/twmb/franz-go/plugin/kotel"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/wkafka"

	"github.com/worldline-go/telemetry_example/internal/database/dbhandler"
	"github.com/worldline-go/telemetry_example/internal/hold"
	"github.com/worldline-go/telemetry_example/internal/model"
)

type Handler struct {
	Counter       *hold.Counter
	Clients       map[string]*klient.Client
	KafkaProducer *wkafka.Producer[*model.Product]
	KafkaTracer   *kotel.Tracer
	DB            *dbhandler.Handler
}

func (h *Handler) Register(group *echo.Group) {
	group.GET("/count", h.GetCount)
	group.POST("/count", h.PostCount)

	group.POST("/call/:service", h.Call)
	group.POST("/message", h.Message)

	group.POST("/products", h.AddProduct)
	group.GET("/products/:name", h.GetProduct)
	group.POST("/products-send/:name", h.SendProduct)
}
