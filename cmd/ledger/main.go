package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/x1unix/sbda-ledger/internal/config"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "Path to config file")
	flag.Parse()

	cfg, err := provideConfig(cfgPath)
	if err != nil {
		fatal("failed to read config:", err)
		return
	}

	logger, err := provideLogger(cfg)
	if err != nil {
		fatal("failed to initialize logger:", err)
		return
	}

	h := web.Handler{}
	srv := web.NewServer(cfg.Server.ListenParams())
	srv.Router.HandleFunc("/test", h.Echo)

	logger.Info("Starting http server", zap.String("addr", cfg.Server.ListenAddress))
	if err = srv.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func provideLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Production {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}

func provideConfig(cfgPath string) (*config.Config, error) {
	if cfgPath == "" {
		return config.FromEnv()
	}

	return config.FromFile(cfgPath)
}

// fatal writes error to stderr and stops program with error exit code.
//
// used when no config or logger available and service is unable to initialize.
func fatal(vararg ...interface{}) {
	// print plain error to stderr, since logger not initialized yet
	// and default global zap logger is no-op logger
	_, _ = fmt.Fprintln(os.Stderr, vararg...)
	os.Exit(1)
}
