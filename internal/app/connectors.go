package app

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/app/db"
	"github.com/x1unix/sbda-ledger/internal/config"
)

// Connectors contains set of I/O connectors for ledger service.
type Connectors struct {
	// DB is database connection
	DB *sqlx.DB

	// Redis is redis connection
	Redis *redis.Client
}

// Close closes all connections
func (c Connectors) Close() {
	_ = c.DB.Close()
	_ = c.Redis.Close()
}

// InstantiateConnectors establishes connections to database, cache, etc.
// and returns set of connectors for further application initialization.
func InstantiateConnectors(ctx context.Context, cfg *config.Config) (*Connectors, error) {
	dbConn, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		return nil, err
	}

	redisConn := redis.NewClient(cfg.Redis.RedisOptions())
	if _, err = redisConn.Ping(ctx).Result(); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("failed to connect to Redis server: %w", err)
	}

	return &Connectors{
		DB:    dbConn,
		Redis: redisConn,
	}, nil
}
