package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/pkg/logger"
)

type Server struct {
	httpServer *http.Server
	logger     *logger.Logger
}

func NewServer(cfg *config.Config, handler http.Handler, appLogger *logger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              net.JoinHostPort(cfg.Http.Host, cfg.Http.Port),
			Handler:           handler,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: appLogger,
	}
}

func (s *Server) Run() {
	go func() {
		if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.WithComponent("server").WithError(err).Error("Error occurred while running http server")
		}
	}()
}

func (s *Server) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.WithComponent("server").WithError(err).Error("Stopping server failed")
	}
}
