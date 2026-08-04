package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/config"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "migration error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: migrate <up|down|version>")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	defer pool.Close()
	directory := os.Getenv("MIGRATIONS_DIR")
	if directory == "" {
		directory = filepath.FromSlash("../../db/migrations")
	}
	runner, err := migrations.New(pool, directory)
	if err != nil {
		return err
	}
	switch args[0] {
	case "up":
		return runner.Up(ctx)
	case "down":
		return runner.Down(ctx)
	case "version":
		version, err := runner.Version(ctx)
		if err == nil {
			fmt.Println(version)
		}
		return err
	default:
		return fmt.Errorf("unknown migration command %q", args[0])
	}
}
