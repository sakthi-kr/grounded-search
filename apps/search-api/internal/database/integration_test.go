//go:build integration

package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
)

func TestOpenAgainstPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set for integration tests")
	}

	cfg := config.Config{
		DatabaseURL:      databaseURL,
		DatabaseTimeout:  5 * time.Second,
		DatabaseMaxConns: 4,
		DatabaseMinConns: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := Open(ctx, cfg)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	var value int
	if err := pool.Raw().QueryRow(ctx, "SELECT 1").Scan(&value); err != nil {
		t.Fatalf("SELECT 1: %v", err)
	}
	if value != 1 {
		t.Fatalf("SELECT 1 returned %d, want 1", value)
	}

	stats := pool.Stats()
	if stats.MaxConns != 4 {
		t.Errorf("MaxConns = %d, want 4", stats.MaxConns)
	}
	if stats.TotalConns < 1 {
		t.Errorf("TotalConns = %d, want at least 1", stats.TotalConns)
	}
}
