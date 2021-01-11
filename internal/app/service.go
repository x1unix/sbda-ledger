package app

import (
	"context"
	"net/http"
	"sync"

	"github.com/x1unix/sbda-ledger/internal/config"
	"github.com/x1unix/sbda-ledger/internal/handler"
	"github.com/x1unix/sbda-ledger/internal/repository"
	"github.com/x1unix/sbda-ledger/internal/service"
	"github.com/x1unix/sbda-ledger/internal/web"
	"github.com/x1unix/sbda-ledger/internal/web/middleware"
	"go.uber.org/zap"
)

type Service struct {
	server *web.Server
	logger *zap.Logger
}

func NewService(logger *zap.Logger, conn *Connectors, cfg *config.Config) *Service {
	srv := web.NewServer(cfg.Server.ListenParams())

	userSvc := service.NewUsersService(logger, repository.NewUserRepository(conn.DB))
	authSvc := service.NewAuthService(logger, userSvc, repository.NewSessionRepository(conn.Redis))

	requireAuth := middleware.NewAuthMiddleware(authSvc)
	authHandler := handler.NewAuthHandler(userSvc, authSvc)

	hWrapper := web.NewWrapper(logger.Named("http"))
	srv.Router.Methods(http.MethodPost).
		Path("/auth").
		HandlerFunc(hWrapper.WrapResourceHandler(authHandler.Login))
	srv.Router.Methods(http.MethodPost).
		Path("/auth/register").
		HandlerFunc(hWrapper.WrapResourceHandler(authHandler.Register))
	srv.Router.Methods(http.MethodGet).
		Path("/auth/session").
		HandlerFunc(hWrapper.WrapResourceHandler(authHandler.GetSession, requireAuth))
	srv.Router.Methods(http.MethodDelete).
		Path("/auth/session").
		HandlerFunc(hWrapper.WrapHandler(authHandler.Logout, requireAuth))

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
