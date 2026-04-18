package client

import (
	"github.com/yazanabuashour/openhealth/internal/localruntime"
)

const (
	EnvDataDir      = localruntime.EnvDataDir
	EnvDatabasePath = localruntime.EnvDatabasePath
)

type LocalConfig = localruntime.Config
type LocalPaths = localruntime.Paths

type LocalClient struct {
	Paths LocalPaths

	session *localruntime.Session
}

func ResolveLocalPaths(config LocalConfig) (LocalPaths, error) {
	return localruntime.ResolvePaths(localruntime.Config(config))
}

func OpenLocal(config LocalConfig) (*LocalClient, error) {
	session, err := localruntime.Open(localruntime.Config(config))
	if err != nil {
		return nil, err
	}

	return &LocalClient{
		Paths:   LocalPaths(session.Paths),
		session: session,
	}, nil
}

func (c *LocalClient) Close() error {
	if c == nil || c.session == nil {
		return nil
	}
	return c.session.Close()
}
