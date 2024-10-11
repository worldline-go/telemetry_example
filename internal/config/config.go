package config

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/worldline-go/igconfig"
	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/klient"
	"github.com/worldline-go/logz"
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

	KafkaConfig wkafka.Config `cfg:"kafka_config"`
	// KafkaConsumer for consuming example
	KafkaConsumer wkafka.ConsumerConfig `cfg:"kafka_consumer"`
	// KafkaTopic for producing example
	KafkaTopic string `cfg:"kafka_topic"`

	// API for talk with http calls
	API map[string]klient.Config `cfg:"api"`

	Telemetry tell.Config
}{}

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

	// print loaded object
	log.Info().Object("config", igconfig.Printer{Value: Application}).Msg("loaded config")

	return nil
}
