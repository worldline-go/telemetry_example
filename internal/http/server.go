package http

import (
	"context"
	"errors"
	"net/http"
	"path"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swag "github.com/worldline-go/echo-swagger"
	"github.com/worldline-go/tell/metric/metricecho"
	"github.com/ziflex/lecho/v3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"

	"github.com/worldline-go/telemetry_example/docs"
	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/hold"
	"github.com/worldline-go/telemetry_example/internal/http/handler"
	"github.com/worldline-go/telemetry_example/internal/http/middle"
)

var shutdownTimeout = 5 * time.Second

type RouterSettings struct {
	// Host like 0.0.0.0:8080
	Host string
	// BasePath for serving
	BasePath string
	// ShutdownTimeout using in shutdown server default is 5 second.
	ShutdownTimeout time.Duration
}

func (rs RouterSettings) SetDefaults() RouterSettings {
	if rs.Host == "" {
		rs.Host = "0.0.0.0:8080"
	}

	if rs.ShutdownTimeout == 0 {
		rs.ShutdownTimeout = time.Second * 5
	}

	return rs
}

type Router struct {
	echo    *echo.Echo
	rs      RouterSettings
	counter hold.Counter
}

func NewRouter(rs RouterSettings) *Router {
	e := echo.New()
	e.HideBanner = true

	e.Logger = lecho.From(log.Logger)
	// e.Validator = util.NewValidator()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middle.Logger(log.Logger, zerolog.InfoLevel))

	// add echo metrics
	e.Use(metricecho.HTTPMetrics(nil))

	// add otel tracing
	e.Use(otelecho.Middleware(config.LoadConfig.AppName, otelecho.WithTracerProvider(otel.GetTracerProvider())))

	docs.SetVersion()

	router := &Router{
		echo: e,
		rs:   rs.SetDefaults(),
	}

	router.Register(rs.BasePath, nil)

	return router
}

// Register
//
// @title Telemetry
// @version 1.0
// @description telemetry example project
//
// @contact.name DeepCore Team
// @contact.email @FINOPS @DEEPCORE
//
// @host
// @BasePath /api/v1
func (r *Router) Register(basePath string, middlewares []echo.MiddlewareFunc) {
	z := r.echo.Group("")
	basement := ""

	if basePath != "" {
		basement = path.Join("/", basePath)
		z = r.echo.Group(basement)
	}

	// if r.rs.PrometheusCollector != nil {
	// log.Info().Msgf("metrics on %s", path.Join(basement, "/metrics"))

	// registry := prometheus.NewRegistry()
	// registry.MustRegister(r.rs.PrometheusCollector)

	// z.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})))
	// }

	v1 := z.Group("/api/v1")
	h := handler.Handlers{
		Counter: &r.counter,
	}

	h.Register(v1, middlewares)

	v1.GET("/swagger/*", swag.WrapHandler)
}

func (r *Router) Start() error {
	if err := r.echo.Start(r.rs.Host); !errors.Is(err, http.ErrServerClosed) {
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
