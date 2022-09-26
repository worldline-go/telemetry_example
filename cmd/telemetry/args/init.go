package args

import "context"

func InitializeFlags() {
	RootCmdFlags()
}

func Execute(ctx context.Context) error {
	InitializeFlags()

	return rootCmd.ExecuteContext(ctx) //nolint:wrapcheck // no need
}
