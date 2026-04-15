package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMigrateCreatesDefaultDataDir(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("OPENHEALTH_DATA_DIR", "")
	t.Setenv("OPENHEALTH_DATABASE_PATH", "")

	var stdout bytes.Buffer
	if err := runMigrate(nil, &stdout); err != nil {
		t.Fatalf("run migrate: %v", err)
	}

	wantDatabasePath := filepath.Join(homeDir, ".local", "share", "openhealth", "openhealth.db")
	if _, err := os.Stat(wantDatabasePath); err != nil {
		t.Fatalf("stat migrated database: %v", err)
	}
	if !strings.Contains(stdout.String(), wantDatabasePath) {
		t.Fatalf("stdout = %q, want path %q", stdout.String(), wantDatabasePath)
	}
}

func TestRunMigrateHonorsExplicitDBPathWithoutHomeDir(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("OPENHEALTH_DATA_DIR", "")
	t.Setenv("OPENHEALTH_DATABASE_PATH", "")

	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	var stdout bytes.Buffer
	if err := runMigrate([]string{"-db", databasePath}, &stdout); err != nil {
		t.Fatalf("run migrate with explicit db path: %v", err)
	}

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("stat explicit database path: %v", err)
	}
	if !strings.Contains(stdout.String(), databasePath) {
		t.Fatalf("stdout = %q, want path %q", stdout.String(), databasePath)
	}
}
