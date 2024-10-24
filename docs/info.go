package docs

import (
	_ "github.com/swaggo/swag"
	"github.com/worldline-go/swagger"

	"github.com/worldline-go/telemetry_example/internal/config"
)

func Info(basePath string) error {
	return swagger.SetInfo(
		swagger.WithBasePath(basePath),
		swagger.WithTitle(config.ServiceName),
		swagger.WithVersion(config.ServiceVersion),
	)
}
