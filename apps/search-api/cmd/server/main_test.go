package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

type fakeServer struct {
	listenErr   error
	shutdownErr error
	shutdownCh  chan struct{}
}

func (server *fakeServer) ListenAndServe() error {
	if server.shutdownCh == nil {
		return server.listenErr
	}

	<-server.shutdownCh
	return server.listenErr
}

func (server *fakeServer) Shutdown(context.Context) error {
	if server.shutdownErr != nil {
		return server.shutdownErr
	}

	if server.shutdownCh != nil {
		close(server.shutdownCh)
	}

	return nil
}

func TestRunReturnsListenError(t *testing.T) {
	t.Parallel()

	expected := errors.New("listen failed")
	server := &fakeServer{listenErr: expected}

	err := run(
		context.Background(),
		discardLogger(),
		server,
		time.Second,
	)

	if !errors.Is(err, expected) {
		t.Fatalf("run() error = %v, want wrapped %v", err, expected)
	}
}

func TestRunShutsDownAfterCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := &fakeServer{
		listenErr:  http.ErrServerClosed,
		shutdownCh: make(chan struct{}),
	}

	err := run(ctx, discardLogger(), server, time.Second)
	if err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func TestRunReturnsShutdownError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	expected := errors.New("shutdown failed")
	server := &fakeServer{
		shutdownErr: expected,
		shutdownCh:  make(chan struct{}),
	}

	err := run(ctx, discardLogger(), server, time.Second)
	if !errors.Is(err, expected) {
		t.Fatalf("run() error = %v, want wrapped %v", err, expected)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
