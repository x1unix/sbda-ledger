package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
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

	cfg.DB.SkipMigration = true
	conns, err := app.InstantiateConnectors(context.Background(), cfg)
	if err != nil {
		log.Fatalf("Failed to get test DB connectors: %s. Run 'docker-compose start' to start DB and Redis", err)
	}

	DB = conns.DB
	Redis = conns.Redis
	if err := TruncateData(); err != nil {
		conns.Close()
		log.Fatal(err)
	}

	exitCode := m.Run()
	if err := TruncateData(); err != nil {
		log.Println("TruncateData returned an error:", err)
	}
	conns.Close()
	os.Exit(exitCode)
}

func TruncateData() error {
	if err := Redis.FlushAll(context.Background()).Err(); err != nil {
		return fmt.Errorf("E2E - Redis.FlushAll failed: %w", err)
	}

	queries := []string{
		"TRUNCATE TABLE loans",
		"TRUNCATE TABLE group_membership",
		"TRUNCATE TABLE groups CASCADE",
		"TRUNCATE TABLE users CASCADE",
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("E2E - %q failed: %w", q, err)
		}
	}
	return nil
}

func TestPing(t *testing.T) {
	require.NoError(t, Client.Ping())
}

func shouldContainError(t *testing.T, err error, part string) {
	if err == nil {
		t.Fatalf("got no error, but expected %q", part)
		return
	}

	if msg := err.Error(); !strings.Contains(msg, part) {
		t.Fatalf("%q should contain %q", msg, part)
	}
}
