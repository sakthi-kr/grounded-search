package config

import (
	"testing"
	"time"
)

func TestLoadFromLookupUsesDatabaseDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromLookup(mapLookup(nil))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.DatabaseURL != defaultDatabaseURL {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, defaultDatabaseURL)
	}
	if cfg.DatabaseTimeout != defaultDatabaseTimeout {
		t.Errorf("DatabaseTimeout = %v, want %v", cfg.DatabaseTimeout, defaultDatabaseTimeout)
	}
	if cfg.DatabaseMaxConns != defaultDatabaseMaxConns {
		t.Errorf("DatabaseMaxConns = %d, want %d", cfg.DatabaseMaxConns, defaultDatabaseMaxConns)
	}
	if cfg.DatabaseMinConns != defaultDatabaseMinConns {
		t.Errorf("DatabaseMinConns = %d, want %d", cfg.DatabaseMinConns, defaultDatabaseMinConns)
	}
}

func TestLoadFromLookupUsesDatabaseOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"DATABASE_URL":             "postgresql://app:secret@db.example.test:5432/search?sslmode=require",
		"DATABASE_CONNECT_TIMEOUT": "750ms",
		"DATABASE_MAX_CONNS":       "24",
		"DATABASE_MIN_CONNS":       "4",
	}))
	if err != nil {
		t.Fatalf("LoadFromLookup() error = %v", err)
	}

	if cfg.DatabaseURL != "postgresql://app:secret@db.example.test:5432/search?sslmode=require" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.DatabaseTimeout != 750*time.Millisecond {
		t.Errorf("DatabaseTimeout = %v, want 750ms", cfg.DatabaseTimeout)
	}
	if cfg.DatabaseMaxConns != 24 {
		t.Errorf("DatabaseMaxConns = %d, want 24", cfg.DatabaseMaxConns)
	}
	if cfg.DatabaseMinConns != 4 {
		t.Errorf("DatabaseMinConns = %d, want 4", cfg.DatabaseMinConns)
	}
}

func TestLoadFromLookupRejectsInvalidDatabaseValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "invalid scheme",
			values: map[string]string{
				"DATABASE_URL": "http://localhost:5432/search",
			},
		},
		{
			name: "missing host",
			values: map[string]string{
				"DATABASE_URL": "postgres:///search",
			},
		},
		{
			name: "missing username",
			values: map[string]string{
				"DATABASE_URL": "postgres://localhost:5432/search",
			},
		},
		{
			name: "missing database name",
			values: map[string]string{
				"DATABASE_URL": "postgres://app:secret@localhost:5432",
			},
		},
		{
			name: "fragment not allowed",
			values: map[string]string{
				"DATABASE_URL": "postgres://app:secret@localhost:5432/search#fragment",
			},
		},
		{
			name: "non-positive timeout",
			values: map[string]string{
				"DATABASE_CONNECT_TIMEOUT": "0s",
			},
		},
		{
			name: "max connections too small",
			values: map[string]string{
				"DATABASE_MAX_CONNS": "0",
			},
		},
		{
			name: "minimum exceeds maximum",
			values: map[string]string{
				"DATABASE_MAX_CONNS": "5",
				"DATABASE_MIN_CONNS": "6",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := LoadFromLookup(mapLookup(test.values)); err == nil {
				t.Fatal("LoadFromLookup() error = nil, want error")
			}
		})
	}
}
