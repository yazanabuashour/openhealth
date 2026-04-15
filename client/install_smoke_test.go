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

func TestClientExampleRunsAgainstLocalRuntime(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "run", "./examples/client_summary")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), client.EnvDataDir+"="+t.TempDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run example: %v\n%s", err, output)
	}
}

func TestExternalConsumerCanRunAgainstLocalRuntime(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	repoRoot := filepath.Dir(cwd)
	moduleDir := t.TempDir()
	dataDir := t.TempDir()

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
	api, err := client.OpenLocal(client.LocalConfig{DataDir: os.Getenv(client.EnvDataDir)})
	if err != nil {
		log.Fatal(err)
	}
	defer api.Close()

	summary, err := api.GetHealthSummaryWithResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if summary.JSON200 == nil {
		log.Fatalf("unexpected status: %s", summary.Status())
	}

	fmt.Println("ok")
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
	for _, toolModule := range []string{
		"github.com/oapi-codegen/oapi-codegen/v2",
		"github.com/sqlc-dev/sqlc",
	} {
		if strings.Contains(modules, toolModule) {
			t.Fatalf("tool module %q leaked into external consumer module graph\n%s", toolModule, modules)
		}
	}

	output := runCommand(
		t,
		moduleDir,
		[]string{client.EnvDataDir + "=" + dataDir},
		"go",
		"run",
		".",
	)
	if strings.TrimSpace(output) != "ok" {
		t.Fatalf("external consumer output = %q, want %q", strings.TrimSpace(output), "ok")
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
