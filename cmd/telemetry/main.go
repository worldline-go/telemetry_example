package main

import (
	"context"
	"os"

	"github.com/worldline-go/initializer"
	"github.com/worldline-go/logz"

	"github.com/worldline-go/telemetry_example/cmd/telemetry/args"
	"github.com/worldline-go/telemetry_example/internal/config"
)

func main() {
	// set service name
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	initializer.Init(
		run,
		initializer.WithOptionsLogz(logz.WithCaller(false)),
		initializer.WithMsgf("%s [%s]", config.ServiceName, config.ServiceVersion),
	)
}

func run(ctx context.Context) error {
	return args.Execute(ctx)
}
