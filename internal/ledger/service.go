package ledger

import (
	"context"
	"net/http"
	"sync"

	"github.com/x1unix/sbda-ledger/internal/config"
	"github.com/x1unix/sbda-ledger/internal/handler"
	"github.com/x1unix/sbda-ledger/internal/web"
	"go.uber.org/zap"
)

type Service struct {
	server *web.Server
	logger *zap.Logger
}

func NewService(logger *zap.Logger, cfg *config.Config) *Service {
	hWrapper := web.NewWrapper(logger.Named("http"))
	h := handler.AuthHandler{}
	srv := web.NewServer(cfg.Server.ListenParams())

	srv.Router.Methods(http.MethodPost).
		Path("/auth/login").
		HandlerFunc(hWrapper.WrapResourceHandler(h.Auth))

	return &Service{
		server: srv,
		logger: logger,
	}
}

// Start starts the service
func (s Service) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Info("starting http service", zap.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		if err := s.server.Shutdown(ctx); err != nil {
			if err == context.Canceled {
				return
			}
			s.logger.Error("failed to shutdown server", zap.Error(err))
		}
	}()

	wg.Wait()
	s.logger.Info("goodbye")
}
