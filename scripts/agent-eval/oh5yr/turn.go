package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type parsedTurn struct {
	metrics      metrics
	finalMessage string
	sessionID    string
	parseError   error
	parseSeconds float64
}

func runScenarioTurn(runRepo string, runDir string, dbPath string, currentVariant variant, currentScenario scenario, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig) (turnResult, parsedTurn, error) {
	turnDir := filepath.Join(runDir, fmt.Sprintf("turn-%d", turnIndex))
	if err := os.MkdirAll(turnDir, 0o755); err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	eventsPath := filepath.Join(turnDir, "events.jsonl")
	stderrPath := filepath.Join(turnDir, "stderr.log")
	stdoutFile, err := os.Create(eventsPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() {
		_ = stdoutFile.Close()
	}()
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return turnResult{}, parsedTurn{}, err
	}
	defer func() {
		_ = stderrFile.Close()
	}()

	args := codexArgsForTurn(runRepo, runDir, currentScenario, turn, turnIndex, sessionID, cache)
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Dir = runRepo
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Stdin = strings.NewReader("")
	cmd.Env = evalEnv(runDir, dbPath, cache)

	start := time.Now()
	err = cmd.Run()
	wallSeconds := roundSeconds(time.Since(start).Seconds())
	exitCode := commandExitCode(err)
	if ctx.Err() == context.DeadlineExceeded {
		exitCode = -1
	}

	parseStart := time.Now()
	parsedMetrics, parseErr := parseMetrics(eventsPath)
	parseSeconds := roundSeconds(time.Since(parseStart).Seconds())
	parsed := parsedTurn{
		metrics:      parsedMetrics.metrics,
		finalMessage: parsedMetrics.finalMessage,
		sessionID:    parsedMetrics.sessionID,
		parseError:   parseErr,
		parseSeconds: parseSeconds,
	}
	result := turnResult{
		Index:                   turnIndex,
		WallSeconds:             wallSeconds,
		ExitCode:                exitCode,
		Metrics:                 parsedMetrics.metrics,
		RawLogArtifactReference: fmt.Sprintf("<run-root>/%s/%s/turn-%d/events.jsonl", currentVariant.ID, currentScenario.ID, turnIndex),
	}
	return result, parsed, err
}

func codexArgsForTurn(runRepo string, runDir string, currentScenario scenario, turn scenarioTurn, turnIndex int, sessionID string, cache cacheConfig) []string {
	baseConfig := []string{
		"-m", modelName,
		"-c", fmt.Sprintf("model_reasoning_effort=%q", reasoningEffort),
		"-c", "shell_environment_policy.inherit=all",
	}
	writableRoots := codexWritableRoots(runDir, cache)
	if len(scenarioTurns(currentScenario)) == 1 {
		args := []string{
			"exec",
			"--json",
			"--ephemeral",
			"--full-auto",
			"--skip-git-repo-check",
			"--ignore-user-config",
			"-C", runRepo,
		}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	if turnIndex == 1 {
		args := []string{
			"exec",
			"--json",
			"--full-auto",
			"--skip-git-repo-check",
			"--ignore-user-config",
			"-C", runRepo,
		}
		args = appendAddDirs(args, writableRoots)
		args = append(args, baseConfig...)
		return append(args, turn.Prompt)
	}
	args := []string{
		"exec",
		"-C", runRepo,
	}
	args = appendAddDirs(args, writableRoots)
	args = append(args,
		"resume",
		"--json",
		"--full-auto",
		"--skip-git-repo-check",
		"--ignore-user-config",
	)
	args = append(args, baseConfig...)
	args = append(args, sessionID, turn.Prompt)
	return args
}

func codexWritableRoots(runDir string, cache cacheConfig) []string {
	roots := []string{runDir}
	if cache.Mode == cacheModeShared {
		roots = append(roots, filepath.Join(cache.RunRoot, "shared-cache"))
	}
	return roots
}

func appendAddDirs(args []string, roots []string) []string {
	for _, root := range roots {
		args = append(args, "--add-dir", root)
	}
	return args
}

func evalEnv(runDir string, dbPath string, cache cacheConfig) []string {
	env := evalCodexEnv(cache.RunRoot)
	paths := evalPathsFor(runDir, cache)
	binDir := filepath.Join(runDir, "bin")
	env = envWithOverride(env, "OPENHEALTH_DATABASE_PATH", dbPath)
	env = envWithOverride(env, "GOCACHE", paths.GoCache)
	env = envWithOverride(env, "GOMODCACHE", paths.GoModCache)
	env = envWithOverride(env, "TMPDIR", paths.Temp)
	env = envWithOverride(env, "PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return env
}

func evalCodexEnv(runRoot string) []string {
	return envWithOverride(os.Environ(), "CODEX_HOME", evalCodexHome(runRoot))
}

func envWithOverride(env []string, key string, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			out = append(out, entry)
		}
	}
	return append(out, prefix+value)
}

func evalCodexHome(runRoot string) string {
	return filepath.Join(runRoot, "codex-home")
}

func sourceCodexHome() (string, error) {
	if home := os.Getenv("CODEX_HOME"); home != "" {
		return home, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

func setupEvalCodexHome(runRoot string) error {
	sourceHome, err := sourceCodexHome()
	if err != nil {
		return err
	}
	return setupEvalCodexHomeFromSource(runRoot, sourceHome)
}

func setupEvalCodexHomeFromSource(runRoot string, sourceHome string) error {
	authBytes, err := os.ReadFile(filepath.Join(sourceHome, "auth.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("missing Codex auth at %s; run codex login before running evals", filepath.Join(sourceHome, "auth.json"))
		}
		return err
	}
	codexHome := evalCodexHome(runRoot)
	if err := os.RemoveAll(codexHome); err != nil {
		return err
	}
	if err := os.MkdirAll(codexHome, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(codexHome, "auth.json"), authBytes, 0o600); err != nil {
		return err
	}
	return nil
}

func buildProductionBinary(runRepo string, runDir string, dbPath string, cache cacheConfig) error {
	binDir := filepath.Join(runDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	outputPath := filepath.Join(binDir, "openhealth")
	cmd := exec.Command("go", "build", "-o", outputPath, "./cmd/openhealth")
	cmd.Dir = runRepo
	cmd.Env = evalEnv(runDir, dbPath, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func warmGoModules(runRepo string, runDir string, dbPath string, cache cacheConfig) error {
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = runRepo
	cmd.Env = evalEnv(runDir, dbPath, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmSharedCache(repoRoot string, cache cacheConfig) error {
	paths := sharedEvalPaths(cache)
	for _, dir := range []string{paths.GoCache, paths.GoModCache, paths.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	dbPath := filepath.Join(filepath.Dir(paths.Temp), "prewarm.db")
	if err := warmGoModules(repoRoot, filepath.Dir(paths.Temp), dbPath, cache); err != nil {
		return err
	}
	cmd := exec.Command("go", prewarmCompileArgs()...)
	cmd.Dir = repoRoot
	cmd.Env = evalEnv(filepath.Dir(paths.Temp), dbPath, cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prewarmCompileArgs() []string {
	args := []string{"test", "-run", "^$"}
	return append(args, prewarmCompilePackages...)
}
