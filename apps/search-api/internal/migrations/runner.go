package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrationNamePattern = regexp.MustCompile(`^(\d{6})_.+\.(up|down)\.sql$`)

type Migration struct {
	Version   int
	Name      string
	Direction string
	Path      string
}

type Runner struct {
	pool      *pgxpool.Pool
	directory string
}

func New(pool *pgxpool.Pool, directory string) (*Runner, error) {
	if pool == nil {
		return nil, fmt.Errorf("migration pool is nil")
	}
	if strings.TrimSpace(directory) == "" {
		return nil, fmt.Errorf("migration directory is empty")
	}
	return &Runner{pool: pool, directory: directory}, nil
}

func (r *Runner) EnsureTable(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
        version bigint PRIMARY KEY,
        name text NOT NULL,
        applied_at timestamptz NOT NULL DEFAULT now()
    )`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func Discover(directory, direction string) ([]Migration, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}
	migrations := make([]Migration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationNamePattern.FindStringSubmatch(entry.Name())
		if matches == nil || matches[2] != direction {
			continue
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version: %w", err)
		}
		migrations = append(migrations, Migration{Version: version, Name: entry.Name(), Direction: direction, Path: filepath.Join(directory, entry.Name())})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	if direction == "down" {
		sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version > migrations[j].Version })
	}
	return migrations, nil
}

func (r *Runner) Up(ctx context.Context) error {
	if err := r.EnsureTable(ctx); err != nil {
		return err
	}
	files, err := Discover(r.directory, "up")
	if err != nil {
		return err
	}
	for _, migration := range files {
		var exists bool
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, migration.Version).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}
		if exists {
			continue
		}
		if err := r.apply(ctx, migration, true); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Down(ctx context.Context) error {
	if err := r.EnsureTable(ctx); err != nil {
		return err
	}
	var version int
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("read current migration version: %w", err)
	}
	if version == 0 {
		return nil
	}
	files, err := Discover(r.directory, "down")
	if err != nil {
		return err
	}
	for _, migration := range files {
		if migration.Version == version {
			return r.apply(ctx, migration, false)
		}
	}
	return fmt.Errorf("down migration for version %d not found", version)
}

func (r *Runner) Version(ctx context.Context) (int, error) {
	if err := r.EnsureTable(ctx); err != nil {
		return 0, err
	}
	var version int
	if err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&version); err != nil {
		return 0, fmt.Errorf("read migration version: %w", err)
	}
	return version, nil
}

func (r *Runner) apply(ctx context.Context, migration Migration, up bool) error {
	sqlBytes, err := os.ReadFile(migration.Path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.Name, err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Name, err)
	}
	defer tx.Rollback(ctx)
	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err = tx.Exec(migrationCtx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", migration.Name, err)
	}
	if up {
		_, err = tx.Exec(migrationCtx, `INSERT INTO schema_migrations(version,name) VALUES($1,$2)`, migration.Version, migration.Name)
	} else {
		_, err = tx.Exec(migrationCtx, `DELETE FROM schema_migrations WHERE version=$1`, migration.Version)
	}
	if err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Name, err)
	}
	if err = tx.Commit(migrationCtx); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Name, err)
	}
	return nil
}

func stripTransactionWrapper(statement string) string {
	trimmed := strings.TrimSpace(statement)
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "BEGIN;") {
		trimmed = strings.TrimSpace(trimmed[len("BEGIN;"):])
		upper = strings.ToUpper(trimmed)
	}
	if strings.HasSuffix(upper, "COMMIT;") {
		trimmed = strings.TrimSpace(trimmed[:len(trimmed)-len("COMMIT;")])
	}
	return trimmed
}
