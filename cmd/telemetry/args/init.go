package args

import (
	"context"
)

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx) //nolint:wrapcheck // no need
}
