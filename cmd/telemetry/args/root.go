package args

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"github.com/worldline-go/initializer"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
	"github.com/worldline-go/tell"
	"github.com/worldline-go/wkafka"
	"golang.org/x/sync/errgroup"

	"github.com/worldline-go/telemetry_example/internal/config"
	"github.com/worldline-go/telemetry_example/internal/database"
	"github.com/worldline-go/telemetry_example/internal/database/dbhandler"
	"github.com/worldline-go/telemetry_example/internal/hold"
	"github.com/worldline-go/telemetry_example/internal/kafka"
	"github.com/worldline-go/telemetry_example/internal/model"
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
	// //////////////////////////////////////////
	// telemetry initialization
	collector, err := tell.New(ctx, config.Application.Telemetry)
	if err != nil {
		return fmt.Errorf("failed to init telemetry; %w", err)
	}
	defer collector.Shutdown()

	if err := telemetry.SetGlobalMeter(); err != nil {
		return fmt.Errorf("failed to set metric; %w", err)
	}

	// //////////////////////////////////////////
	// database connection
	var db *sqlx.DB
	if config.Application.EnableDatabase {
		if err := database.MigrateDB(ctx, config.Application.Database.Migrate); err != nil {
			return fmt.Errorf("failed to migrate database; %w", err)
		}

		var err error
		db, err = database.Connect(ctx, config.Application.Database.DBDatasource, config.Application.Database.DBType)
		if err != nil {
			return fmt.Errorf("failed to connect to database; %w", err)
		}
		defer db.Close()
	}

	// db handler
	dbHandler := dbhandler.New(db)

	// //////////////////////////////////////////
	// http clients
	clients := make(map[string]*klient.Client)
	for name := range config.Application.API {
		client, err := config.Application.API[name].New(
			klient.WithHeaderAdd(http.Header{
				"Content-Type": []string{"application/json"},
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to create client [%s]; %w", name, err)
		}

		clients[name] = client
	}

	// //////////////////////////////////////////
	// kafka connection
	var kafkaClient *wkafka.Client
	var kafkaProducer *wkafka.Producer[*model.Product]
	var kafkaTracer *kotel.Tracer
	var kafkaOtel *kotel.Kotel
	if config.Application.EnableKafkaConsumer || config.Application.EnableKafkaProducer {
		kafkaTracer = kotel.NewTracer()
		kafkaOtel = kotel.NewKotel(kotel.WithTracer(kafkaTracer))
	}

	switch {
	case config.Application.EnableKafkaConsumer:
		kafkaClient, err = wkafka.New(ctx,
			config.Application.KafkaConfig,
			wkafka.WithConsumer(config.Application.KafkaConsumer),
			wkafka.WithClientInfo(config.ServiceName, config.ServiceVersion),
			wkafka.WithKGOOptions(kgo.WithHooks(kafkaOtel.Hooks()...)),
		)
		if err != nil {
			return fmt.Errorf("failed to create kafka client; %w", err)
		}
	case config.Application.EnableKafkaProducer:
		kafkaClient, err = wkafka.New(ctx,
			config.Application.KafkaConfig,
			wkafka.WithClientInfo(config.ServiceName, config.ServiceVersion),
			wkafka.WithKGOOptions(kgo.WithHooks(kafkaOtel.Hooks()...)),
		)
		if err != nil {
			return fmt.Errorf("failed to create kafka client; %w", err)
		}

		kafkaProducer, err = wkafka.NewProducer[*model.Product](kafkaClient, config.Application.KafkaTopic)
		if err != nil {
			return fmt.Errorf("failed to create kafka producer; %w", err)
		}
	}
	if kafkaClient != nil {
		defer kafkaClient.Close()
	}

	// //////////////////////////////////////////
	// set handlers
	handlerServer := &handler.Handler{
		Counter:       &hold.Counter{},
		Clients:       clients,
		KafkaProducer: kafkaProducer,
		KafkaTracer:   kafkaTracer,
		DB:            dbHandler,
	}

	handlerKafka := kafka.Kafka{
		DB:     dbHandler,
		Tracer: kafkaTracer,
	}

	// //////////////////////////////////////////
	// set router
	router := server.NewRouter(
		server.RouterSettings{
			Addr: net.JoinHostPort(config.Application.Host, config.Application.Port),
		},
		handlerServer,
	)

	// //////////////////////////////////////////
	// run listeners

	if !config.Application.EnableKafkaConsumer {
		// just run server
		router.StopWithContext(ctx, initializer.WaitGroup(ctx))
		return router.Start()
	}

	// run multiple listeners
	g, ctx := errgroup.WithContext(ctx)

	// run kafka consumer
	g.Go(func() error {
		return kafkaClient.Consume(ctx, wkafka.WithCallback(handlerKafka.Consume))
	})

	// run http server
	g.Go(func() error {
		router.StopWithContext(ctx, initializer.WaitGroup(ctx))
		return router.Start()
	})

	return g.Wait()
}
