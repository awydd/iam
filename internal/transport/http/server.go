package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/logger"
	"github.com/awydd/iam/internal/transport/http/router"
	"github.com/awydd/iam/internal/wire"
	"golang.org/x/net/netutil"
)

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	shutdownTO time.Duration
}

func New(cfg *conf.HTTP, app *wire.App) (*Server, error) {
	engine := router.New(cfg, app)

	addr := ":" + cfg.Port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("http: listen %s: %w", addr, err)
	}

	if cfg.MaxConnections > 0 {
		ln = netutil.LimitListener(ln, cfg.MaxConnections)
	}

	srv := &http.Server{
		Handler:           engine,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &Server{
		httpServer: srv,
		listener:   ln,
		shutdownTO: cfg.ShutdownTimeout,
	}, nil
}

func (s *Server) Start() error {
	logger.Info("http server listening on %s", s.listener.Addr().String())
	if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http: serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTO)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http: shutdown: %w", err)
	}
	return nil
}
