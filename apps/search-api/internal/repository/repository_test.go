package repository

import "testing"

func TestNewRejectsNilPool(t *testing.T) {
	t.Parallel()
	if _, err := New(nil); err == nil {
		t.Fatal("expected error")
	}
}
