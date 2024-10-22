package config

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/igconfig"
	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
	"github.com/worldline-go/telemetry_example/internal/database/dbutil"
	"github.com/worldline-go/tell"
	"github.com/worldline-go/wkafka"
)

var (
	ServiceName    = "telemetry"
	ServiceVersion = "v0.0.0"
)

type Prefix struct {
	Vault  string `cfg:"vault"`
	Consul string `cfg:"consul"`
}

type OverrideHold struct {
	Memory *string
	Value  string
}

var Application = struct {
	LogLevel string `cfg:"log_level" default:"info"`
	Host     string `cfg:"host"      default:"0.0.0.0"`
	Port     string `cfg:"port"      default:"8080"`
	BasePath string `cfg:"base_path"`

	EnableKafkaConsumer bool `cfg:"enable_kafka_consumer"`
	EnableKafkaProducer bool `cfg:"enable_kafka_producer"`
	EnableDatabase      bool `cfg:"enable_database"`

	KafkaConfig wkafka.Config `cfg:"kafka_config"`
	// KafkaConsumer for consuming example
	KafkaConsumer wkafka.ConsumerConfig `cfg:"kafka_consumer"`
	// KafkaTopic for producing example
	KafkaTopic string `cfg:"kafka_topic"`

	// API for talk with http calls
	API map[string]klient.Config `cfg:"api"`

	Telemetry tell.Config

	Database Database `cfg:"database"`
}{}

type Database struct {
	DBDatasource string `cfg:"db_datasource" log:"false"`
	DBType       string `cfg:"db_type"       default:"pgx"`
	DBSchema     string `cfg:"db_schema"     default:"public"`

	Migrate Migrate `cfg:"migrate"`
}

type Migrate struct {
	DBDatasource string `cfg:"db_datasource" log:"false"`
	DBType       string `cfg:"db_type"       default:"pgx"`
	DBSchema     string `cfg:"db_schema"     default:"public"`
	DBTable      string `cfg:"db_table"      default:"migration"`
}

func Load(ctx context.Context) error {
	loaders := []loader.Loader{
		&loader.Default{},
		&loader.File{},
		&loader.Env{},
	}

	if err := igconfig.LoadWithLoadersWithContext(ctx, ServiceName, &Application, loaders...); err != nil {
		return fmt.Errorf("unable to load prefix settings: %w", err)
	}

	// set log again to get changes
	if err := logz.SetLogLevel(Application.LogLevel); err != nil {
		return err //nolint:wrapcheck // no need
	}

	if Application.EnableDatabase {
		dbDatasource, err := dbutil.SetDBSchema(Application.Database.DBDatasource, Application.Database.DBSchema)
		if err != nil {
			return fmt.Errorf("failed to set db schema: %w", err)
		}

		Application.Database.DBDatasource = dbDatasource

		if Application.Database.Migrate.DBDatasource == "" {
			Application.Database.Migrate.DBDatasource = Application.Database.DBDatasource
		}
	}

	// print loaded object
	log.Info().Object("config", igconfig.Printer{Value: Application}).Msg("loaded config")

	return nil
}
