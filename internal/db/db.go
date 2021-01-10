package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/config"
)

// Connect creates a database connection with appropriate pool configuration.
func Connect(ctx context.Context, cfg *config.Database) (*sqlx.DB, error) {
	poolCfg, err := cfg.PoolConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB connection string: %w", err)
	}

	conn, err := sqlx.ConnectContext(ctx, "pgx", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// stdlib package doesn't have a compat layer for pgxpool
	// so had to use standard sql api for pool configuration.
	conn.SetConnMaxIdleTime(poolCfg.MaxConnIdleTime)
	conn.SetConnMaxLifetime(poolCfg.MaxConnLifetime)
	conn.SetMaxOpenConns(int(poolCfg.MaxConns))

	if err = conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	mp := migrationParams{
		dbName:        poolCfg.ConnConfig.Database,
		versionTable:  cfg.VersionTable,
		migrationsDir: cfg.MigrationsDirectory,
		targetVersion: cfg.SchemaVersion,
	}
	if err = runMigration(conn, mp); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
