package args

import (
	"context"
	"errors"
)

var ErrShutdown = errors.New("shutting down signal received")

func InitializeFlags() {
	RootCmdFlags()
}

func Execute(ctx context.Context) error {
	InitializeFlags()

	return rootCmd.ExecuteContext(ctx) //nolint:wrapcheck // no need
}
