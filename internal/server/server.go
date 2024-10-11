package server

import (
	"context"
	"errors"
	"net/http"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/worldline-go/logz/logecho"
	"github.com/worldline-go/tell/metric/metricecho"
	"github.com/ziflex/lecho/v3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"

	"github.com/worldline-go/telemetry_example/docs"
	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/server/handler"
)

var shutdownTimeout = 5 * time.Second

type RouterSettings struct {
	// Addr like 0.0.0.0:8080
	Addr string
	// BasePath for serving
	BasePath string
	// ShutdownTimeout using in shutdown server default is 5 second.
	ShutdownTimeout time.Duration
}

func (rs RouterSettings) SetDefaults() RouterSettings {
	if rs.Addr == "" {
		rs.Addr = "0.0.0.0:8080"
	}

	if rs.ShutdownTimeout == 0 {
		rs.ShutdownTimeout = time.Second * 5
	}

	return rs
}

type Router struct {
	echo *echo.Echo
	rs   RouterSettings
}

func NewRouter(rs RouterSettings, handler *handler.Handler) *Router {
	e := echo.New()
	e.HideBanner = true

	e.Logger = lecho.From(log.Logger)
	// e.Validator = util.NewValidator()

	e.Use(
		middleware.Recover(),
		middleware.CORS(),
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(logecho.RequestLoggerConfig()),
		logecho.ZerologLogger(),
	)

	// add echo metrics
	e.Use(metricecho.HTTPMetrics(nil))

	// add otel tracing
	e.Use(otelecho.Middleware(config.ServiceName, otelecho.WithTracerProvider(otel.GetTracerProvider())))

	if err := docs.Info(path.Join("/api/v1/", rs.BasePath)); err != nil {
		log.Warn().Err(err).Msg("failed to set swagger info")
	}

	router := &Router{
		echo: e,
		rs:   rs.SetDefaults(),
	}

	router.Register(rs.BasePath, nil, handler)

	return router
}

// Register
//
// @title Telemetry
// @version 1.0
// @description telemetry example project
//
// @contact.name @FINOPS
// @contact.email @FINOPS
//
// @host
// @BasePath /api/v1
func (r *Router) Register(basePath string, middlewares []echo.MiddlewareFunc, h *handler.Handler) {
	var z interface {
		GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
		Group(prefix string, m ...echo.MiddlewareFunc) *echo.Group
	}

	z = r.echo
	basement := ""

	if basePath != "" {
		basement = path.Join("/", basePath)
		z = r.echo.Group(basement)
	}

	z.GET("/api/swagger/*", echoSwagger.WrapHandler)
	v1 := z.Group("/api/v1")

	h.Register(v1, middlewares)
}

func (r *Router) Start() error {
	if err := r.echo.Start(r.rs.Addr); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (r *Router) Stop() error {
	if r.echo == nil {
		log.Info().Msg("server not running")

		return nil
	}

	log.Info().Msg("stopping service...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := r.echo.Shutdown(ctxShutdown); err != nil {
		return err
	}

	return nil
}
