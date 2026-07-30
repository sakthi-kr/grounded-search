package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
)

func TestNewAppliesConfiguration(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Host:              "127.0.0.1",
		Port:              9090,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      4 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	handler := http.HandlerFunc(func(
		http.ResponseWriter,
		*http.Request,
	) {
	})

	server := New(cfg, handler)

	if server.Address() != "127.0.0.1:9090" {
		t.Errorf(
			"Address() = %q, want 127.0.0.1:9090",
			server.Address(),
		)
	}
	if server.httpServer.Handler == nil {
		t.Fatal("Handler is nil")
	}
	if server.httpServer.ReadHeaderTimeout != 2*time.Second {
		t.Errorf(
			"ReadHeaderTimeout = %v, want 2s",
			server.httpServer.ReadHeaderTimeout,
		)
	}
	if server.httpServer.ReadTimeout != 3*time.Second {
		t.Errorf(
			"ReadTimeout = %v, want 3s",
			server.httpServer.ReadTimeout,
		)
	}
	if server.httpServer.WriteTimeout != 4*time.Second {
		t.Errorf(
			"WriteTimeout = %v, want 4s",
			server.httpServer.WriteTimeout,
		)
	}
	if server.httpServer.IdleTimeout != 5*time.Second {
		t.Errorf(
			"IdleTimeout = %v, want 5s",
			server.httpServer.IdleTimeout,
		)
	}
}
