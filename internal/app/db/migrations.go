package db

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type migrationParams struct {
	dbName        string
	versionTable  string
	migrationsDir string
	targetVersion uint
}

func runMigration(conn *sqlx.DB, p migrationParams) error {
	d, err := postgres.WithInstance(conn.DB, &postgres.Config{
		MigrationsTable: p.versionTable,
		DatabaseName:    p.dbName,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+p.migrationsDir, p.dbName, d)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if p.targetVersion == 0 {
		err = m.Up()
	} else {
		err = m.Migrate(p.targetVersion)
	}

	if err == migrate.ErrNoChange {
		return nil
	}

	zap.L().Info("migration finished")
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
