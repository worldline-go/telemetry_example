package args

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/worldline-go/initializer"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
	"github.com/worldline-go/tell"

	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/hold"
	"github.com/worldline-go/telemetry_example/internal/server"
	"github.com/worldline-go/telemetry_example/internal/server/handler"
	"github.com/worldline-go/telemetry_example/internal/telemetry"
)

var rootCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "telemetry example project",
	Long:  "example of trace, metrics, logs",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if err := logz.SetLogLevel(config.Application.LogLevel); err != nil {
			return err
		}

		return nil
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// load configuration
		if err := config.Load(cmd.Context()); err != nil {
			return err
		}

		if err := runRoot(cmd.Context()); err != nil {
			return err
		}

		return nil
	},
}

func runRoot(ctx context.Context) error {
	collector, err := tell.New(ctx, config.Application.Telemetry)
	if err != nil {
		return fmt.Errorf("failed to init telemetry; %w", err)
	}
	defer collector.Shutdown()

	if err := telemetry.SetGlobalMeter(); err != nil {
		return fmt.Errorf("failed to set metric; %w", err)
	}

	clients := make(map[string]*klient.Client)
	for name := range config.Application.API {
		client, err := config.Application.API[name].New()
		if err != nil {
			return fmt.Errorf("failed to create client [%s]; %w", name, err)
		}

		clients[name] = client
	}

	handlerServer := &handler.Handler{
		Counter: &hold.Counter{},
		Clients: clients,
	}

	// get router
	router := server.NewRouter(
		server.RouterSettings{
			Addr: net.JoinHostPort(config.Application.Host, config.Application.Port),
		},
		handlerServer,
	)

	initializer.ShutdownAdd(router.Stop, "http-server")

	// run server
	return router.Start()
}
