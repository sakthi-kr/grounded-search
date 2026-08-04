//go:build integration

package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/authorization"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/migrations"
)

func TestDatabaseBackedAuthorization(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	runner, _ := migrations.New(pool, filepath.Join("..", "..", "..", "..", "db", "migrations"))
	if err := runner.Up(ctx); err != nil {
		t.Fatal(err)
	}
	seed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "db", "seeds", "development.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, string(seed)); err != nil {
		t.Fatal(err)
	}
	repo, _ := New(pool)
	user, doc, err := repo.LoadAuthorization(ctx, "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000001002")
	if err != nil {
		t.Fatal(err)
	}
	result := authorization.Evaluate("00000000-0000-0000-0000-000000000001", user, doc)
	if result.Allowed || result.Reason != authorization.ReasonExplicitUserDeny {
		t.Fatalf("unexpected result: %+v", result)
	}
}
