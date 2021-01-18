package e2e

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/x1unix/sbda-ledger/internal/app"
	"github.com/x1unix/sbda-ledger/pkg/ledger"
)

var (
	Client *ledger.Client
	DB     *sqlx.DB
	Redis  *redis.Client
)

func formatClientUrl(addr string) string {
	if addr[0] == ':' {
		addr = "localhost" + addr
	}

	return "http://" + addr
}

func TestMain(m *testing.M) {
	cfg, err := app.ProvideConfig("../config.dev.yaml")
	if err != nil {
		log.Fatal("Failed to read dev config:", err)
	}

	Client = ledger.NewClient(&http.Client{}, formatClientUrl(cfg.Server.ListenAddress))
	if err := Client.Ping(); err != nil {
		log.Fatalf("Failed to ping test Ledger API: %s. Run 'make run' to start test API", err)
	}

	conns, err := app.InstantiateConnectors(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to get test DB connectors: %s. Run 'docker-compose start' to start DB and Redis", err)
	}

	DB = conns.DB
	Redis = conns.Redis
	exitCode := m.Run()
	conns.Close()
	os.Exit(exitCode)
}
