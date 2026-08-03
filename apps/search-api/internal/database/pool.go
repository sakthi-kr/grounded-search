package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
)

const (
	defaultMaxConnLifetime   = 30 * time.Minute
	defaultMaxConnIdleTime   = 5 * time.Minute
	defaultHealthCheckPeriod = 30 * time.Second
)

// Pool wraps pgxpool.Pool and owns database connectivity for the search API.
type Pool struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// Stats is a small, stable view of connection-pool statistics.
type Stats struct {
	AcquiredConns int32
	IdleConns     int32
	TotalConns    int32
	MaxConns      int32
}

// Open parses the configured connection string, creates the pool, and verifies
// that PostgreSQL is reachable before returning.
func Open(ctx context.Context, cfg config.Config) (*Pool, error) {
	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, cfg.DatabaseTimeout)
	defer cancel()

	pgxPool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	pool := &Pool{
		pool:         pgxPool,
		queryTimeout: cfg.DatabaseTimeout,
	}

	if err := pool.Ping(ctx); err != nil {
		pgxPool.Close()
		return nil, err
	}

	return pool, nil
}

func buildPoolConfig(cfg config.Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DatabaseMaxConns)
	poolConfig.MinConns = int32(cfg.DatabaseMinConns)
	poolConfig.MaxConnLifetime = defaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = defaultHealthCheckPeriod

	return poolConfig, nil
}

// Ping verifies that PostgreSQL can answer within the configured timeout.
func (pool *Pool) Ping(ctx context.Context) error {
	if pool == nil || pool.pool == nil {
		return fmt.Errorf("PostgreSQL pool is not initialised")
	}

	pingCtx, cancel := context.WithTimeout(ctx, pool.queryTimeout)
	defer cancel()

	if err := pool.pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return nil
}

// Close releases all database connections.
func (pool *Pool) Close() {
	if pool == nil || pool.pool == nil {
		return
	}

	pool.pool.Close()
}

// Stats returns a stable subset of pgxpool statistics.
func (pool *Pool) Stats() Stats {
	if pool == nil || pool.pool == nil {
		return Stats{}
	}

	stats := pool.pool.Stat()

	return Stats{
		AcquiredConns: stats.AcquiredConns(),
		IdleConns:     stats.IdleConns(),
		TotalConns:    stats.TotalConns(),
		MaxConns:      stats.MaxConns(),
	}
}

// Raw exposes the underlying pool to repository packages inside this module.
// HTTP handlers should depend on repositories rather than calling Raw directly.
func (pool *Pool) Raw() *pgxpool.Pool {
	if pool == nil {
		return nil
	}

	return pool.pool
}
