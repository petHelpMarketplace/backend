package app

import (
	"context"
	"fmt"
	"net/http"
	"pethelp-backend/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	conf       *config.Servers
	logger     *zap.Logger
	httpServer *http.Server
}

func NewHTTPServer(config *config.Servers, logger *zap.Logger) *Server {
	return &Server{
		conf:   config,
		logger: logger,
	}
}

func (s *Server) ListenAndServe(router *gin.Engine) error {
	s.httpServer = &http.Server{
		Addr:         s.conf.Web.Address,
		ReadTimeout:  s.conf.Web.ReadTimeout,
		WriteTimeout: s.conf.Web.WriteTimeout,
		IdleTimeout:  s.conf.Web.IdleTimeout,
		Handler:      router,
	}

	s.logger.Info("Starting HTTP server...")
	err := s.httpServer.ListenAndServe()
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		s.logger.Warn("Attempted to shut down HTTP server, but httpServer instance is nil. Was it ever started?")
		return nil
	}

	s.logger.Info("Attempting graceful shutdown of HTTP server...")
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		s.logger.Error("HTTP server graceful shutdown failed", zap.Error(err))
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}
	s.logger.Info("HTTP server gracefully shutdown.")
	return nil
}
