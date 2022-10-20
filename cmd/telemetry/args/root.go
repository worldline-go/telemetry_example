package args

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/worldline-go/logz"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/internal/config"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/internal/http"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/telemetry_example/pkg/cancel"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/instrumentation/metricecho"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
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

func runRoot(ctxParent context.Context) (err error) {
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	ctx, ctxCancel := context.WithCancel(ctxParent)
	defer ctxCancel()

	// cancel registry
	cReg := cancel.Registry{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sig:
			log.Warn().Msg("received shutdown signal...")
			ctxCancel()
		case <-ctx.Done():
			log.Warn().Msg("service closed")
		}

		cReg.Cancel()
	}()

	metricEchoViews, err := metricecho.GetViews()
	if err != nil {
		return fmt.Errorf("failed to get metricecho views; %w", err)
	}

	collector, err := tell.New(ctx, config.Application.Telemetry, metricEchoViews...)
	if err != nil {
		return fmt.Errorf("failed to init telemetry; %w", err)
	}

	cReg.Register(cancel.Function{
		Fn:   collector.Shutdown,
		Name: "collector",
	})

	if err := runtime.Start(); err != nil {
		return fmt.Errorf("failed to start runtime metrics; %w", err)
	}

	// run server
	router := http.NewRouter(http.RouterSettings{
		Host:      config.Application.Host + ":" + config.Application.Port,
		Telemetry: collector,
		// TracerProvider: tProvider,
		// MetricProvider: collector.MeterProvider,

		// PrometheusCollector: collector.MetricReaders.Prometheus
	})

	cReg.Register(cancel.Function{
		Fn:   router.Stop,
		Name: "http server",
	})

	if err := router.Start(); err != nil {
		return err
	}

	return nil
}
