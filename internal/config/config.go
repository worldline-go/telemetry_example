package config

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
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

var LoadConfig = struct {
	Prefix  Prefix `cfg:"prefix"`
	AppName string `cfg:"app_name"`
}{
	AppName: ServiceName,
}

var Application = struct {
	LogLevel string `cfg:"log_level" default:"info"`
	Host     string `cfg:"host"      default:"0.0.0.0"`
	Port     string `cfg:"port"      default:"8080"`
	BasePath string `cfg:"base_path"`

	KafkaConfig wkafka.Config `cfg:"kafka_config"`
	// KafkaConsumer for consuming example
	KafkaConsumer wkafka.ConsumerConfig `cfg:"kafka_consumer"`
	// KafkaTopic for producing example
	KafkaTopic string `cfg:"kafka_topic"`

	// API for talk with http calls
	API map[string]klient.Config `cfg:"api"`

	Telemetry tell.Config
}{}

func Load(ctx context.Context, visit func(fn func(*pflag.Flag)), overrideValues map[string]OverrideHold) error {
	loaders := []loader.Loader{
		&loader.Default{},
		&loader.File{},
		&loader.Env{},
	}

	if err := igconfig.LoadWithLoadersWithContext(ctx, ServiceName, &LoadConfig, loaders...); err != nil {
		return fmt.Errorf("unable to load prefix settings: %v", err)
	}

	// override used cmd values
	visit(func(f *pflag.Flag) {
		if v, ok := overrideValues[f.Name]; ok {
			*v.Memory = v.Value
		}
	})

	// set log again to get changes
	if err := logz.SetLogLevel(Application.LogLevel); err != nil {
		return err //nolint:wrapcheck // no need
	}

	// print loaded object
	log.Info().Object("config", igconfig.Printer{Value: Application}).Msg("loaded config")

	return nil
}
