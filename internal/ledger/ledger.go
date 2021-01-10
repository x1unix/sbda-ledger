package ledger

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/x1unix/sbda-ledger/internal/config"
	"go.uber.org/zap"
)

var once = &sync.Once{}

var (
	ctx        context.Context
	cancelFunc context.CancelFunc
)

// ApplicationContext returns global application context for graceful shutdown
func ApplicationContext() context.Context {
	once.Do(func() {
		ctx, cancelFunc = context.WithCancel(context.Background())

		go func() {
			signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT}
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, signals...)
			defer signal.Reset(signals...)
			<-sigChan
			cancelFunc()
		}()
	})

	return ctx
}

// ProvideLogger returns logger according to environment
func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Production {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}

// ProvideConfig reads and provides config.
//
// If config path is empty, config is loaded from environment variables and defaults.
func ProvideConfig(cfgPath string) (*config.Config, error) {
	if cfgPath == "" {
		return config.FromEnv()
	}

	return config.FromFile(cfgPath)
}

// Fatal writes error to stderr and stops program with error exit code.
//
// used when no config or logger available and service is unable to initialize.
func Fatal(vararg ...interface{}) {
	// print plain error to stderr, since logger not initialized yet
	// and default global zap logger is no-op logger
	_, _ = fmt.Fprintln(os.Stderr, vararg...)
	os.Exit(1)
}
