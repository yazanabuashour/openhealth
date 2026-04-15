package app

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	EnvDataDir      = "OPENHEALTH_DATA_DIR"
	EnvDatabasePath = "OPENHEALTH_DATABASE_PATH"
	defaultDBName   = "openhealth.db"
)

type LocalPathConfig struct {
	DataDir      string
	DatabasePath string
}

type localPathRuntime struct {
	getenv      func(string) string
	userHomeDir func() (string, error)
}

func ResolveLocalPaths(config LocalPathConfig) (string, string, error) {
	return resolveLocalPaths(config, localPathRuntime{
		getenv:      os.Getenv,
		userHomeDir: os.UserHomeDir,
	})
}

func resolveLocalPaths(config LocalPathConfig, rt localPathRuntime) (string, string, error) {
	switch {
	case config.DatabasePath != "":
		return cleanPath(filepath.Dir(config.DatabasePath)), cleanPath(config.DatabasePath), nil
	case config.DataDir != "":
		dataDir := cleanPath(config.DataDir)
		return dataDir, filepath.Join(dataDir, defaultDBName), nil
	}

	if databasePath := rt.getenv(EnvDatabasePath); databasePath != "" {
		return cleanPath(filepath.Dir(databasePath)), cleanPath(databasePath), nil
	}
	if dataDir := rt.getenv(EnvDataDir); dataDir != "" {
		dataDir = cleanPath(dataDir)
		return dataDir, filepath.Join(dataDir, defaultDBName), nil
	}

	homeDir, err := rt.userHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("resolve user home directory: %w", err)
	}

	dataDir := defaultDataDir(cleanPath(homeDir), rt.getenv)
	return dataDir, filepath.Join(dataDir, defaultDBName), nil
}

func defaultDataDir(homeDir string, getenv func(string) string) string {
	if xdgDataHome := getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(cleanPath(xdgDataHome), "openhealth")
	}
	return filepath.Join(homeDir, ".local", "share", "openhealth")
}

func cleanPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}
