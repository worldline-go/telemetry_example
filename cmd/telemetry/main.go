package main

import (
	"context"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/logz"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/cmd/telemetry/args"
)

func main() {
	logz.InitializeLog(nil)

	if err := args.Execute(context.Background()); err != nil {
		if !errors.Is(err, args.ErrShutdown) {
			log.Error().Err(err).Msg("failed to execute command")
		}

		os.Exit(1)
	}
}
