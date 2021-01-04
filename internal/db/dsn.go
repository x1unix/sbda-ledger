package db

import (
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DSN struct {
	poolCfg *pgxpool.Config
}

// UnmarshalText implements encoding.TextUnmarshaler
func (dsn *DSN) UnmarshalText(src []byte) (err error) {
	dsn.poolCfg, err = pgxpool.ParseConfig(string(src))
	if err != nil {
		return fmt.Errorf("failed to read database DSN connection string: %w", err)
	}

	return nil
}

// ConnConfig returns pgx connection params
func (dsn DSN) ConnConfig() *pgx.ConnConfig {
	return dsn.poolCfg.ConnConfig
}

// PoolConfig returns pgx pool configuration
func (dsn DSN) PoolConfig() *pgxpool.Config {
	return dsn.poolCfg
}
