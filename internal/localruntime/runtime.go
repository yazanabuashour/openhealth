package localruntime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yazanabuashour/openhealth/internal/app"
	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

const (
	EnvDatabasePath = app.EnvDatabasePath
)

type Config struct {
	DatabasePath string
	Timeout      time.Duration
}

type Paths struct {
	DataDir      string
	DatabasePath string
}

type Session struct {
	Paths   Paths
	Service health.Service

	close func() error
}

func ResolvePaths(config Config) (Paths, error) {
	dataDir, databasePath, err := app.ResolveLocalPaths(app.LocalPathConfig{
		DatabasePath: config.DatabasePath,
	})
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		DataDir:      dataDir,
		DatabasePath: databasePath,
	}, nil
}

func Open(config Config) (*Session, error) {
	paths, err := ResolvePaths(config)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(paths.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create local data directory %s: %w", paths.DataDir, err)
	}

	db, err := sqlite.Open(paths.DatabasePath)
	if err != nil {
		return nil, err
	}

	if err := sqlite.ApplyMigrations(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Session{
		Paths:   paths,
		Service: health.NewService(sqlite.NewRepository(db)),
		close:   db.Close,
	}, nil
}

func (s *Session) Close() error {
	if s == nil || s.close == nil {
		return nil
	}
	return s.close()
}
