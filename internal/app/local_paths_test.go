package app

import (
	"errors"
	"testing"
)

func TestResolveLocalPathsUsesExplicitConfigOverrides(t *testing.T) {
	t.Parallel()

	dataDir, databasePath, err := resolveLocalPaths(LocalPathConfig{
		DataDir: "/tmp/openhealth",
	}, localPathRuntime{
		getenv: func(string) string { return "" },
		userHomeDir: func() (string, error) {
			return "/home/tester", nil
		},
	})
	if err != nil {
		t.Fatalf("resolve local paths: %v", err)
	}
	if dataDir != "/tmp/openhealth" {
		t.Fatalf("dataDir = %q, want %q", dataDir, "/tmp/openhealth")
	}
	if databasePath != "/tmp/openhealth/openhealth.db" {
		t.Fatalf("databasePath = %q, want %q", databasePath, "/tmp/openhealth/openhealth.db")
	}

	dataDir, databasePath, err = resolveLocalPaths(LocalPathConfig{
		DatabasePath: "custom/openhealth.sqlite",
	}, localPathRuntime{
		getenv: func(string) string { return EnvDatabasePath + "-ignored" },
		userHomeDir: func() (string, error) {
			return "/home/tester", nil
		},
	})
	if err != nil {
		t.Fatalf("resolve local paths with explicit database path: %v", err)
	}
	if dataDir != "custom" {
		t.Fatalf("dataDir = %q, want %q", dataDir, "custom")
	}
	if databasePath != "custom/openhealth.sqlite" {
		t.Fatalf("databasePath = %q, want %q", databasePath, "custom/openhealth.sqlite")
	}
}

func TestResolveLocalPathsUsesOSDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		env              map[string]string
		homeDir          string
		wantDataDir      string
		wantDatabasePath string
	}{
		{
			name:             "xdg data home",
			env:              map[string]string{"XDG_DATA_HOME": "/tmp/data-home"},
			homeDir:          "/home/tester",
			wantDataDir:      "/tmp/data-home/openhealth",
			wantDatabasePath: "/tmp/data-home/openhealth/openhealth.db",
		},
		{
			name:             "fallback",
			env:              map[string]string{},
			homeDir:          "/home/tester",
			wantDataDir:      "/home/tester/.local/share/openhealth",
			wantDatabasePath: "/home/tester/.local/share/openhealth/openhealth.db",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dataDir, databasePath, err := resolveLocalPaths(LocalPathConfig{}, localPathRuntime{
				getenv: func(key string) string {
					return tc.env[key]
				},
				userHomeDir: func() (string, error) {
					return tc.homeDir, nil
				},
			})
			if err != nil {
				t.Fatalf("resolve local paths: %v", err)
			}
			if dataDir != tc.wantDataDir {
				t.Fatalf("dataDir = %q, want %q", dataDir, tc.wantDataDir)
			}
			if databasePath != tc.wantDatabasePath {
				t.Fatalf("databasePath = %q, want %q", databasePath, tc.wantDatabasePath)
			}
		})
	}
}

func TestResolveLocalPathsSupportsEnvOverrides(t *testing.T) {
	t.Parallel()

	dataDir, databasePath, err := resolveLocalPaths(LocalPathConfig{}, localPathRuntime{
		getenv: func(key string) string {
			switch key {
			case EnvDatabasePath:
				return "/tmp/override/custom.db"
			case EnvDataDir:
				return "/tmp/override-data"
			default:
				return ""
			}
		},
		userHomeDir: func() (string, error) {
			return "/home/tester", nil
		},
	})
	if err != nil {
		t.Fatalf("resolve local paths: %v", err)
	}
	if dataDir != "/tmp/override" {
		t.Fatalf("dataDir = %q, want %q", dataDir, "/tmp/override")
	}
	if databasePath != "/tmp/override/custom.db" {
		t.Fatalf("databasePath = %q, want %q", databasePath, "/tmp/override/custom.db")
	}
}

func TestResolveLocalPathsPropagatesHomeDirErrors(t *testing.T) {
	t.Parallel()

	_, _, err := resolveLocalPaths(LocalPathConfig{}, localPathRuntime{
		getenv: func(string) string { return "" },
		userHomeDir: func() (string, error) {
			return "", errors.New("boom")
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
