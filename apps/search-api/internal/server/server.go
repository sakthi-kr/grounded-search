package server

import (
	"context"
	"net/http"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
)

// Server wraps the standard-library HTTP server.
type Server struct {
	httpServer *http.Server
}

// New creates a configured HTTP server.
func New(
	cfg config.Config,
	handler http.Handler,
) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Address(),
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}
}

// Address returns the configured listen address.
func (server *Server) Address() string {
	return server.httpServer.Addr
}

// ListenAndServe starts the HTTP server.
func (server *Server) ListenAndServe() error {
	return server.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (server *Server) Shutdown(ctx context.Context) error {
	return server.httpServer.Shutdown(ctx)
}
