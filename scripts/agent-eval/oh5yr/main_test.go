package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/client"
)

func testMetrics(toolCalls int, nonCachedTokens int) metrics {
	input := nonCachedTokens
	cached := 0
	output := 10
	return metrics{
		AssistantCalls:       1,
		ToolCalls:            toolCalls,
		CommandExecutions:    toolCalls,
		UsageExposed:         true,
		InputTokens:          &input,
		CachedInputTokens:    &cached,
		NonCachedInputTokens: &input,
		OutputTokens:         &output,
		EventTypeCounts:      map[string]int{},
	}
}

func codeFirstProductionMetrics(scenarioID string) metrics {
	if isFinalAnswerOnlyValidationScenario(scenarioID) {
		return testMetrics(0, 50)
	}
	return testMetrics(1, 50)
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func containsArgPair(args []string, key string, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestPrepareRunDirResetsAndCreatesRuntimeDirs(t *testing.T) {
	t.Parallel()

	runDir := filepath.Join(t.TempDir(), "production", "add-two")
	if err := os.MkdirAll(filepath.Join(runDir, "repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "openhealth.db"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := cacheConfig{Mode: cacheModeIsolated, RunRoot: t.TempDir()}
	if err := prepareRunDir(runDir, cache); err != nil {
		t.Fatalf("prepareRunDir() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "openhealth.db")); !os.IsNotExist(err) {
		t.Fatalf("stale database stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "repo")); !os.IsNotExist(err) {
		t.Fatalf("stale repo stat error = %v, want not exist", err)
	}

	paths := evalPathsFor(runDir, cache)
	for _, dir := range []string{runDir, paths.GoCache, paths.GoModCache, paths.Temp} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
}

func TestCopyRepoSkipsVariantContaminatingInstructions(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	src := filepath.Join(temp, "src")
	dst := filepath.Join(temp, "dst")
	for _, path := range []string{
		filepath.Join(src, ".agents", "skills", "openhealth"),
		filepath.Join(src, "docs", "agent-eval-results"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{
		filepath.Join(src, "AGENTS.md"):                                     "repo agent instructions",
		filepath.Join(src, "README.md"):                                     "kept",
		filepath.Join(src, ".agents", "skills", "openhealth", "SKILL.md"):   "stale skill",
		filepath.Join(src, "docs", "agent-eval-results", "previous.md"):     "previous report",
		filepath.Join(src, "docs", "agent-evals.md"):                        "eval docs",
		filepath.Join(src, "scripts", "agent-eval", "oh5yr", "main.go"):     "harness",
		filepath.Join(src, "docs", "agent-eval-assets", "variants", "x.md"): "asset",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := copyRepo(src, dst); err != nil {
		t.Fatalf("copyRepo() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Fatalf("kept file stat error = %v", err)
	}
	for _, skipped := range []string{
		"AGENTS.md",
		filepath.Join(".agents", "skills", "openhealth", "SKILL.md"),
		filepath.Join("docs", "agent-eval-results", "previous.md"),
		filepath.Join("docs", "agent-evals.md"),
		filepath.Join("scripts", "agent-eval", "oh5yr", "main.go"),
		filepath.Join("docs", "agent-eval-assets", "variants", "x.md"),
	} {
		if _, err := os.Stat(filepath.Join(dst, skipped)); !os.IsNotExist(err) {
			t.Fatalf("copied skipped path %s: stat error = %v", skipped, err)
		}
	}
}

func TestVariantsIncludeCLI(t *testing.T) {
	t.Parallel()

	ids := map[string]bool{}
	for _, variant := range variants() {
		ids[variant.ID] = true
	}
	for _, want := range []string{"production", "cli"} {
		if !ids[want] {
			t.Fatalf("variants() missing %q: %#v", want, variants())
		}
	}
	for _, retired := range []string{"generated-client", "agentops-code"} {
		if ids[retired] {
			t.Fatalf("variants() includes retired variant %q: %#v", retired, variants())
		}
	}
}

func TestSanitizeMetricEvidenceRedactsCustomRunRoots(t *testing.T) {
	t.Parallel()

	command := "/bin/zsh -lc 'go run ./cmd/openhealth weight list --db /tmp/openhealth-oh743-final-r1/cli/latest-only/openhealth.db'"
	got := sanitizeMetricEvidence(command)
	if strings.Contains(got, "/tmp/") || strings.Contains(got, "openhealth-oh743") {
		t.Fatalf("sanitizeMetricEvidence() = %q, want run root redacted", got)
	}
	if !strings.Contains(got, "<run-root>") {
		t.Fatalf("sanitizeMetricEvidence() = %q, want <run-root>", got)
	}
}

func TestWriteMarkdownReportsRunnableCLIStatus(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, report{
		Issue:           issueID,
		Date:            "2026-04-16",
		Harness:         "test",
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		CodexVersion:    "test",
		CLIStatus:       "runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands",
	}); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Status: `runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands`.") {
		t.Fatalf("markdown = %q, want runnable CLI status", text)
	}
	if strings.Contains(text, "not run because") {
		t.Fatalf("markdown = %q, should not report CLI as skipped", text)
	}
}

func TestParseRunOptionsDefaultsParallelismAndCacheMode(t *testing.T) {
	t.Parallel()

	options, err := parseRunOptions(nil)
	if err != nil {
		t.Fatalf("parseRunOptions: %v", err)
	}
	if options.Parallelism != defaultRunParallelism {
		t.Fatalf("parallelism = %d, want %d", options.Parallelism, defaultRunParallelism)
	}
	if options.CacheMode != cacheModeShared {
		t.Fatalf("cache mode = %q, want %q", options.CacheMode, cacheModeShared)
	}

	options, err = parseRunOptions([]string{"--parallel", "1", "--cache-mode", "isolated"})
	if err != nil {
		t.Fatalf("parseRunOptions --parallel 1 --cache-mode isolated: %v", err)
	}
	if options.Parallelism != 1 {
		t.Fatalf("parallelism = %d, want 1", options.Parallelism)
	}
	if options.CacheMode != cacheModeIsolated {
		t.Fatalf("cache mode = %q, want %q", options.CacheMode, cacheModeIsolated)
	}

	if _, err := parseRunOptions([]string{"--parallel", "0"}); err == nil || !strings.Contains(err.Error(), "parallel must be greater than or equal to 1") {
		t.Fatalf("parseRunOptions --parallel 0 error = %v, want validation error", err)
	}
	if _, err := parseRunOptions([]string{"--cache-mode", "bad"}); err == nil || !strings.Contains(err.Error(), "cache-mode must be") {
		t.Fatalf("parseRunOptions --cache-mode bad error = %v, want validation error", err)
	}
}

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

	multi := scenario{ID: "multi", Turns: []scenarioTurn{{Prompt: "first"}, {Prompt: "second"}}}
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
	if resumeArgs[len(resumeArgs)-2] != "session-123" || resumeArgs[len(resumeArgs)-1] != "second" {
		t.Fatalf("resume args must end with session id and prompt: %v", resumeArgs)
	}
}

func TestRunEvalJobsPreservesOrderAndHarnessErrors(t *testing.T) {
	t.Parallel()

	selectedVariants := []variant{
		{ID: "production", Title: "Production"},
		{ID: "cli", Title: "CLI"},
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
		if currentVariant.ID == "cli" && currentScenario.ID == "fast" {
			return runResult{}, errors.New("boom")
		}
		return runResult{
			Variant:       currentVariant.ID,
			Scenario:      currentScenario.ID,
			ScenarioTitle: currentScenario.Title,
			Passed:        true,
		}, nil
	})

	wantKeys := []string{"production/slow", "production/fast", "cli/slow", "cli/fast"}
	if len(results) != len(wantKeys) {
		t.Fatalf("results = %d, want %d", len(results), len(wantKeys))
	}
	for i, want := range wantKeys {
		if got := resultKey(results[i].Variant, results[i].Scenario); got != want {
			t.Fatalf("result %d key = %q, want %q; results = %#v", i, got, want, results)
		}
	}
	failed := results[3]
	if failed.ExitCode != -1 || failed.Verification.Passed || !strings.Contains(failed.Verification.Details, "harness error: boom") {
		t.Fatalf("failed result = %#v, want harness error", failed)
	}
	if failed.PromptSummary == "" {
		t.Fatalf("failed result prompt summary is empty: %#v", failed)
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

func TestReportIncludesParallelMetadata(t *testing.T) {
	t.Parallel()

	value := report{
		Issue:                 issueID,
		Date:                  "parallel-test",
		Harness:               "test",
		Parallelism:           4,
		CacheMode:             cacheModeShared,
		CachePrewarmSeconds:   1.25,
		HarnessElapsedSeconds: 12.34,
		PhaseTotals:           phaseTimings{AgentRun: 10, CopyRepo: 2, Total: 15},
		EffectiveSpeedup:      2.5,
		ParallelEfficiency:    0.62,
		Model:                 modelName,
		ReasoningEffort:       reasoningEffort,
		CodexVersion:          "test",
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	text := string(encoded)
	for _, want := range []string{`"parallelism":4`, `"cache_mode":"shared"`, `"cache_prewarm_seconds":1.25`, `"harness_elapsed_seconds":12.34`, `"effective_parallel_speedup":2.5`} {
		if !strings.Contains(text, want) {
			t.Fatalf("json = %s, want %s", text, want)
		}
	}

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, value); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	markdown := string(content)
	for _, want := range []string{"Parallelism: `4`", "Cache mode: `shared`", "Cache prewarm seconds: `1.25`", "Harness elapsed seconds: `12.34`", "Effective parallel speedup: `2.50x`", "## Phase Timings", "| agent_run | 10.00 |"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown = %q, want %q", markdown, want)
		}
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
		"OPENHEALTH_DATABASE_PATH=" + filepath.Join(runDir, "repo", "openhealth.db"),
		"GOCACHE=" + filepath.Join(runRoot, "shared-cache", "gocache"),
		"GOMODCACHE=" + filepath.Join(runRoot, "shared-cache", "gomodcache"),
		"TMPDIR=" + filepath.Join(runDir, "tmp"),
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
	for _, want := range []string{"test -run ^$", "./cmd/openhealth-agentops", "./cmd/openhealth", "./agentops"} {
		if !strings.Contains(args, want) {
			t.Fatalf("prewarm args = %q, want %q", args, want)
		}
	}
}

func TestAggregateMetricsRequiresEveryTurnUsage(t *testing.T) {
	t.Parallel()

	first := turnResult{Metrics: testMetrics(1, 50)}
	second := turnResult{Metrics: testMetrics(2, 70)}
	got := aggregateMetrics([]turnResult{first, second})
	if got.ToolCalls != 3 || got.CommandExecutions != 3 || !got.UsageExposed || got.NonCachedInputTokens == nil || *got.NonCachedInputTokens != 120 {
		t.Fatalf("aggregateMetrics = %#v, want summed tools and tokens", got)
	}

	second.Metrics.UsageExposed = false
	got = aggregateMetrics([]turnResult{first, second})
	if got.UsageExposed || got.NonCachedInputTokens != nil {
		t.Fatalf("aggregateMetrics with missing usage = %#v, want aggregate usage hidden", got)
	}
}

func TestWriteMarkdownIncludesTurnDetails(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, report{
		Issue:           issueID,
		Date:            "turn-test",
		Harness:         "test",
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		CodexVersion:    "test",
		Results: []runResult{{
			Variant:  "production",
			Scenario: "mt-weight-clarify-then-add",
			Turns: []turnResult{
				{Index: 1, ExitCode: 0, WallSeconds: 1.2, Metrics: testMetrics(0, 30), RawLogArtifactReference: "<run-root>/production/mt-weight-clarify-then-add/turn-1/events.jsonl"},
				{Index: 2, ExitCode: 0, WallSeconds: 2.3, Metrics: testMetrics(1, 40), RawLogArtifactReference: "<run-root>/production/mt-weight-clarify-then-add/turn-2/events.jsonl"},
			},
		}},
	}); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	markdown := string(content)
	for _, want := range []string{"## Turn Details", "`production/mt-weight-clarify-then-add` turn 1", "turn-2/events.jsonl"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown = %q, want %q", markdown, want)
		}
	}
}

func TestWriteMarkdownIncludesCodeFirstSummary(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, report{
		Issue:           issueID,
		Date:            "code-first-test",
		Harness:         "test",
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		CodexVersion:    "test",
		CodeFirst: &codeFirstSummary{
			CandidateVariant: "production",
			BaselineVariant:  "cli",
			BeatsCLI:         true,
			Recommendation:   "prefer_agentops_production_for_routine_openhealth_operations",
			Criteria: []codeFirstCriterion{
				{Name: "candidate_passes_all_scenarios", Passed: true, Details: "ok"},
			},
			Entries: []codeFirstComparisonRow{
				{Scenario: "add-two", CandidateTools: 1, CLITools: 2, ToolDelta: intPointer(1)},
			},
		},
	}); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(content)
	for _, want := range []string{"## Code-First CLI Comparison", "Candidate: `production`", "Beats CLI: `yes`", "`candidate_passes_all_scenarios`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("markdown = %q, want %q", text, want)
		}
	}
}

func TestShouldSkipEvalPath(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"docs/agent-evals.md",
		"docs/agent-eval-assets",
		"docs/agent-eval-assets/variants/cli/SKILL.md",
		"docs/agent-eval-results",
		"docs/agent-eval-results/oh-5yr-2026-04-16.md",
		"scripts/agent-eval",
		"scripts/agent-eval/oh5yr/main.go",
	} {
		if !shouldSkipEvalPath(path) {
			t.Fatalf("shouldSkipEvalPath(%q) = false, want true", path)
		}
	}

	for _, path := range []string{
		"docs/maintainers.md",
		"scripts/validate-agent-skill.sh",
		"skills/openhealth/SKILL.md",
	} {
		if shouldSkipEvalPath(path) {
			t.Fatalf("shouldSkipEvalPath(%q) = true, want false", path)
		}
	}
}

func TestGeneratedFileInspectionIgnoresBroadListings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		command         string
		output          string
		wantDirect      bool
		wantBroadSearch bool
		wantBroadGen    bool
	}{
		{
			name:            "rg files listing",
			command:         "/bin/zsh -lc rg --files",
			output:          "client/client.gen.go\ninternal/api/generated/server.gen.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "direct rg files listing",
			command:         "rg --files",
			output:          "client/client.gen.go\ninternal/api/generated/server.gen.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "mixed targeted and root rg files listing",
			command:         "rg --files .agents/skills/openhealth repo .",
			output:          ".agents/skills/openhealth/SKILL.md\nclient/client.gen.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "find listing",
			command:         "/bin/zsh -lc find . -type f",
			output:          "./client/client.gen.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "targeted skill file listing",
			command:         "rg --files .agents/skills/openhealth",
			output:          ".agents/skills/openhealth/SKILL.md\n",
			wantDirect:      false,
			wantBroadSearch: false,
			wantBroadGen:    false,
		},
		{
			name:       "direct read",
			command:    "/bin/zsh -lc sed -n '1,40p' client/client.gen.go",
			output:     "package client\n",
			wantDirect: true,
		},
		{
			name:       "skill guidance mentions generated file",
			command:    "/bin/zsh -lc sed -n '1,220p' .agents/skills/openhealth/SKILL.md",
			output:     "Do not inspect client.gen.go\n",
			wantDirect: false,
		},
		{
			name:            "broad content search with generated output",
			command:         "/bin/zsh -lc rg 'CreateHealthWeight' .",
			output:          "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "implicit broad content search with generated output",
			command:         "/bin/zsh -lc rg -n 'CreateHealthWeight'",
			output:          "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:       "targeted content search with generated output",
			command:    "rg 'CreateHealthWeight' client",
			output:     "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			wantDirect: true,
		},
		{
			name:            "direct grep with generated output",
			command:         "grep -R CreateHealthWeight .",
			output:          "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:       "non inspection command",
			command:    "/bin/zsh -lc go test ./...",
			output:     "ok github.com/yazanabuashour/openhealth/client\n",
			wantDirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inspectsGeneratedFileCommand(tt.command, tt.output); got != tt.wantDirect {
				t.Fatalf("inspectsGeneratedFileCommand(%q, %q) = %v, want %v", tt.command, tt.output, got, tt.wantDirect)
			}
			if got := isBroadRepoSearchCommand(tt.command); got != tt.wantBroadSearch {
				t.Fatalf("isBroadRepoSearchCommand(%q) = %v, want %v", tt.command, got, tt.wantBroadSearch)
			}
			gotBroadGen := isBroadRepoSearchCommand(tt.command) && mentionsGeneratedPath(tt.output)
			if gotBroadGen != tt.wantBroadGen {
				t.Fatalf("broad generated path metric = %v, want %v", gotBroadGen, tt.wantBroadGen)
			}
		})
	}
}

func TestCLIAndDirectSQLiteMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		command    string
		wantCLI    bool
		wantSQLite bool
	}{
		{
			name:    "go run cli",
			command: "go run ./cmd/openhealth weight list --limit 2",
			wantCLI: true,
		},
		{
			name:    "go run cli after shell setup",
			command: `/bin/zsh -lc 'cd repo && go run ./cmd/openhealth weight list --limit 2'`,
			wantCLI: true,
		},
		{
			name:    "installed cli",
			command: "/usr/local/bin/openhealth weight add --date 2026-03-29 --value 152.2",
			wantCLI: true,
		},
		{
			name:    "search for go run cli text",
			command: `/bin/zsh -lc 'rg -n "go run ./cmd/openhealth" skills/openhealth'`,
		},
		{
			name:    "grep for installed cli text",
			command: `grep -R "openhealth weight" skills/openhealth`,
		},
		{
			name: "agentops temp runner",
			command: `tmp="$(mktemp -d)" && repo="$(pwd)" && cat > "$tmp/go.mod" <<EOF
require github.com/yazanabuashour/openhealth v0.0.0
replace github.com/yazanabuashour/openhealth => $repo
EOF
(cd "$tmp" && GOPROXY=off GOSUMDB=off go run -mod=mod .)`,
		},
		{
			name:    "agentops json runner",
			command: `go run ./cmd/openhealth-agentops weight <<'EOF'{"action":"list_weights"}EOF`,
		},
		{
			name:       "sqlite executable",
			command:    `sqlite3 "$OPENHEALTH_DATABASE_PATH" "select * from health_weight_entry"`,
			wantSQLite: true,
		},
		{
			name: "python sqlite import",
			command: `python - <<'PY'
import sqlite3
sqlite3.connect("openhealth.db")
PY`,
			wantSQLite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := usesOpenHealthCLI(tt.command); got != tt.wantCLI {
				t.Fatalf("usesOpenHealthCLI(%q) = %v, want %v", tt.command, got, tt.wantCLI)
			}
			if got := usesDirectSQLite(tt.command); got != tt.wantSQLite {
				t.Fatalf("usesDirectSQLite(%q) = %v, want %v", tt.command, got, tt.wantSQLite)
			}
		})
	}
}

func TestNonISODateRejectAssistantPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "strict format rejection",
			message: "Invalid date: use YYYY-MM-DD.",
			want:    true,
		},
		{
			name:    "reject wording",
			message: "I can't record 2026/03/31 because that date format is unsupported.",
			want:    true,
		},
		{
			name:    "bare date mention is not a rejection",
			message: "The date is 2026/03/31.",
			want:    false,
		},
		{
			name:    "successful write wording is not a rejection",
			message: "Recorded 2026/03/31 152.2 lb.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := nonISODateRejectAssistantPass(tt.message); got != tt.want {
				t.Fatalf("nonISODateRejectAssistantPass(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestSelectVariantsAndScenarios(t *testing.T) {
	t.Parallel()

	selectedVariants, err := selectVariants("production,cli")
	if err != nil {
		t.Fatalf("selectVariants: %v", err)
	}
	if got := []string{selectedVariants[0].ID, selectedVariants[1].ID}; strings.Join(got, ",") != "production,cli" {
		t.Fatalf("selected variants = %v", got)
	}
	selectedScenarios, err := selectScenarios("add-two,bounded-range,latest-only")
	if err != nil {
		t.Fatalf("selectScenarios: %v", err)
	}
	if got := []string{selectedScenarios[0].ID, selectedScenarios[1].ID, selectedScenarios[2].ID}; strings.Join(got, ",") != "add-two,bounded-range,latest-only" {
		t.Fatalf("selected scenarios = %v", got)
	}
	if _, err := selectVariants("missing"); err == nil {
		t.Fatal("selectVariants missing id error = nil")
	}
	if _, err := selectScenarios("missing"); err == nil {
		t.Fatal("selectScenarios missing id error = nil")
	}
}

func TestExpandedWeightScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		scenarioID   string
		finalMessage string
		wantWeights  []weightState
	}{
		{
			name:         "latest text",
			scenarioID:   "latest-only",
			finalMessage: "2026-03-30 151.6 lb",
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			},
		},
		{
			name:       "history text",
			scenarioID: "history-limit-two",
			finalMessage: strings.Join([]string{
				"2026-03-30 151.6 lb",
				"2026-03-29 152.2 lb",
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
				{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			},
		},
		{
			name:       "history json lines",
			scenarioID: "history-limit-two",
			finalMessage: strings.Join([]string{
				`{"date":"2026-03-30","value":151.6,"unit":"lb"}`,
				`{"date":"2026-03-29","value":152.2,"unit":"lb"}`,
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
				{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			},
		},
		{
			name:         "non iso reject",
			scenarioID:   "non-iso-date-reject",
			finalMessage: "Invalid date: use YYYY-MM-DD.",
			wantWeights:  []weightState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sc, ok := scenarioByID(tt.scenarioID)
			if !ok {
				t.Fatalf("missing scenario %q", tt.scenarioID)
			}
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			if err := seedScenario(databasePath, sc); err != nil {
				t.Fatalf("seedScenario: %v", err)
			}
			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !weightsEqual(verification.Weights, tt.wantWeights) {
				t.Fatalf("weights = %s, want %s", describeWeights(verification.Weights), describeWeights(tt.wantWeights))
			}
		})
	}
}

func TestBloodPressureScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		finalMessage string
		wantReadings []bloodPressureState
	}{
		{
			scenarioID: "bp-latest-only",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID: "bp-history-limit-two",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
				"2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
				{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			},
		},
		{
			scenarioID: "bp-bounded-range",
			finalMessage: strings.Join([]string{
				"2026-03-30 118/76",
				"2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID: "bp-bounded-range-natural",
			finalMessage: strings.Join([]string{
				"March 30, 2026",
				"- 118/76",
				"",
				"March 29, 2026",
				"- 122/78, pulse 64",
			}, "\n"),
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID:   "bp-invalid-input",
			finalMessage: "Invalid blood pressure: systolic, diastolic, and pulse must be positive.",
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "bp-non-iso-date-reject",
			finalMessage: "Invalid date: use YYYY-MM-DD.",
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "bp-correct-existing",
			finalMessage: "Updated 2026-03-29 to 121/77 pulse 63.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 121, Diastolic: 77, Pulse: intPointer(63)},
			},
		},
		{
			scenarioID:   "bp-correct-missing-reject",
			finalMessage: "No existing reading was found for 2026-03-31, so it was not updated.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
		{
			scenarioID:   "bp-correct-ambiguous-reject",
			finalMessage: "Multiple readings exist for 2026-03-29, so the correction is ambiguous and was not updated.",
			wantReadings: []bloodPressureState{
				{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenarioID, func(t *testing.T) {
			t.Parallel()
			sc, ok := scenarioByID(tt.scenarioID)
			if !ok {
				t.Fatalf("missing scenario %q", tt.scenarioID)
			}
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			if err := seedScenario(databasePath, sc); err != nil {
				t.Fatalf("seedScenario: %v", err)
			}
			if tt.scenarioID == "bp-correct-existing" {
				api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
				if err != nil {
					t.Fatalf("OpenLocal: %v", err)
				}
				readings, err := api.ListBloodPressure(context.Background(), client.BloodPressureListOptions{Limit: 1})
				if err != nil {
					t.Fatalf("ListBloodPressure: %v", err)
				}
				if len(readings) != 1 {
					t.Fatalf("seed readings = %d, want 1", len(readings))
				}
				pulse63 := 63
				if _, err := api.ReplaceBloodPressure(context.Background(), readings[0].ID, client.BloodPressureRecordInput{
					RecordedAt: readings[0].RecordedAt,
					Systolic:   121,
					Diastolic:  77,
					Pulse:      &pulse63,
				}); err != nil {
					t.Fatalf("ReplaceBloodPressure: %v", err)
				}
				_ = api.Close()
			}
			verification, err := verifyScenario(databasePath, sc, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenario: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !bloodPressuresEqual(verification.BloodPressures, tt.wantReadings) {
				t.Fatalf("blood pressures = %s, want %s", describeBloodPressures(verification.BloodPressures), describeBloodPressures(tt.wantReadings))
			}
		})
	}
}

func TestMixedAndMultiTurnScenariosVerifyExpectedOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scenarioID   string
		turnIndex    int
		finalMessage string
		wantWeights  []weightState
		wantReadings []bloodPressureState
		manualSeed   bool
	}{
		{
			scenarioID:   "mixed-add-latest",
			turnIndex:    1,
			finalMessage: "Latest weight: 2026-03-31 150.8 lb. Latest blood pressure: 2026-03-31 119/77 pulse 62.",
			wantWeights:  []weightState{{Date: "2026-03-31", Value: 150.8, Unit: "lb"}},
			wantReadings: []bloodPressureState{{Date: "2026-03-31", Systolic: 119, Diastolic: 77, Pulse: intPointer(62)}},
			manualSeed:   true,
		},
		{
			scenarioID: "mixed-bounded-range",
			turnIndex:  1,
			finalMessage: strings.Join([]string{
				"Weight 2026-03-30 151.6 lb",
				"Weight 2026-03-29 152.2 lb",
				"Blood pressure 2026-03-30 118/76",
				"Blood pressure 2026-03-29 122/78 pulse 64",
			}, "\n"),
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
				{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
				{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			},
		},
		{
			scenarioID:   "mixed-invalid-direct-reject",
			turnIndex:    1,
			finalMessage: "Invalid request: weight unit stone is unsupported and blood-pressure values must be positive.",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "mt-weight-clarify-then-add",
			turnIndex:    1,
			finalMessage: "Which year should I use for 03/29?",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{},
		},
		{
			scenarioID:   "mt-mixed-latest-then-correct",
			turnIndex:    2,
			finalMessage: "Updated 2026-03-30: weight 151.0 lb and blood pressure 117/75 pulse 63.",
			wantWeights: []weightState{
				{Date: "2026-03-30", Value: 151.0, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
			manualSeed: true,
		},
		{
			scenarioID:   "mt-bp-latest-then-correct",
			turnIndex:    1,
			finalMessage: "2026-03-30 118/76",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
		},
		{
			scenarioID:   "mt-bp-latest-then-correct",
			turnIndex:    2,
			finalMessage: "Updated 2026-03-30 to 117/75 pulse 63.",
			wantWeights:  []weightState{},
			wantReadings: []bloodPressureState{
				{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			},
			manualSeed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenarioID, func(t *testing.T) {
			t.Parallel()
			sc, ok := scenarioByID(tt.scenarioID)
			if !ok {
				t.Fatalf("missing scenario %q", tt.scenarioID)
			}
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			if !tt.manualSeed {
				if err := seedScenario(databasePath, sc); err != nil {
					t.Fatalf("seedScenario: %v", err)
				}
			}
			if tt.manualSeed {
				api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
				if err != nil {
					t.Fatalf("OpenLocal: %v", err)
				}
				if err := upsertWeights(context.Background(), api, tt.wantWeights); err != nil {
					t.Fatalf("upsertWeights: %v", err)
				}
				if err := recordBloodPressures(context.Background(), api, tt.wantReadings); err != nil {
					t.Fatalf("recordBloodPressures: %v", err)
				}
				_ = api.Close()
			}

			verification, err := verifyScenarioTurn(databasePath, sc, tt.turnIndex, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenarioTurn: %v", err)
			}
			if !verification.Passed {
				t.Fatalf("verification failed: %#v", verification)
			}
			if !weightsEqual(verification.Weights, tt.wantWeights) {
				t.Fatalf("weights = %s, want %s", describeWeights(verification.Weights), describeWeights(tt.wantWeights))
			}
			if !bloodPressuresEqual(verification.BloodPressures, tt.wantReadings) {
				t.Fatalf("blood pressures = %s, want %s", describeBloodPressures(verification.BloodPressures), describeBloodPressures(tt.wantReadings))
			}
		})
	}
}

func TestWeightOnlyMultiTurnRejectsBloodPressureWrites(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("mt-weight-clarify-then-add")
	if !ok {
		t.Fatal("missing mt-weight-clarify-then-add scenario")
	}
	tests := []struct {
		name         string
		turnIndex    int
		finalMessage string
		weights      []weightState
		readings     []bloodPressureState
	}{
		{
			name:         "turn one",
			turnIndex:    1,
			finalMessage: "Which year should I use for 03/29?",
			readings:     []bloodPressureState{{Date: "2026-03-29", Systolic: 122, Diastolic: 78}},
		},
		{
			name:         "turn two",
			turnIndex:    2,
			finalMessage: "Stored 2026-03-29 152.2 lb.",
			weights:      []weightState{{Date: "2026-03-29", Value: 152.2, Unit: "lb"}},
			readings:     []bloodPressureState{{Date: "2026-03-29", Systolic: 122, Diastolic: 78}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			databasePath := filepath.Join(t.TempDir(), "openhealth.db")
			api, err := client.OpenLocal(client.LocalConfig{DatabasePath: databasePath})
			if err != nil {
				t.Fatalf("OpenLocal: %v", err)
			}
			if len(tt.weights) > 0 {
				if err := upsertWeights(context.Background(), api, tt.weights); err != nil {
					t.Fatalf("upsertWeights: %v", err)
				}
			}
			if err := recordBloodPressures(context.Background(), api, tt.readings); err != nil {
				t.Fatalf("recordBloodPressures: %v", err)
			}
			_ = api.Close()

			verification, err := verifyScenarioTurn(databasePath, sc, tt.turnIndex, tt.finalMessage)
			if err != nil {
				t.Fatalf("verifyScenarioTurn: %v", err)
			}
			if verification.DatabasePass || verification.Passed {
				t.Fatalf("verification = %#v, want database failure for stray blood-pressure write", verification)
			}
		})
	}
}

func TestMentionsDatesInOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "newest first iso",
			message: "Stored rows: 2026-03-30 151.6 lb, 2026-03-29 152.2 lb.",
			want:    true,
		},
		{
			name:    "oldest first iso",
			message: "Stored rows: 2026-03-29 152.2 lb, 2026-03-30 151.6 lb.",
			want:    false,
		},
		{
			name:    "newest first short dates",
			message: "03/30: 151.6 lb; 03/29: 152.2 lb.",
			want:    true,
		},
		{
			name:    "missing date",
			message: "Only 2026-03-30 is present.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mentionsDatesInOrder(tt.message, "2026-03-30", "2026-03-29")
			if got != tt.want {
				t.Fatalf("mentionsDatesInOrder(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestBoundedRangeAssistantPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name: "allows excluded date in exclusion sentence",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"No entries from 03/28/2026 are included.",
			}, "\n"),
			want: true,
		},
		{
			name: "ignores natural date order in prose before result rows",
			message: strings.Join([]string{
				"Here are the weights for March 29 and March 30, 2026:",
				"",
				"- 2026-03-30 12:00:00Z: 151.6 lb",
				"- 2026-03-29 12:00:00Z: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name: "allows bounded prose with bullet result rows",
			message: strings.Join([]string{
				"Using the configured local database, the weights for March 29 and March 30, 2026, newest first:",
				"",
				"- 2026-03-30: 151.6 lb",
				"- 2026-03-29: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name:    "allows compact same-line result rows",
			message: "2026-03-30: 151.6 lb; 2026-03-29: 152.2 lb",
			want:    true,
		},
		{
			name: "rejects excluded date as result row",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"3. 2026-03-28 153.0 lb",
			}, "\n"),
			want: false,
		},
		{
			name:    "rejects missing newest date",
			message: "1. 2026-03-29 152.2 lb",
			want:    false,
		},
		{
			name: "rejects oldest first",
			message: strings.Join([]string{
				"1. 2026-03-29 152.2 lb",
				"2. 2026-03-30 151.6 lb",
			}, "\n"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := boundedRangeAssistantPass(tt.message); got != tt.want {
				t.Fatalf("boundedRangeAssistantPass(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestMixedBoundedRangeAssistantPassAllowsJSONAnswer(t *testing.T) {
	t.Parallel()

	message := `{"weights":[{"date":"2026-03-30","value":151.6,"unit":"lb"},{"date":"2026-03-29","value":152.2,"unit":"lb"}],"blood_pressure":[{"date":"2026-03-30","systolic":118,"diastolic":76},{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}`
	if !mixedBoundedRangeAssistantPass(message) {
		t.Fatalf("mixedBoundedRangeAssistantPass rejected JSON answer: %s", message)
	}
}

func TestNaturalBoundedRangeScenarioSeedsExpectedRows(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("bounded-range-natural")
	if !ok {
		t.Fatal("missing bounded-range-natural scenario")
	}
	if !strings.Contains(sc.Prompt, "Mar 29") || !strings.Contains(sc.Prompt, "Mar 30") {
		t.Fatalf("prompt = %q, want natural Mar 29 and Mar 30 wording", sc.Prompt)
	}

	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	if err := seedScenario(databasePath, sc); err != nil {
		t.Fatalf("seedScenario: %v", err)
	}
	weights, err := listWeights(databasePath)
	if err != nil {
		t.Fatalf("listWeights: %v", err)
	}
	got := weightStates(weights)
	want := []weightState{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
	}
	if !weightsEqual(got, want) {
		t.Fatalf("seeded weights = %s, want %s", describeWeights(got), describeWeights(want))
	}
}

func TestCompareReportsCapturesDeltas(t *testing.T) {
	t.Parallel()

	nonCachedBefore := 100
	nonCachedAfter := 80
	baseline := report{Results: []runResult{{
		Variant:     "production",
		Scenario:    "bounded-range",
		Passed:      false,
		WallSeconds: 40,
		Metrics: metrics{
			ToolCalls:              10,
			AssistantCalls:         4,
			NonCachedInputTokens:   &nonCachedBefore,
			GeneratedFileInspected: true,
		},
	}}}
	current := report{Results: []runResult{{
		Variant:     "production",
		Scenario:    "bounded-range",
		Passed:      true,
		WallSeconds: 35.5,
		Metrics: metrics{
			ToolCalls:              7,
			AssistantCalls:         5,
			NonCachedInputTokens:   &nonCachedAfter,
			GeneratedFileInspected: false,
		},
	}}}

	comparison := compareReports(baseline, current, "docs/agent-eval-results/baseline.json")
	if comparison.BaselineReport != "docs/agent-eval-results/baseline.json" {
		t.Fatalf("baseline = %q", comparison.BaselineReport)
	}
	if len(comparison.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(comparison.Entries))
	}
	entry := comparison.Entries[0]
	if entry.Result != "fixed" {
		t.Fatalf("result = %q, want fixed", entry.Result)
	}
	if entry.ToolCallsDelta == nil || *entry.ToolCallsDelta != -3 {
		t.Fatalf("tool delta = %v, want -3", entry.ToolCallsDelta)
	}
	if entry.AssistantCallsDelta == nil || *entry.AssistantCallsDelta != 1 {
		t.Fatalf("assistant delta = %v, want 1", entry.AssistantCallsDelta)
	}
	if entry.WallSecondsDelta == nil || *entry.WallSecondsDelta != -4.5 {
		t.Fatalf("wall delta = %v, want -4.5", entry.WallSecondsDelta)
	}
	if entry.NonCachedInputTokensDelta == nil || *entry.NonCachedInputTokensDelta != -20 {
		t.Fatalf("token delta = %v, want -20", entry.NonCachedInputTokensDelta)
	}
	if entry.GeneratedFileInspectionChange != "improved_to_no" {
		t.Fatalf("generated change = %q, want improved_to_no", entry.GeneratedFileInspectionChange)
	}
}

func TestMetricNotesUseReportDate(t *testing.T) {
	t.Parallel()

	notes := metricNotes("2026-05-01", []runResult{
		{
			Variant:  "production",
			Scenario: "add-two",
			Metrics: metrics{
				GeneratedFileInspected: true,
			},
		},
		{
			Variant:  "production",
			Scenario: "bounded-range",
			Metrics: metrics{
				ModuleCacheInspected: true,
			},
		},
	})
	if len(notes) != 2 {
		t.Fatalf("notes = %d, want 2", len(notes))
	}
	for _, note := range notes {
		if !strings.Contains(note, "2026-05-01") {
			t.Fatalf("note = %q, want report date", note)
		}
		if strings.Contains(note, "2026-04-17") {
			t.Fatalf("note = %q, should not contain hard-coded date", note)
		}
	}
}

func TestProductionStopLossTriggersPivot(t *testing.T) {
	t.Parallel()

	summary := productionStopLoss([]runResult{
		{
			Variant:  "production",
			Scenario: "add-two",
			Passed:   true,
			Verification: verificationResult{
				DatabasePass:  true,
				AssistantPass: true,
			},
			Metrics: metrics{
				ToolCalls:              16,
				GeneratedFileInspected: true,
			},
		},
		{
			Variant:  "production",
			Scenario: "bounded-range",
			Passed:   true,
			Verification: verificationResult{
				DatabasePass:  true,
				AssistantPass: true,
			},
			Metrics: metrics{
				ToolCalls:       5,
				BroadRepoSearch: true,
			},
		},
		{
			Variant:  "production",
			Scenario: "bounded-range-natural",
			Passed:   true,
			Verification: verificationResult{
				DatabasePass:  true,
				AssistantPass: true,
			},
			Metrics: metrics{
				ToolCalls:       7,
				BroadRepoSearch: true,
			},
		},
	})
	if summary == nil {
		t.Fatal("productionStopLoss returned nil")
	}
	if !summary.Triggered {
		t.Fatal("stop loss did not trigger")
	}
	if summary.Recommendation != "continue_cli_baseline_for_agent_operations" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
	joined := strings.Join(summary.Triggers, "\n")
	for _, want := range []string{"direct generated-file inspection", "broad repo search", "above threshold"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("triggers = %q, want %q", joined, want)
		}
	}
}

func TestProductionStopLossIgnoresValidationBroadSearchForRoutineThreshold(t *testing.T) {
	t.Parallel()

	summary := productionStopLoss([]runResult{
		{
			Variant:  "production",
			Scenario: "bounded-range",
			Passed:   true,
			Verification: verificationResult{
				DatabasePass:  true,
				AssistantPass: true,
			},
			Metrics: metrics{
				ToolCalls:       5,
				BroadRepoSearch: true,
			},
		},
		{
			Variant:  "production",
			Scenario: "invalid-input",
			Passed:   true,
			Verification: verificationResult{
				DatabasePass:  true,
				AssistantPass: true,
			},
			Metrics: metrics{
				ToolCalls:       1,
				BroadRepoSearch: true,
			},
		},
	})
	if summary == nil {
		t.Fatal("productionStopLoss returned nil")
	}
	if summary.Triggered {
		t.Fatalf("stop loss triggered from validation broad search: %v", summary.Triggers)
	}
	if summary.Recommendation != "continue_production_hardening" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
}

func TestCodeFirstSummaryBeatsCLI(t *testing.T) {
	t.Parallel()

	results := []runResult{}
	for _, sc := range scenarios() {
		results = append(results,
			runResult{
				Variant:  "production",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  codeFirstProductionMetrics(sc.ID),
			},
			runResult{
				Variant:  "cli",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  testMetrics(2, 100),
			},
		)
	}

	summary := codeFirstSummaryFor(results, "production")
	if summary == nil {
		t.Fatal("codeFirstSummaryFor returned nil")
	}
	if !summary.BeatsCLI {
		t.Fatalf("beats_cli = false, criteria = %#v", summary.Criteria)
	}
	if summary.Recommendation != "prefer_agentops_production_for_routine_openhealth_operations" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
}

func TestCodeFirstSummarySupportsProductionCandidate(t *testing.T) {
	t.Parallel()

	results := []runResult{}
	for _, sc := range scenarios() {
		results = append(results,
			runResult{
				Variant:  "production",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  codeFirstProductionMetrics(sc.ID),
			},
			runResult{
				Variant:  "cli",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  testMetrics(2, 100),
			},
		)
	}

	summary := codeFirstSummaryFor(results, "production")
	if summary == nil {
		t.Fatal("codeFirstSummaryFor returned nil")
	}
	if !summary.BeatsCLI {
		t.Fatalf("beats_cli = false, criteria = %#v", summary.Criteria)
	}
	if summary.CandidateVariant != "production" {
		t.Fatalf("candidate = %q, want production", summary.CandidateVariant)
	}
	if summary.Recommendation != "prefer_agentops_production_for_routine_openhealth_operations" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
}

func TestCodeFirstSummaryFailsWhenRoutineScenarioExceedsCLI(t *testing.T) {
	t.Parallel()

	results := []runResult{}
	for _, sc := range scenarios() {
		candidateTools := 1
		cliTools := 1
		if sc.ID == "update-existing" {
			candidateTools = 4
		}
		candidateMetrics := testMetrics(candidateTools, 50)
		if isFinalAnswerOnlyValidationScenario(sc.ID) {
			candidateMetrics = testMetrics(0, 50)
		}
		results = append(results,
			runResult{
				Variant:  "production",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  candidateMetrics,
			},
			runResult{
				Variant:  "cli",
				Scenario: sc.ID,
				Passed:   true,
				Metrics:  testMetrics(cliTools, 100),
			},
		)
	}

	summary := codeFirstSummaryFor(results, "production")
	if summary == nil {
		t.Fatal("codeFirstSummaryFor returned nil")
	}
	if summary.BeatsCLI {
		t.Fatalf("beats_cli = true, want false")
	}
	joined := ""
	for _, criterion := range summary.Criteria {
		joined += criterion.Name + ":" + criterion.Details + "\n"
	}
	if !strings.Contains(joined, "update-existing") {
		t.Fatalf("criteria = %q, want update-existing detail", joined)
	}
}

func TestCodeFirstSummaryFailsInvalidScenariosThatUseTools(t *testing.T) {
	t.Parallel()

	results := []runResult{}
	for _, sc := range scenarios() {
		productionMetrics := codeFirstProductionMetrics(sc.ID)
		if sc.ID == "invalid-input" {
			productionMetrics = testMetrics(1, 50)
		}
		results = append(results,
			runResult{Variant: "production", Scenario: sc.ID, Passed: true, Metrics: productionMetrics},
			runResult{Variant: "cli", Scenario: sc.ID, Passed: true, Metrics: testMetrics(2, 100)},
		)
	}

	summary := codeFirstSummaryFor(results, "production")
	if summary == nil {
		t.Fatal("codeFirstSummaryFor returned nil")
	}
	criterion := criterionByName(summary.Criteria, "validation_scenarios_are_final_answer_only")
	if criterion.Passed || !strings.Contains(criterion.Details, "invalid-input") {
		t.Fatalf("validation criterion = %#v, want invalid-input failure", criterion)
	}
}

func TestCodeFirstSummaryFailsWhenNonCachedTokenMajorityLoses(t *testing.T) {
	t.Parallel()

	results := []runResult{}
	for index, sc := range scenarios() {
		productionTokens := 120
		if index == 0 {
			productionTokens = 50
		}
		productionMetrics := testMetrics(1, productionTokens)
		if isFinalAnswerOnlyValidationScenario(sc.ID) {
			productionMetrics = testMetrics(0, productionTokens)
		}
		results = append(results,
			runResult{Variant: "production", Scenario: sc.ID, Passed: true, Metrics: productionMetrics},
			runResult{Variant: "cli", Scenario: sc.ID, Passed: true, Metrics: testMetrics(2, 100)},
		)
	}

	summary := codeFirstSummaryFor(results, "production")
	if summary == nil {
		t.Fatal("codeFirstSummaryFor returned nil")
	}
	criterion := criterionByName(summary.Criteria, "non_cached_token_majority")
	if criterion.Passed || !strings.Contains(criterion.Details, "required") {
		t.Fatalf("token majority criterion = %#v, want failure", criterion)
	}
}

func criterionByName(criteria []codeFirstCriterion, name string) codeFirstCriterion {
	for _, criterion := range criteria {
		if criterion.Name == name {
			return criterion
		}
	}
	return codeFirstCriterion{}
}
