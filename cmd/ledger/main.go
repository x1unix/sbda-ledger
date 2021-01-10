package main

import (
	"flag"

	"github.com/x1unix/sbda-ledger/internal/ledger"
	"go.uber.org/zap"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "Path to config file (optional)")
	flag.Parse()

	cfg, err := ledger.ProvideConfig(cfgPath)
	if err != nil {
		ledger.Fatal("failed to read config:", err)
		return
	}

	logger, err := ledger.ProvideLogger(cfg)
	if err != nil {
		ledger.Fatal("failed to initialize logger:", err)
		return
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()
	defer func() {
		logger.Info("Fuck!")
	}()

	app := ledger.NewService(logger, cfg)
	app.Start(ledger.ApplicationContext())
}
