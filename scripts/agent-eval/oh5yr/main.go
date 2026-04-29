package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	issueID               = "oh-5yr"
	modelName             = "gpt-5.4-mini"
	reasoningEffort       = "medium"
	defaultRunParallelism = 4
	cacheModeShared       = "shared"
	cacheModeIsolated     = "isolated"
)

var prewarmCompilePackages = []string{"./cmd/openhealth", "./internal/runner"}

func main() {
	if len(os.Args) < 2 {
		failf("usage: oh5yr <run|seed|verify>")
	}

	switch os.Args[1] {
	case "run":
		runCommand(os.Args[2:])
	case "seed":
		seedCommand(os.Args[2:])
	case "verify":
		verifyCommand(os.Args[2:])
	default:
		failf("unknown command %q", os.Args[1])
	}
}

func parseRunOptions(args []string) (runOptions, error) {
	options := runOptions{
		Date:        time.Now().Format(time.DateOnly),
		Parallelism: defaultRunParallelism,
		CacheMode:   cacheModeShared,
	}
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&options.RunRoot, "run-root", options.RunRoot, "directory for raw run artifacts outside the repo")
	fs.StringVar(&options.Date, "date", options.Date, "report date in YYYY-MM-DD form")
	fs.StringVar(&options.CompareTo, "compare-to", options.CompareTo, "optional baseline JSON report path for comparison")
	fs.StringVar(&options.VariantFilter, "variant", options.VariantFilter, "optional comma-separated variant ids to run")
	fs.StringVar(&options.ScenarioFilter, "scenario", options.ScenarioFilter, "optional comma-separated scenario ids to run")
	fs.IntVar(&options.Parallelism, "parallel", options.Parallelism, "number of scenario jobs to run concurrently")
	fs.StringVar(&options.CacheMode, "cache-mode", options.CacheMode, "Go cache mode: shared or isolated")
	if err := fs.Parse(args); err != nil {
		return runOptions{}, err
	}
	if fs.NArg() != 0 {
		return runOptions{}, errors.New("run does not accept positional arguments")
	}
	if options.Parallelism < 1 {
		return runOptions{}, errors.New("parallel must be greater than or equal to 1")
	}
	switch options.CacheMode {
	case cacheModeShared, cacheModeIsolated:
	default:
		return runOptions{}, fmt.Errorf("cache-mode must be %q or %q", cacheModeShared, cacheModeIsolated)
	}
	return options, nil
}

func runCommand(args []string) {
	options, err := parseRunOptions(args)
	if err != nil {
		failf("parse flags: %v", err)
	}

	repoRoot, err := repoRoot()
	if err != nil {
		failf("resolve repo root: %v", err)
	}

	runRoot := options.RunRoot
	if runRoot == "" {
		runRoot, err = os.MkdirTemp("", "openhealth-oh-5yr-*")
		if err != nil {
			failf("create run root: %v", err)
		}
	} else if err := os.MkdirAll(runRoot, 0o755); err != nil {
		failf("create run root %s: %v", runRoot, err)
	}
	runRoot, err = filepath.Abs(runRoot)
	if err != nil {
		failf("absolute run root: %v", err)
	}
	if isWithin(runRoot, repoRoot) {
		failf("run root must be outside the repository: %s", runRoot)
	}
	selectedVariants, err := selectVariants(options.VariantFilter)
	if err != nil {
		failf("select variants: %v", err)
	}
	selectedScenarios, err := selectScenarios(options.ScenarioFilter)
	if err != nil {
		failf("select scenarios: %v", err)
	}
	cacheConfig := cacheConfig{
		Mode:    options.CacheMode,
		RunRoot: runRoot,
	}
	if err := setupEvalCodexHome(runRoot); err != nil {
		failf("prepare eval Codex home: %v", err)
	}

	marker := filepath.Join(runRoot, "history-marker")
	if err := os.WriteFile(marker, []byte(time.Now().Format(time.RFC3339Nano)), 0o644); err != nil {
		failf("write history marker: %v", err)
	}
	markerInfo, err := os.Stat(marker)
	if err != nil {
		failf("stat history marker: %v", err)
	}

	codexVersion := commandOutputWithEnv(evalCodexEnv(runRoot), "codex", "--version")
	cachePrewarmSeconds := 0.0
	if options.CacheMode == cacheModeShared {
		start := time.Now()
		if err := prewarmSharedCache(repoRoot, cacheConfig); err != nil {
			failf("prewarm shared Go cache: %v", err)
		}
		cachePrewarmSeconds = roundSeconds(time.Since(start).Seconds())
	}
	harnessStart := time.Now()
	jobs := evalJobsFor(selectedVariants, selectedScenarios)
	results := runEvalJobs(repoRoot, runRoot, jobs, options.Parallelism, cacheConfig, runOne)
	harnessElapsedSeconds := roundSeconds(time.Since(harnessStart).Seconds())
	phaseTotals := aggregatePhaseTimings(results)
	effectiveSpeedup := 0.0
	parallelEfficiency := 0.0
	if harnessElapsedSeconds > 0 {
		effectiveSpeedup = roundSeconds(totalAgentWallSeconds(results) / harnessElapsedSeconds)
		if options.Parallelism > 0 {
			parallelEfficiency = roundSeconds(effectiveSpeedup / float64(options.Parallelism))
		}
	}

	newSessionFiles := countNewSessionFiles(markerInfo.ModTime(), runRoot)
	multiTurnJobs := countMultiTurnJobs(jobs)
	historyStatus, limitation := historyIsolationStatus(newSessionFiles, multiTurnJobs)

	outDir := filepath.Join(repoRoot, "docs", "agent-eval-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		failf("create result directory: %v", err)
	}
	jsonPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.json", issueID, options.Date))
	mdPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.md", issueID, options.Date))

	outReport := report{
		Issue:                 issueID,
		Date:                  options.Date,
		Model:                 modelName,
		ReasoningEffort:       reasoningEffort,
		Harness:               "codex exec --json --full-auto from throwaway run directories with isolated CODEX_HOME; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots",
		Parallelism:           options.Parallelism,
		CacheMode:             options.CacheMode,
		CachePrewarmSeconds:   cachePrewarmSeconds,
		HarnessElapsedSeconds: harnessElapsedSeconds,
		PhaseTotals:           phaseTotals,
		EffectiveSpeedup:      effectiveSpeedup,
		ParallelEfficiency:    parallelEfficiency,
		CodexVersion:          codexVersion,
		HistoryIsolation: historyIsolationSummary{
			Status:                     historyStatus,
			EphemeralFlagRequired:      true,
			RunDirectoryOutsideRepo:    true,
			NewSessionFilesAfterRun:    newSessionFiles,
			SingleTurnEphemeralRuns:    countSingleTurnJobs(jobs),
			MultiTurnPersistedSessions: countMultiTurnJobs(jobs),
			MultiTurnPersistedTurns:    countMultiTurnPersistedTurns(jobs),
			OpenHealthWorkspaceUsed:    false,
			DesktopAppUsed:             false,
			VerificationMethod:         "Single-turn scenarios use codex exec --ephemeral from <run-root>/.../repo. Multi-turn scenarios create one persisted Codex exec session per variant/scenario inside <run-root>/codex-home and resume it for later turns; all raw logs stay under <run-root>.",
			VerificationLimitation:     limitation,
		},
		CommandTemplate: []string{
			"OPENHEALTH_DATABASE_PATH=<run-root>/<variant>/<scenario>/repo/openhealth.db",
			"CODEX_HOME=<run-root>/codex-home, seeded with auth.json only",
			"PATH=<run-root>/<variant>/<scenario>/bin:$PATH",
			"GOCACHE=<run-root>/shared-cache/gocache when --cache-mode shared; otherwise <run-root>/<variant>/<scenario>/gocache",
			"GOMODCACHE=<run-root>/shared-cache/gomodcache when --cache-mode shared; otherwise <run-root>/<variant>/<scenario>/gomodcache",
			"single turn: codex exec --json --ephemeral --full-auto --skip-git-repo-check --ignore-user-config --add-dir <run-root>/<variant>/<scenario> --add-dir <run-root>/shared-cache when --cache-mode shared -C <run-root>/<variant>/<scenario>/repo -m gpt-5.4-mini -c model_reasoning_effort=\"medium\" -c shell_environment_policy.inherit=all <natural user prompt>",
			"multi turn: first turn uses codex exec without --ephemeral and with --ignore-user-config; later turns use codex exec -C <run-root>/<variant>/<scenario>/repo --add-dir <writable-eval-roots> resume --ignore-user-config <thread-id> --json with per-turn logs",
		},
		MetricNotes:       metricNotes(options.Date, results),
		StopLoss:          productionStopLoss(results),
		Results:           results,
		RawLogsCommitted:  false,
		RawLogsNote:       "Raw codex exec event logs and stderr files were retained under <run-root> during execution and intentionally not committed.",
		TokenUsageCaveat:  "Token metrics come from codex exec turn.completed usage events when exposed; unavailable usage must be recorded as not_exposed.",
		AppServerFallback: "not used: codex exec --json exposed enough event detail for this run",
	}
	baseline, baselineRef, err := baselineReport(repoRoot, outDir, jsonPath, options.CompareTo)
	if err != nil {
		failf("read baseline report: %v", err)
	}
	if baseline != nil {
		outReport.Comparison = compareReports(*baseline, outReport, baselineRef)
	}
	if err := writeJSON(jsonPath, outReport); err != nil {
		failf("write JSON report: %v", err)
	}
	if err := writeMarkdown(mdPath, outReport); err != nil {
		failf("write Markdown report: %v", err)
	}

	fmt.Printf("run_root=%s\n", runRoot)
	fmt.Printf("json_report=%s\n", jsonPath)
	fmt.Printf("markdown_report=%s\n", mdPath)
}

func seedCommand(args []string) {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	dbPath := fs.String("db", "", "SQLite database path")
	scenarioID := fs.String("scenario", "", "scenario id")
	if err := fs.Parse(args); err != nil {
		failf("parse flags: %v", err)
	}
	if *dbPath == "" || *scenarioID == "" {
		failf("seed requires --db and --scenario")
	}
	sc, ok := scenarioByID(*scenarioID)
	if !ok {
		failf("unknown scenario %q", *scenarioID)
	}
	if err := seedScenario(*dbPath, sc); err != nil {
		failf("seed scenario: %v", err)
	}
}

func verifyCommand(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	dbPath := fs.String("db", "", "SQLite database path")
	scenarioID := fs.String("scenario", "", "scenario id")
	eventsPath := fs.String("events-jsonl", "", "codex exec JSONL event log")
	if err := fs.Parse(args); err != nil {
		failf("parse flags: %v", err)
	}
	if *dbPath == "" || *scenarioID == "" {
		failf("verify requires --db and --scenario")
	}

	sc, ok := scenarioByID(*scenarioID)
	if !ok {
		failf("unknown scenario %q", *scenarioID)
	}
	finalMessage := ""
	if *eventsPath != "" {
		parsed, err := parseMetrics(*eventsPath)
		if err != nil {
			failf("parse events: %v", err)
		}
		finalMessage = parsed.finalMessage
	}

	result, err := verifyScenario(*dbPath, sc, finalMessage)
	if err != nil {
		failf("verify scenario: %v", err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		failf("encode verification result: %v", err)
	}
	if !result.Passed {
		os.Exit(1)
	}
}
