package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
)

type MigrationParams struct {
	// TargetVersion is target database version (optional).
	// Leave empty to migrate to the latest version.
	TargetVersion int32

	// MigrationsDirectory is migrations source directory.
	MigrationsDirectory string

	// VersionTable is name of table which contains schema version.
	VersionTable string
}

// InstantiateConnection instantiates DB connection pool and prepares database
// by performing migrations from specified migrations directory.
func InstantiateConnection(ctx context.Context, cfg *pgxpool.Config, params MigrationParams) (*pgxpool.Pool, error) {
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to a database: %w", err)
	}

	if err = RunMigration(ctx, pool, params); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return pool, nil
}

// RunMigration performs schema migration using pool connection and provided params.
func RunMigration(ctx context.Context, pool *pgxpool.Pool, params MigrationParams) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), params.VersionTable)
	if err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}

	if err = migrator.LoadMigrations(params.MigrationsDirectory); err != nil {
		return fmt.Errorf("failed to load migrations from %q: %w", params.MigrationsDirectory, err)
	}

	if params.TargetVersion > 0 {
		return migrator.MigrateTo(ctx, params.TargetVersion)
	}

	return migrator.Migrate(ctx)
}
