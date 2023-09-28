package args

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/worldline-go/initializer"
	"github.com/worldline-go/logz"
	"github.com/worldline-go/tell"

	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/http"
	"github.com/worldline-go/telemetry_example/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
)

var rootCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "telemetry example project",
	Long:  "example of trace, metrics, logs",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := logz.SetLogLevel(config.Application.LogLevel); err != nil {
			return err
		}

		return nil
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// appname and version
		log.WithLevel(zerolog.NoLevel).Msgf("%s [%s]", strings.ToTitle(config.AppName), config.AppVersion)

		ow := make(map[string]config.OverrideHold)
		ow["host"] = config.OverrideHold{Memory: &config.Application.Host, Value: config.Application.Host}
		ow["port"] = config.OverrideHold{Memory: &config.Application.Port, Value: config.Application.Port}
		ow["log-level"] = config.OverrideHold{Memory: &config.Application.LogLevel, Value: config.Application.LogLevel}
		// load configuration
		if err := config.Load(cmd.Context(), cmd.Flags().Visit, ow); err != nil {
			return err
		}

		if err := runRoot(cmd.Context()); err != nil {
			return err
		}

		return nil
	},
}

func RootCmdFlags() {
	rootCmd.Flags().StringVarP(&config.Application.Host, "host", "H", config.Application.Host, "Host to listen on")
	rootCmd.Flags().StringVarP(&config.Application.Port, "port", "P", config.Application.Port, "Port to listen on")
	rootCmd.PersistentFlags().StringVarP(&config.Application.LogLevel, "log-level", "l", config.Application.LogLevel, "Log level")
}

func runRoot(ctx context.Context) (err error) {
	// get router
	router := http.NewRouter(http.RouterSettings{
		Host: config.Application.Host + ":" + config.Application.Port,
	})
	initializer.Shutdown.Add(router.Stop)

	collector, err := tell.New(ctx, config.Application.Telemetry)
	if err != nil {
		return fmt.Errorf("failed to init telemetry; %w", err)
	}
	defer collector.Shutdown()

	telemetry.AddGlobalAttr(attribute.Key("special").String("X"))
	if err := telemetry.SetGlobalMeter(); err != nil {
		return fmt.Errorf("failed to set metric; %w", err)
	}

	// if err := runtime.Start(); err != nil {
	// 	return fmt.Errorf("failed to start runtime metrics; %w", err)
	// }

	// run server
	return router.Start()
}
