package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/httpapi"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/mlclient"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/server"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: cfg.LogLevel},
	))

	modelInfoClient, err := mlclient.New(
		cfg.MLServiceURL,
		cfg.MLServiceTimeout,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ML client configuration error: %v\n", err)
		os.Exit(1)
	}

	handler := httpapi.NewHandler(
		logger,
		version,
		time.Now().UTC(),
		modelInfoClient,
	)
	httpServer := server.New(cfg, handler)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Info(
		"search API starting",
		"address", httpServer.Address(),
		"version", version,
		"ml_service_url", cfg.MLServiceURL,
	)

	if err := run(ctx, logger, httpServer, cfg.ShutdownTimeout); err != nil {
		logger.Error("search API stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("search API stopped")
}

type runnableServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func run(
	ctx context.Context,
	logger *slog.Logger,
	httpServer runnableServer,
	shutdownTimeout time.Duration,
) error {
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and serve: %w", err)

	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	select {
	case err := <-serverErrors:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server exit after shutdown: %w", err)

	case <-shutdownCtx.Done():
		return fmt.Errorf(
			"server did not stop before shutdown deadline: %w",
			shutdownCtx.Err(),
		)
	}
}
