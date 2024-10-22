package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ConnMaxLifetime = 15 * time.Minute
	MaxIdleConns    = 3
	MaxOpenConns    = 3
)

// Connect attempts to connect to database server.
func Connect(ctx context.Context, dbDatasource, dbType string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, dbType, dbDatasource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetConnMaxLifetime(ConnMaxLifetime)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetMaxOpenConns(MaxOpenConns)

	return db, nil
}
