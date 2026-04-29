package main

import (
	"io/fs"
	"os"
	"path/filepath"
)

type evalPaths struct {
	GoCache    string
	GoModCache string
	Temp       string
}

func evalPathsFor(runDir string, cache cacheConfig) evalPaths {
	if cache.Mode == cacheModeShared {
		paths := sharedEvalPaths(cache)
		paths.Temp = filepath.Join(runDir, "tmp")
		return paths
	}
	return evalPaths{
		GoCache:    filepath.Join(runDir, "gocache"),
		GoModCache: filepath.Join(runDir, "gomodcache"),
		Temp:       filepath.Join(runDir, "tmp"),
	}
}

func sharedEvalPaths(cache cacheConfig) evalPaths {
	root := filepath.Join(cache.RunRoot, "shared-cache")
	return evalPaths{
		GoCache:    filepath.Join(root, "gocache"),
		GoModCache: filepath.Join(root, "gomodcache"),
		Temp:       filepath.Join(root, "tmp"),
	}
}

func prepareRunDir(runDir string, cache cacheConfig) error {
	_ = filepath.WalkDir(runDir, func(path string, entry fs.DirEntry, err error) error {
		if err == nil {
			_ = os.Chmod(path, 0o755)
		}
		return nil
	})
	if err := os.RemoveAll(runDir); err != nil {
		return err
	}
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	paths := evalPathsFor(runDir, cache)
	dirs := []string{paths.Temp}
	if cache.Mode == cacheModeIsolated {
		dirs = append(dirs, paths.GoCache, paths.GoModCache)
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
