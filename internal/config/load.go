package config

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/worldline-go/igconfig"
	"github.com/worldline-go/igconfig/loader"
	"github.com/worldline-go/logz"
)

type OverrideHold struct {
	Memory *string
	Value  string
}

func Load(ctx context.Context, visit func(fn func(*pflag.Flag)), overrideValues map[string]OverrideHold) error {
	logConfig := log.With().Str("component", "config").Logger()
	ctxConfig := logConfig.WithContext(ctx)

	loaders := []loader.Loader{
		&loader.Default{},
		&loader.Consul{},
		&loader.Vault{},
		&loader.File{},
		&loader.Env{},
	}

	loader.VaultSecretAdditionalPaths = append(loader.VaultSecretAdditionalPaths,
		loader.AdditionalPath{Map: "migrate", Name: "migrations"},
	)

	if err := igconfig.LoadWithLoadersWithContext(ctxConfig, "", &LoadConfig, loaders[3]); err != nil {
		return fmt.Errorf("unable to load prefix settings: %v", err)
	}

	loader.ConsulConfigPathPrefix = LoadConfig.Prefix.Consul
	loader.VaultSecretBasePath = LoadConfig.Prefix.Vault

	if err := igconfig.LoadWithLoadersWithContext(ctxConfig, LoadConfig.AppName, &Application, loaders...); err != nil {
		return fmt.Errorf("unable to load configuration settings: %v", err)
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
