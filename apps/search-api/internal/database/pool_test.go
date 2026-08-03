package database

import (
	"testing"
	"time"

	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
)

func TestBuildPoolConfigAppliesDatabaseSettings(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		DatabaseURL:      "postgres://user:password@localhost:5432/example?sslmode=disable",
		DatabaseTimeout:  4 * time.Second,
		DatabaseMaxConns: 12,
		DatabaseMinConns: 3,
	}

	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		t.Fatalf("buildPoolConfig() error = %v", err)
	}

	if poolConfig.MaxConns != 12 {
		t.Errorf("MaxConns = %d, want 12", poolConfig.MaxConns)
	}
	if poolConfig.MinConns != 3 {
		t.Errorf("MinConns = %d, want 3", poolConfig.MinConns)
	}
	if poolConfig.MaxConnLifetime != defaultMaxConnLifetime {
		t.Errorf(
			"MaxConnLifetime = %v, want %v",
			poolConfig.MaxConnLifetime,
			defaultMaxConnLifetime,
		)
	}
	if poolConfig.MaxConnIdleTime != defaultMaxConnIdleTime {
		t.Errorf(
			"MaxConnIdleTime = %v, want %v",
			poolConfig.MaxConnIdleTime,
			defaultMaxConnIdleTime,
		)
	}
	if poolConfig.HealthCheckPeriod != defaultHealthCheckPeriod {
		t.Errorf(
			"HealthCheckPeriod = %v, want %v",
			poolConfig.HealthCheckPeriod,
			defaultHealthCheckPeriod,
		)
	}

	connectionConfig := poolConfig.ConnConfig
	if connectionConfig.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", connectionConfig.Host)
	}
	if connectionConfig.Port != 5432 {
		t.Errorf("Port = %d, want 5432", connectionConfig.Port)
	}
	if connectionConfig.Database != "example" {
		t.Errorf(
			"Database = %q, want example",
			connectionConfig.Database,
		)
	}
	if connectionConfig.User != "user" {
		t.Errorf("User = %q, want user", connectionConfig.User)
	}
}

func TestBuildPoolConfigRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		DatabaseURL:      "://invalid",
		DatabaseTimeout:  time.Second,
		DatabaseMaxConns: 10,
		DatabaseMinConns: 1,
	}

	if _, err := buildPoolConfig(cfg); err == nil {
		t.Fatal("buildPoolConfig() error = nil, want error")
	}
}

func TestNilPoolMethodsAreSafe(t *testing.T) {
	t.Parallel()

	var pool *Pool

	if err := pool.Ping(t.Context()); err == nil {
		t.Fatal("Ping() error = nil, want error")
	}

	pool.Close()

	if stats := pool.Stats(); stats != (Stats{}) {
		t.Errorf("Stats() = %+v, want zero value", stats)
	}

	if raw := pool.Raw(); raw != nil {
		t.Errorf("Raw() = %v, want nil", raw)
	}
}
