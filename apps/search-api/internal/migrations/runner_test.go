package migrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverOrdersMigrations(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for _, name := range []string{"000003_acl.up.sql", "000001_identity.up.sql", "000002_documents.up.sql", "000001_identity.down.sql", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	up, err := Discover(directory, "up")
	if err != nil {
		t.Fatal(err)
	}
	if len(up) != 3 || up[0].Version != 1 || up[2].Version != 3 {
		t.Fatalf("unexpected up migrations: %+v", up)
	}
	down, err := Discover(directory, "down")
	if err != nil {
		t.Fatal(err)
	}
	if len(down) != 1 || down[0].Version != 1 {
		t.Fatalf("unexpected down migrations: %+v", down)
	}
}

func TestNewRejectsInvalidArguments(t *testing.T) {
	t.Parallel()
	if _, err := New(nil, "db/migrations"); err == nil {
		t.Fatal("expected nil pool error")
	}
}

func TestStripTransactionWrapper(t *testing.T) {
	t.Parallel()
	got := stripTransactionWrapper("BEGIN;\nCREATE TABLE example(id integer);\nCOMMIT;")
	want := "CREATE TABLE example(id integer);"
	if got != want {
		t.Fatalf("stripTransactionWrapper() = %q, want %q", got, want)
	}
}
