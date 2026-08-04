package main

import "testing"

func TestRunRejectsArguments(t *testing.T) {
	t.Setenv("DATABASE_URL", "://invalid")
	if err := run(nil); err == nil {
		t.Fatal("expected usage error")
	}
	if err := run([]string{"unknown"}); err == nil {
		t.Fatal("expected configuration or command error")
	}
}
