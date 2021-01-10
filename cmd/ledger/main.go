package main

import (
	"flag"

	"github.com/x1unix/sbda-ledger/internal/app"
	"go.uber.org/zap"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "Path to config file (optional)")
	flag.Parse()

	cfg, err := app.ProvideConfig(cfgPath)
	if err != nil {
		app.Fatal("failed to read config:", err)
		return
	}

	logger, err := app.ProvideLogger(cfg)
	if err != nil {
		app.Fatal("failed to initialize logger:", err)
		return
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	ctx := app.ApplicationContext()
	conns, err := app.InstantiateConnectors(ctx, cfg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	defer conns.Close()
	svc := app.NewService(logger, conns, cfg)
	svc.Start(app.ApplicationContext())
}
