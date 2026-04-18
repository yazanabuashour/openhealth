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

func TestRunWeightAddAndList(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	for _, args := range [][]string{
		{"-db", databasePath, "--date", "2026-03-28", "--value", "153.0"},
		{"-db", databasePath, "--date", "2026-03-29", "--value", "152.2"},
		{"-db", databasePath, "--date", "2026-03-30", "--value", "151.6", "--unit", "lb"},
	} {
		var stdout bytes.Buffer
		if err := runWeightAdd(args, &stdout); err != nil {
			t.Fatalf("run weight add %v: %v", args, err)
		}
		if !strings.Contains(stdout.String(), " lb created") {
			t.Fatalf("stdout = %q, want created status", stdout.String())
		}
	}

	var stdout bytes.Buffer
	if err := runWeightList([]string{"-db", databasePath, "--from", "2026-03-29", "--to", "2026-03-30"}, &stdout); err != nil {
		t.Fatalf("run weight list: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "2026-03-30 151.6 lb") || !strings.Contains(got, "2026-03-29 152.2 lb") {
		t.Fatalf("stdout = %q, want bounded rows", got)
	}
	if strings.Contains(got, "2026-03-28") {
		t.Fatalf("stdout = %q, did not expect 2026-03-28", got)
	}
	if strings.Index(got, "2026-03-30") > strings.Index(got, "2026-03-29") {
		t.Fatalf("stdout = %q, want newest first", got)
	}
}

func TestRunWeightAddRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{
			name: "short date",
			args: []string{"--date", "03/29", "--value", "152.2"},
		},
		{
			name: "unsupported unit",
			args: []string{"--date", "2026-03-29", "--value", "152.2", "--unit", "stone"},
		},
		{
			name: "nonpositive value",
			args: []string{"--date", "2026-03-29", "--value", "-5"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")
			args := append([]string{"-db", databasePath}, tt.args...)
			var stdout bytes.Buffer
			if err := runWeightAdd(args, &stdout); err == nil {
				t.Fatalf("runWeightAdd(%v) succeeded, want error", args)
			}
			if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
				t.Fatalf("database stat error = %v, want not exist", err)
			}
		})
	}
}

func TestRunBloodPressureAddAndList(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")

	for _, args := range [][]string{
		{"-db", databasePath, "--date", "2026-03-28", "--systolic", "124", "--diastolic", "80"},
		{"-db", databasePath, "--date", "2026-03-29", "--systolic", "122", "--diastolic", "78", "--pulse", "64"},
		{"-db", databasePath, "--date", "2026-03-30", "--systolic", "118", "--diastolic", "76"},
	} {
		var stdout bytes.Buffer
		if err := runBloodPressureAdd(args, &stdout); err != nil {
			t.Fatalf("run blood-pressure add %v: %v", args, err)
		}
		if !strings.Contains(stdout.String(), " created") {
			t.Fatalf("stdout = %q, want created status", stdout.String())
		}
	}

	var stdout bytes.Buffer
	if err := runBloodPressureList([]string{"-db", databasePath, "--from", "2026-03-29", "--to", "2026-03-30"}, &stdout); err != nil {
		t.Fatalf("run blood-pressure list: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "2026-03-30 118/76") || !strings.Contains(got, "2026-03-29 122/78 pulse 64") {
		t.Fatalf("stdout = %q, want bounded rows", got)
	}
	if strings.Contains(got, "2026-03-28") {
		t.Fatalf("stdout = %q, did not expect 2026-03-28", got)
	}
	if strings.Index(got, "2026-03-30") > strings.Index(got, "2026-03-29") {
		t.Fatalf("stdout = %q, want newest first", got)
	}
}

func TestRunBloodPressureAddRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
	}{
		{
			name: "short date",
			args: []string{"--date", "03/29", "--systolic", "122", "--diastolic", "78"},
		},
		{
			name: "nonpositive systolic",
			args: []string{"--date", "2026-03-29", "--systolic", "0", "--diastolic", "78"},
		},
		{
			name: "nonpositive diastolic",
			args: []string{"--date", "2026-03-29", "--systolic", "122", "--diastolic", "-1"},
		},
		{
			name: "nonpositive pulse",
			args: []string{"--date", "2026-03-29", "--systolic", "122", "--diastolic", "78", "--pulse", "0"},
		},
		{
			name: "nonpositive pulse single dash",
			args: []string{"--date", "2026-03-29", "--systolic", "122", "--diastolic", "78", "-pulse=0"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			databasePath := filepath.Join(t.TempDir(), "nested", "openhealth.db")
			args := append([]string{"-db", databasePath}, tt.args...)
			var stdout bytes.Buffer
			if err := runBloodPressureAdd(args, &stdout); err == nil {
				t.Fatalf("runBloodPressureAdd(%v) succeeded, want error", args)
			}
			if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
				t.Fatalf("database stat error = %v, want not exist", err)
			}
		})
	}
}
