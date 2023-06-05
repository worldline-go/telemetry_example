package main

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/logz"

	"gitlab.global.ingenico.com/finops/sandbox/telemetry_example/cmd/telemetry/args"
)

func main() {
	logz.InitializeLog(logz.WithCaller(false))

	if err := args.Execute(context.Background()); err != nil {
		if !errors.Is(err, args.ErrShutdown) {
			log.Error().Err(err).Msg("failed to execute command")
		}

		os.Exit(1)
	}
}
