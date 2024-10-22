package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/worldline-go/igmigrator"

	"github.com/worldline-go/telemetry_example/internal/config"
)

func MigrateDB(ctx context.Context, migrate config.Migrate) error {
	if migrate.DBDatasource == "" {
		return fmt.Errorf("migrate database datasource is empty")
	}

	db, err := sqlx.Connect(migrate.DBType, migrate.DBDatasource)
	if err != nil {
		return fmt.Errorf("migrate database connect: %w", err)
	}

	defer db.Close()

	prevVersion, newVersion, err := igmigrator.Migrate(ctx, db, &igmigrator.Config{
		MigrationsDir:  "migrations",
		Schema:         migrate.DBSchema,
		MigrationTable: migrate.DBTable,
	})
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	if newVersion != prevVersion {
		log.Info().Msgf("ran migrations from version %d to %d", prevVersion, newVersion)
	}

	return nil
}
