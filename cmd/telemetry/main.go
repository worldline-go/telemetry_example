package main

import (
	"context"
	"sync"

	"github.com/worldline-go/initializer"
	"github.com/worldline-go/logz"

	"github.com/worldline-go/telemetry_example/cmd/telemetry/args"
)

func main() {
	initializer.Init(
		run,
		initializer.WithOptionsLogz(logz.WithCaller(false)),
	)
}

func run(ctx context.Context, _ *sync.WaitGroup) error {
	return args.Execute(ctx)
}
