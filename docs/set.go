package docs

import (
	"path"

	"github.com/swaggo/swag"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/internal/config"
)

func SetVersion() {
	if spec, ok := swag.GetSwagger("swagger").(*swag.Spec); ok {
		spec.Title = config.AppName
		spec.Version = config.AppVersion
		spec.BasePath = path.Join("/", config.Application.BasePath, spec.BasePath)
	}
}
