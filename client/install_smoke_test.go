package client_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yazanabuashour/openhealth/client"
)

func TestExternalConsumerCanRunAgainstLocalRuntime(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	repoRoot := filepath.Dir(cwd)
	moduleDir := t.TempDir()
	databasePath := filepath.Join(t.TempDir(), "openhealth.db")

	writeFile(
		t,
		filepath.Join(moduleDir, "main.go"),
		`package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yazanabuashour/openhealth/client"
)

func main() {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: os.Getenv(client.EnvDatabasePath)})
	if err != nil {
		log.Fatal(err)
	}
	defer api.Close()

	summary, err := api.Summary(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ok %d\n", summary.ActiveMedicationCount)
}
`,
	)

	runCommand(t, moduleDir, nil, "go", "mod", "init", "example.com/openhealth-consumer")
	runCommand(
		t,
		moduleDir,
		nil,
		"go",
		"mod",
		"edit",
		"-replace",
		fmt.Sprintf("github.com/yazanabuashour/openhealth=%s", repoRoot),
	)
	runCommand(t, moduleDir, nil, "go", "get", "github.com/yazanabuashour/openhealth")
	runCommand(t, moduleDir, nil, "go", "mod", "tidy")

	modules := runCommand(t, moduleDir, nil, "go", "list", "-m", "all")
	for _, toolModule := range []string{"github.com/sqlc-dev/sqlc"} {
		if strings.Contains(modules, toolModule) {
			t.Fatalf("tool module %q leaked into external consumer module graph\n%s", toolModule, modules)
		}
	}

	output := runCommand(
		t,
		moduleDir,
		[]string{client.EnvDatabasePath + "=" + databasePath},
		"go",
		"run",
		".",
	)
	if strings.TrimSpace(output) != "ok 0" {
		t.Fatalf("external consumer output = %q, want %q", strings.TrimSpace(output), "ok 0")
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runCommand(t *testing.T, dir string, extraEnv []string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Env = append(cmd.Env, "GOWORK=off")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
	return string(output)
}
