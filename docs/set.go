package docs

import (
	"path"

	"github.com/swaggo/swag"

	"gitlab.global.ingenico.com/finops/sandbox/telemetry_example/internal/config"
)

func SetVersion() {
	if spec, ok := swag.GetSwagger("swagger").(*swag.Spec); ok {
		spec.Title = config.AppName
		spec.Version = config.AppVersion
		spec.BasePath = path.Join("/", config.Application.BasePath, spec.BasePath)
	}
}
