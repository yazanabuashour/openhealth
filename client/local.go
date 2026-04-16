package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/yazanabuashour/openhealth/internal/app"
	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/httpapi"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

const (
	EnvDataDir      = app.EnvDataDir
	EnvDatabasePath = app.EnvDatabasePath
	localBaseURL    = "http://openhealth.local"
)

type LocalConfig struct {
	DataDir      string
	DatabasePath string
	Timeout      time.Duration
}

type LocalPaths struct {
	DataDir      string
	DatabasePath string
}

type LocalClient struct {
	*ClientWithResponses
	Paths LocalPaths

	service health.Service
	close   func() error
}

func ResolveLocalPaths(config LocalConfig) (LocalPaths, error) {
	dataDir, databasePath, err := app.ResolveLocalPaths(app.LocalPathConfig{
		DataDir:      config.DataDir,
		DatabasePath: config.DatabasePath,
	})
	if err != nil {
		return LocalPaths{}, err
	}

	return LocalPaths{
		DataDir:      dataDir,
		DatabasePath: databasePath,
	}, nil
}

func OpenLocal(config LocalConfig) (*LocalClient, error) {
	paths, err := ResolveLocalPaths(config)
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

	service := health.NewService(sqlite.NewRepository(db))
	httpClient := &http.Client{
		Timeout:   localTimeout(config.Timeout),
		Transport: &localRoundTripper{handler: httpapi.NewHandler(service)},
	}

	api, err := NewClientWithResponses(localBaseURL, WithHTTPClient(httpClient))
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &LocalClient{
		ClientWithResponses: api,
		Paths:               paths,
		service:             service,
		close:               db.Close,
	}, nil
}

func (c *LocalClient) Close() error {
	if c == nil || c.close == nil {
		return nil
	}
	return c.close()
}

type localRoundTripper struct {
	handler http.Handler
}

func (t *localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if t.handler == nil {
		return nil, fmt.Errorf("local handler is required")
	}
	defer closeRequestBody(req.Body)

	request := req.Clone(req.Context())
	request.RequestURI = request.URL.RequestURI()
	if request.Host == "" {
		request.Host = request.URL.Host
	}

	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}

func localTimeout(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return 30 * time.Second
}

func closeRequestBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_ = body.Close()
}
