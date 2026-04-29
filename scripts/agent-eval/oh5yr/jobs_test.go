package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexArgsExposeWritableRootsForSharedCacheAndResume(t *testing.T) {
	t.Parallel()

	cache := cacheConfig{Mode: cacheModeShared, RunRoot: filepath.Join("run-root")}
	single := scenario{ID: "single", Prompt: "single prompt"}
	singleArgs := codexArgsForTurn("run-root/production/single/repo", "run-root/production/single", single, scenarioTurn{Prompt: "single prompt"}, 1, "", cache)
	if !containsArgPair(singleArgs, "--add-dir", "run-root/production/single") {
		t.Fatalf("single args missing run dir writable root: %v", singleArgs)
	}
	if !containsArgPair(singleArgs, "--add-dir", filepath.Join("run-root", "shared-cache")) {
		t.Fatalf("single args missing shared cache writable root: %v", singleArgs)
	}
	if !containsArg(singleArgs, "--ephemeral") {
		t.Fatalf("single args missing --ephemeral: %v", singleArgs)
	}
	if !containsArg(singleArgs, "--ignore-user-config") {
		t.Fatalf("single args missing --ignore-user-config: %v", singleArgs)
	}

	multi := scenario{ID: "multi", Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
	firstTurnArgs := codexArgsForTurn("run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "first"}, 1, "", cache)
	if !containsArg(firstTurnArgs, "--ignore-user-config") {
		t.Fatalf("first multi-turn args missing --ignore-user-config: %v", firstTurnArgs)
	}
	if containsArg(firstTurnArgs, "--ephemeral") {
		t.Fatalf("first multi-turn args must persist the session: %v", firstTurnArgs)
	}

	resumeArgs := codexArgsForTurn("run-root/production/multi/repo", "run-root/production/multi", multi, scenarioTurn{Prompt: "second"}, 2, "session-123", cache)
	if len(resumeArgs) < 5 || resumeArgs[0] != "exec" || resumeArgs[1] != "-C" || resumeArgs[2] != "run-root/production/multi/repo" {
		t.Fatalf("resume args must set exec workspace before resume: %v", resumeArgs)
	}
	if !containsArgPair(resumeArgs, "--add-dir", "run-root/production/multi") {
		t.Fatalf("resume args missing run dir writable root: %v", resumeArgs)
	}
	if !containsArgPair(resumeArgs, "--add-dir", filepath.Join("run-root", "shared-cache")) {
		t.Fatalf("resume args missing shared cache writable root: %v", resumeArgs)
	}
	if containsArg(resumeArgs, "--ephemeral") {
		t.Fatalf("resume args must persist the multi-turn session: %v", resumeArgs)
	}
	if !containsArg(resumeArgs, "--ignore-user-config") {
		t.Fatalf("resume args missing --ignore-user-config: %v", resumeArgs)
	}
	if resumeArgs[len(resumeArgs)-2] != "session-123" || resumeArgs[len(resumeArgs)-1] != "second" {
		t.Fatalf("resume args must end with session id and prompt: %v", resumeArgs)
	}
}

func TestRunEvalJobsPreservesOrderAndHarnessErrors(t *testing.T) {
	t.Parallel()

	selectedVariants := []variant{
		{ID: "production", Title: "Production"},
	}
	selectedScenarios := []scenario{
		{ID: "slow", Title: "Slow scenario", Prompt: "slow prompt"},
		{ID: "fast", Title: "Fast scenario", Prompt: "fast prompt"},
	}
	jobs := evalJobsFor(selectedVariants, selectedScenarios)
	results := runEvalJobs("repo", "run", jobs, 2, cacheConfig{Mode: cacheModeIsolated, RunRoot: "run"}, func(_ string, _ string, currentVariant variant, currentScenario scenario, _ cacheConfig) (runResult, error) {
		if currentScenario.ID == "slow" {
			time.Sleep(20 * time.Millisecond)
		}
		if currentVariant.ID == "production" && currentScenario.ID == "fast" {
			return runResult{}, errors.New("boom")
		}
		return runResult{
			Variant:       currentVariant.ID,
			Scenario:      currentScenario.ID,
			ScenarioTitle: currentScenario.Title,
			Passed:        true,
		}, nil
	})

	wantKeys := []string{"production/slow", "production/fast"}
	if len(results) != len(wantKeys) {
		t.Fatalf("results = %d, want %d", len(results), len(wantKeys))
	}
	for i, want := range wantKeys {
		if got := resultKey(results[i].Variant, results[i].Scenario); got != want {
			t.Fatalf("result %d key = %q, want %q; results = %#v", i, got, want, results)
		}
	}
	failed := results[1]
	if failed.ExitCode != -1 || failed.Verification.Passed || !strings.Contains(failed.Verification.Details, "harness error: boom") {
		t.Fatalf("failed result = %#v, want harness error", failed)
	}
	if failed.PromptSummary == "" {
		t.Fatalf("failed result prompt summary is empty: %#v", failed)
	}
}

func TestEvalJobsProductionOnly(t *testing.T) {
	t.Parallel()

	jobs := evalJobsFor([]variant{
		{ID: "production", Title: "Production"},
	}, []scenario{
		{ID: "medication-add-list", Title: "Medication", Prompt: "prompt"},
	})

	if len(jobs) != 1 {
		t.Fatalf("jobs = %#v, want one production job", jobs)
	}
	if jobs[0].Variant.ID != "production" || jobs[0].Scenario.ID != "medication-add-list" {
		t.Fatalf("job = %#v, want production medication-add-list", jobs[0])
	}
}

func TestHistoryIsolationStatusFlagsUnexpectedSessionCounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		newSessionFiles    int
		expectedPersisted  int
		wantStatus         string
		wantLimitationPart string
	}{
		{
			name:              "single turn no sessions",
			newSessionFiles:   0,
			expectedPersisted: 0,
			wantStatus:        "passed",
		},
		{
			name:               "single turn extra session",
			newSessionFiles:    1,
			expectedPersisted:  0,
			wantStatus:         "review",
			wantLimitationPart: "only single-turn ephemeral",
		},
		{
			name:              "multi turn exact sessions",
			newSessionFiles:   2,
			expectedPersisted: 2,
			wantStatus:        "passed",
		},
		{
			name:               "multi turn missing session",
			newSessionFiles:    1,
			expectedPersisted:  2,
			wantStatus:         "review",
			wantLimitationPart: "Fewer session files",
		},
		{
			name:               "multi turn extra session",
			newSessionFiles:    3,
			expectedPersisted:  2,
			wantStatus:         "review",
			wantLimitationPart: "More session files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			status, limitation := historyIsolationStatus(tt.newSessionFiles, tt.expectedPersisted)
			if status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", status, tt.wantStatus)
			}
			if tt.wantLimitationPart == "" {
				if limitation != "" {
					t.Fatalf("limitation = %q, want empty", limitation)
				}
				return
			}
			if !strings.Contains(limitation, tt.wantLimitationPart) {
				t.Fatalf("limitation = %q, want to contain %q", limitation, tt.wantLimitationPart)
			}
		})
	}
}

func TestCountNewSessionFilesUsesEvalCodexHome(t *testing.T) {
	t.Parallel()

	runRoot := t.TempDir()
	sessionsDir := filepath.Join(evalCodexHome(runRoot), "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := time.Now()
	oldPath := filepath.Join(sessionsDir, "old.jsonl")
	newPath := filepath.Join(sessionsDir, "new.jsonl")
	otherPath := filepath.Join(sessionsDir, "other.jsonl")
	for path, content := range map[string]string{
		oldPath:   runRoot,
		newPath:   runRoot,
		otherPath: "different run root",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(oldPath, marker.Add(-time.Hour), marker.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, marker.Add(time.Hour), marker.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(otherPath, marker.Add(time.Hour), marker.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	if got := countNewSessionFiles(marker, runRoot); got != 1 {
		t.Fatalf("countNewSessionFiles() = %d, want 1", got)
	}
}

func TestSetupEvalCodexHomeCopiesOnlyAuth(t *testing.T) {
	sourceHome := filepath.Join(t.TempDir(), "source-codex")
	if err := os.MkdirAll(filepath.Join(sourceHome, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "auth.json"), []byte(`{"token":"secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "config.toml"), []byte("model = \"custom\""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceHome, "sessions", "session.jsonl"), []byte("session"), 0o644); err != nil {
		t.Fatal(err)
	}
	runRoot := t.TempDir()
	if err := setupEvalCodexHomeFromSource(runRoot, sourceHome); err != nil {
		t.Fatalf("setupEvalCodexHomeFromSource() error = %v", err)
	}
	codexHome := evalCodexHome(runRoot)
	authPath := filepath.Join(codexHome, "auth.json")
	authBytes, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("read copied auth: %v", err)
	}
	if string(authBytes) != `{"token":"secret"}` {
		t.Fatalf("auth content = %q, want copied source auth", authBytes)
	}
	info, err := os.Lstat(authPath)
	if err != nil {
		t.Fatalf("lstat auth: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("auth copy must not be a symlink")
	}
	for _, unwanted := range []string{"config.toml", filepath.Join("sessions", "session.jsonl")} {
		if _, err := os.Stat(filepath.Join(codexHome, unwanted)); !os.IsNotExist(err) {
			t.Fatalf("unexpected copied %s: stat error = %v", unwanted, err)
		}
	}
	homeInfo, err := os.Stat(codexHome)
	if err != nil {
		t.Fatalf("stat eval codex home: %v", err)
	}
	if homeInfo.Mode().Perm()&0o077 != 0 {
		t.Fatalf("eval codex home permissions = %v, want no group/other access", homeInfo.Mode().Perm())
	}
}

func TestSetupEvalCodexHomeRequiresAuth(t *testing.T) {
	err := setupEvalCodexHomeFromSource(t.TempDir(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "run codex login") {
		t.Fatalf("setupEvalCodexHomeFromSource() error = %v, want login guidance", err)
	}
}

func TestCacheModeEnvPathSelectionAndPrewarmArgs(t *testing.T) {
	t.Parallel()

	runRoot := t.TempDir()
	runDir := filepath.Join(runRoot, "production", "latest-only")
	dbPath := evalDatabasePath(filepath.Join(runDir, "repo"))
	shared := cacheConfig{Mode: cacheModeShared, RunRoot: runRoot}
	isolated := cacheConfig{Mode: cacheModeIsolated, RunRoot: runRoot}

	sharedEnv := strings.Join(evalEnv(runDir, dbPath, shared), "\n")
	for _, want := range []string{
		"CODEX_HOME=" + filepath.Join(runRoot, "codex-home"),
		"OPENHEALTH_DATABASE_PATH=" + filepath.Join(runDir, "repo", "openhealth.db"),
		"GOCACHE=" + filepath.Join(runRoot, "shared-cache", "gocache"),
		"GOMODCACHE=" + filepath.Join(runRoot, "shared-cache", "gomodcache"),
		"TMPDIR=" + filepath.Join(runDir, "tmp"),
		"PATH=" + filepath.Join(runDir, "bin"),
	} {
		if !strings.Contains(sharedEnv, want) {
			t.Fatalf("shared env missing %q in %s", want, sharedEnv)
		}
	}

	isolatedEnv := strings.Join(evalEnv(runDir, dbPath, isolated), "\n")
	for _, want := range []string{
		"GOCACHE=" + filepath.Join(runDir, "gocache"),
		"GOMODCACHE=" + filepath.Join(runDir, "gomodcache"),
		"TMPDIR=" + filepath.Join(runDir, "tmp"),
	} {
		if !strings.Contains(isolatedEnv, want) {
			t.Fatalf("isolated env missing %q in %s", want, isolatedEnv)
		}
	}

	args := strings.Join(prewarmCompileArgs(), " ")
	for _, want := range []string{"test -run ^$", "./cmd/openhealth", "./internal/runner"} {
		if !strings.Contains(args, want) {
			t.Fatalf("prewarm args = %q, want %q", args, want)
		}
	}
}
