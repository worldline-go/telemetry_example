package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/worldline-go/klient"

	"github.com/worldline-go/telemetry_example/internal/hold"
)

type Handler struct {
	Counter *hold.Counter
	Clients map[string]*klient.Client
}

func (h *Handler) Register(group *echo.Group, middlewares []echo.MiddlewareFunc) {
	group.GET("/count", h.GetCount, middlewares...)
	group.POST("/count", h.PostCount, middlewares...)

	group.POST("/call/:service", h.Call, middlewares...)
	group.POST("/message", h.Message, middlewares...)
}
