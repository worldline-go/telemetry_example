package middle

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func Logger(logger zerolog.Logger, level zerolog.Level) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			// Do not log */ping endpoints as it could be called every 10 seconds or more.
			// In future instead of checking suffix this could check route name(?).
			return strings.HasSuffix(path, "/ping")
		},
		LogRequestID:     true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.WithLevel(level).
				Str("id", v.RequestID).
				Str("remote_ip", v.RemoteIP).
				Str("host", v.Host).
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("user_agent", v.UserAgent).
				Int("status", v.Status).
				Err(v.Error).
				Int64("latency", v.Latency.Nanoseconds()).
				Str("latency_human", v.Latency.String()).
				Str("bytes_in", v.ContentLength).
				Int64("bytes_out", v.ResponseSize).
				Msg("request")

			return nil
		},
	})
}
