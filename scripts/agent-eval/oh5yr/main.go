package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yazanabuashour/openhealth/client"
	storagesqlite "github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

const (
	issueID               = "oh-5yr"
	modelName             = "gpt-5.4-mini"
	reasoningEffort       = "medium"
	defaultRunParallelism = 4
	cacheModeShared       = "shared"
	cacheModeIsolated     = "isolated"
)

var prewarmCompilePackages = []string{"./cmd/openhealth-agentops", "./cmd/openhealth", "./agentops"}

type scenario struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prompt string         `json:"prompt,omitempty"`
	Turns  []scenarioTurn `json:"turns,omitempty"`
}

type scenarioTurn struct {
	Prompt string `json:"prompt"`
}

type variant struct {
	ID    string
	Title string
}

type report struct {
	Issue                 string                  `json:"issue"`
	Date                  string                  `json:"date"`
	Model                 string                  `json:"model"`
	ReasoningEffort       string                  `json:"reasoning_effort"`
	Harness               string                  `json:"harness"`
	Parallelism           int                     `json:"parallelism"`
	CacheMode             string                  `json:"cache_mode"`
	CachePrewarmSeconds   float64                 `json:"cache_prewarm_seconds,omitempty"`
	HarnessElapsedSeconds float64                 `json:"harness_elapsed_seconds"`
	PhaseTotals           phaseTimings            `json:"phase_totals"`
	EffectiveSpeedup      float64                 `json:"effective_parallel_speedup,omitempty"`
	ParallelEfficiency    float64                 `json:"parallel_efficiency,omitempty"`
	CodexVersion          string                  `json:"codex_version"`
	HistoryIsolation      historyIsolationSummary `json:"history_isolation"`
	CommandTemplate       []string                `json:"command_template"`
	CLIStatus             string                  `json:"cli_status"`
	MetricNotes           []string                `json:"metric_notes,omitempty"`
	StopLoss              *stopLossSummary        `json:"stop_loss,omitempty"`
	CodeFirst             *codeFirstSummary       `json:"code_first,omitempty"`
	Results               []runResult             `json:"results"`
	Comparison            *comparisonSummary      `json:"comparison,omitempty"`
	RawLogsCommitted      bool                    `json:"raw_logs_committed"`
	RawLogsNote           string                  `json:"raw_logs_note"`
	TokenUsageCaveat      string                  `json:"token_usage_caveat"`
	AppServerFallback     string                  `json:"app_server_fallback"`
}

type historyIsolationSummary struct {
	Status                     string `json:"status"`
	EphemeralFlagRequired      bool   `json:"ephemeral_flag_required"`
	RunDirectoryOutsideRepo    bool   `json:"run_directory_outside_repo"`
	NewSessionFilesAfterRun    int    `json:"new_session_files_after_run"`
	SingleTurnEphemeralRuns    int    `json:"single_turn_ephemeral_runs"`
	MultiTurnPersistedSessions int    `json:"multi_turn_persisted_sessions"`
	MultiTurnPersistedTurns    int    `json:"multi_turn_persisted_turns"`
	OpenHealthWorkspaceUsed    bool   `json:"openhealth_workspace_used"`
	DesktopAppUsed             bool   `json:"desktop_app_used"`
	VerificationMethod         string `json:"verification_method"`
	VerificationLimitation     string `json:"verification_limitation,omitempty"`
}

type runResult struct {
	Variant                 string             `json:"variant"`
	Scenario                string             `json:"scenario"`
	ScenarioTitle           string             `json:"scenario_title"`
	Passed                  bool               `json:"passed"`
	ExitCode                int                `json:"exit_code"`
	WallSeconds             float64            `json:"wall_seconds"`
	PhaseTimings            phaseTimings       `json:"phase_timings"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	Turns                   []turnResult       `json:"turns,omitempty"`
	PromptSummary           string             `json:"prompt_summary"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
}

type turnResult struct {
	Index                   int                `json:"turn_index"`
	WallSeconds             float64            `json:"wall_seconds"`
	ExitCode                int                `json:"exit_code"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
}

type phaseTimings struct {
	PrepareRunDir  float64 `json:"prepare_run_dir_seconds,omitempty"`
	CopyRepo       float64 `json:"copy_repo_seconds,omitempty"`
	InstallVariant float64 `json:"install_variant_seconds,omitempty"`
	WarmCache      float64 `json:"warm_cache_seconds,omitempty"`
	SeedDB         float64 `json:"seed_db_seconds,omitempty"`
	AgentRun       float64 `json:"agent_run_seconds,omitempty"`
	ParseMetrics   float64 `json:"parse_metrics_seconds,omitempty"`
	Verify         float64 `json:"verify_seconds,omitempty"`
	Total          float64 `json:"total_seconds,omitempty"`
}

type metrics struct {
	AssistantCalls                       int            `json:"assistant_calls"`
	ToolCalls                            int            `json:"tool_calls"`
	CommandExecutions                    int            `json:"command_executions"`
	FileInspectionCommands               int            `json:"file_inspection_commands"`
	GeneratedFileInspected               bool           `json:"generated_file_inspected"`
	GeneratedPathFromBroadSearch         bool           `json:"generated_path_from_broad_search"`
	BroadRepoSearch                      bool           `json:"broad_repo_search"`
	ModuleCacheInspected                 bool           `json:"module_cache_inspected"`
	CLIUsed                              bool           `json:"cli_used"`
	DirectSQLiteAccess                   bool           `json:"direct_sqlite_access"`
	GeneratedFileEvidence                []string       `json:"generated_file_evidence,omitempty"`
	GeneratedPathFromBroadSearchEvidence []string       `json:"generated_path_from_broad_search_evidence,omitempty"`
	BroadRepoSearchEvidence              []string       `json:"broad_repo_search_evidence,omitempty"`
	ModuleCacheEvidence                  []string       `json:"module_cache_evidence,omitempty"`
	CLIUsageEvidence                     []string       `json:"cli_usage_evidence,omitempty"`
	DirectSQLiteEvidence                 []string       `json:"direct_sqlite_evidence,omitempty"`
	UsageExposed                         bool           `json:"usage_exposed"`
	InputTokens                          *int           `json:"input_tokens,omitempty"`
	CachedInputTokens                    *int           `json:"cached_input_tokens,omitempty"`
	NonCachedInputTokens                 *int           `json:"non_cached_input_tokens,omitempty"`
	OutputTokens                         *int           `json:"output_tokens,omitempty"`
	EventTypeCounts                      map[string]int `json:"event_type_counts"`
	CommandMetricLimitations             string         `json:"command_metric_limitations"`
}

type verificationResult struct {
	Passed         bool                 `json:"passed"`
	DatabasePass   bool                 `json:"database_pass"`
	AssistantPass  bool                 `json:"assistant_pass"`
	Details        string               `json:"details"`
	Weights        []weightState        `json:"weights,omitempty"`
	BloodPressures []bloodPressureState `json:"blood_pressures,omitempty"`
}

type comparisonSummary struct {
	BaselineReport string            `json:"baseline_report"`
	Entries        []comparisonEntry `json:"entries"`
}

type stopLossSummary struct {
	Policy         string   `json:"policy"`
	Triggered      bool     `json:"triggered"`
	Recommendation string   `json:"recommendation"`
	Triggers       []string `json:"triggers,omitempty"`
}

type codeFirstSummary struct {
	CandidateVariant string                   `json:"candidate_variant"`
	BaselineVariant  string                   `json:"baseline_variant"`
	BeatsCLI         bool                     `json:"beats_cli"`
	Recommendation   string                   `json:"recommendation"`
	Criteria         []codeFirstCriterion     `json:"criteria"`
	Entries          []codeFirstComparisonRow `json:"entries"`
}

type codeFirstCriterion struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

type codeFirstComparisonRow struct {
	Scenario       string `json:"scenario"`
	CandidatePass  bool   `json:"candidate_pass"`
	CLIPass        bool   `json:"cli_pass"`
	CandidateTools int    `json:"candidate_tools"`
	CLITools       int    `json:"cli_tools"`
	ToolDelta      *int   `json:"tool_delta,omitempty"`
}

type comparisonEntry struct {
	Variant                            string   `json:"variant"`
	Scenario                           string   `json:"scenario"`
	Result                             string   `json:"result"`
	ToolCallsDelta                     *int     `json:"tool_calls_delta,omitempty"`
	AssistantCallsDelta                *int     `json:"assistant_calls_delta,omitempty"`
	WallSecondsDelta                   *float64 `json:"wall_seconds_delta,omitempty"`
	NonCachedInputTokensDelta          *int     `json:"non_cached_input_tokens_delta,omitempty"`
	GeneratedFileInspectionChange      string   `json:"generated_file_inspection_change"`
	GeneratedPathFromBroadSearchChange string   `json:"generated_path_from_broad_search_change"`
	BroadRepoSearchChange              string   `json:"broad_repo_search_change"`
	ModuleCacheInspectionChange        string   `json:"module_cache_inspection_change"`
	CLIUsageChange                     string   `json:"cli_usage_change"`
	DirectSQLiteAccessChange           string   `json:"direct_sqlite_access_change"`
}

type weightState struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type bloodPressureState struct {
	Date      string `json:"date"`
	Systolic  int    `json:"systolic"`
	Diastolic int    `json:"diastolic"`
	Pulse     *int   `json:"pulse,omitempty"`
}

type codexEvent struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id"`
	Item     item   `json:"item"`
	Usage    *usage `json:"usage"`
}

type item struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Text             string `json:"text"`
	Command          string `json:"command"`
	AggregatedOutput string `json:"aggregated_output"`
}

type usage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens"`
	OutputTokens      int `json:"output_tokens"`
}

type runOptions struct {
	RunRoot          string
	Date             string
	CompareTo        string
	VariantFilter    string
	ScenarioFilter   string
	CandidateVariant string
	Parallelism      int
	CacheMode        string
}

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
		Date:             time.Now().Format(time.DateOnly),
		CandidateVariant: "production",
		Parallelism:      defaultRunParallelism,
		CacheMode:        cacheModeShared,
	}
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&options.RunRoot, "run-root", options.RunRoot, "directory for raw run artifacts outside the repo")
	fs.StringVar(&options.Date, "date", options.Date, "report date in YYYY-MM-DD form")
	fs.StringVar(&options.CompareTo, "compare-to", options.CompareTo, "optional baseline JSON report path for comparison")
	fs.StringVar(&options.VariantFilter, "variant", options.VariantFilter, "optional comma-separated variant ids to run")
	fs.StringVar(&options.ScenarioFilter, "scenario", options.ScenarioFilter, "optional comma-separated scenario ids to run")
	fs.StringVar(&options.CandidateVariant, "candidate", options.CandidateVariant, "candidate variant id to compare directly with cli")
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

	marker := filepath.Join(runRoot, "history-marker")
	if err := os.WriteFile(marker, []byte(time.Now().Format(time.RFC3339Nano)), 0o644); err != nil {
		failf("write history marker: %v", err)
	}
	markerInfo, err := os.Stat(marker)
	if err != nil {
		failf("stat history marker: %v", err)
	}

	codexVersion := commandOutput("codex", "--version")
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
		Harness:               "codex exec --json --full-auto from throwaway run directories; single-turn scenarios use --ephemeral, multi-turn scenarios resume a persisted eval session with explicit writable eval roots",
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
			VerificationMethod:         "Single-turn scenarios use codex exec --ephemeral from <run-root>/.../repo. Multi-turn scenarios create one persisted Codex exec session per variant/scenario and resume it for later turns; all raw logs stay under <run-root>.",
			VerificationLimitation:     limitation,
		},
		CommandTemplate: []string{
			"OPENHEALTH_DATABASE_PATH=<run-root>/<variant>/<scenario>/repo/openhealth.db",
			"GOCACHE=<run-root>/shared-cache/gocache when --cache-mode shared; otherwise <run-root>/<variant>/<scenario>/gocache",
			"GOMODCACHE=<run-root>/shared-cache/gomodcache when --cache-mode shared; otherwise <run-root>/<variant>/<scenario>/gomodcache",
			"single turn: codex exec --json --ephemeral --full-auto --skip-git-repo-check --add-dir <run-root>/<variant>/<scenario> --add-dir <run-root>/shared-cache when --cache-mode shared -C <run-root>/<variant>/<scenario>/repo -m gpt-5.4-mini -c model_reasoning_effort=\"medium\" -c shell_environment_policy.inherit=all <natural user prompt>",
			"multi turn: first turn uses codex exec without --ephemeral; later turns use codex exec -C <run-root>/<variant>/<scenario>/repo --add-dir <writable-eval-roots> resume <thread-id> --json with per-turn logs",
		},
		CLIStatus:         "runnable: cli variant uses go run ./cmd/openhealth weight and blood-pressure commands with the configured Go cache mode",
		MetricNotes:       metricNotes(options.Date, results),
		StopLoss:          productionStopLoss(results),
		CodeFirst:         codeFirstSummaryFor(results, options.CandidateVariant),
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

type evalJob struct {
	Index    int
	Variant  variant
	Scenario scenario
}

type evalJobResult struct {
	Index  int
	Result runResult
}

type cacheConfig struct {
	Mode    string
	RunRoot string
}

type runOneFunc func(repoRoot string, runRoot string, currentVariant variant, currentScenario scenario, cache cacheConfig) (runResult, error)

func evalJobsFor(selectedVariants []variant, selectedScenarios []scenario) []evalJob {
	jobs := make([]evalJob, 0, len(selectedVariants)*len(selectedScenarios))
	for _, currentVariant := range selectedVariants {
		for _, currentScenario := range selectedScenarios {
			jobs = append(jobs, evalJob{
				Index:    len(jobs),
				Variant:  currentVariant,
				Scenario: currentScenario,
			})
		}
	}
	return jobs
}

func runEvalJobs(repoRoot string, runRoot string, jobs []evalJob, parallelism int, cache cacheConfig, runOne runOneFunc) []runResult {
	if parallelism < 1 {
		parallelism = 1
	}
	workerCount := parallelism
	if workerCount > len(jobs) {
		workerCount = len(jobs)
	}
	results := make([]runResult, len(jobs))
	if workerCount == 0 {
		return results
	}

	jobsCh := make(chan evalJob)
	resultsCh := make(chan evalJobResult)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsCh {
				result, err := runOne(repoRoot, runRoot, job.Variant, job.Scenario, cache)
				if err != nil {
					result = harnessErrorResult(job.Variant, job.Scenario, err)
				}
				resultsCh <- evalJobResult{Index: job.Index, Result: result}
			}
		}()
	}

	go func() {
		for _, job := range jobs {
			jobsCh <- job
		}
		close(jobsCh)
		wg.Wait()
		close(resultsCh)
	}()

	for result := range resultsCh {
		results[result.Index] = result.Result
	}
	return results
}

func harnessErrorResult(currentVariant variant, currentScenario scenario, err error) runResult {
	return runResult{
		Variant:       currentVariant.ID,
		Scenario:      currentScenario.ID,
		ScenarioTitle: currentScenario.Title,
		ExitCode:      -1,
		Verification: verificationResult{
			Passed:  false,
			Details: fmt.Sprintf("harness error: %v", err),
		},
		PromptSummary: promptSummary(currentScenario),
	}
}

func runOne(repoRoot string, runRoot string, currentVariant variant, currentScenario scenario, cache cacheConfig) (runResult, error) {
	totalStart := time.Now()
	runDir := filepath.Join(runRoot, currentVariant.ID, currentScenario.ID)
	runRepo := filepath.Join(runDir, "repo")
	dbPath := evalDatabasePath(runRepo)
	timings := phaseTimings{}

	if err := timedPhase(&timings.PrepareRunDir, func() error { return prepareRunDir(runDir, cache) }); err != nil {
		return runResult{}, fmt.Errorf("prepare run dir: %w", err)
	}
	if err := timedPhase(&timings.CopyRepo, func() error { return copyRepo(repoRoot, runRepo) }); err != nil {
		return runResult{}, fmt.Errorf("copy repo: %w", err)
	}
	if err := timedPhase(&timings.InstallVariant, func() error { return installVariant(repoRoot, runRepo, currentVariant) }); err != nil {
		return runResult{}, fmt.Errorf("install variant: %w", err)
	}
	if cache.Mode == cacheModeIsolated {
		if err := timedPhase(&timings.WarmCache, func() error { return warmGoModules(runRepo, runDir, dbPath, cache) }); err != nil {
			return runResult{}, fmt.Errorf("warm go modules: %w", err)
		}
	}
	if err := timedPhase(&timings.SeedDB, func() error { return seedScenario(dbPath, currentScenario) }); err != nil {
		return runResult{}, fmt.Errorf("seed scenario: %w", err)
	}

	turns := scenarioTurns(currentScenario)
	turnResults := make([]turnResult, 0, len(turns))
	sessionID := ""
	var agentErr error
	for i, turn := range turns {
		turnIndex := i + 1
		result, parsed, err := runScenarioTurn(runRepo, runDir, dbPath, currentVariant, currentScenario, turn, turnIndex, sessionID, cache)
		timings.AgentRun += result.WallSeconds

		timings.ParseMetrics += parsed.parseSeconds
		if parsed.parseError != nil {
			result.Metrics.CommandMetricLimitations = fmt.Sprintf("failed to parse event log: %v", parsed.parseError)
		}

		verifyStart := time.Now()
		verification, verifyErr := verifyScenarioTurn(dbPath, currentScenario, turnIndex, parsed.finalMessage)
		timings.Verify += roundSeconds(time.Since(verifyStart).Seconds())
		if verifyErr != nil {
			verification = verificationResult{
				Passed:  false,
				Details: fmt.Sprintf("verification error: %v", verifyErr),
			}
		}
		result.Verification = verification
		turnResults = append(turnResults, result)
		if err != nil && agentErr == nil {
			agentErr = err
		}
		if verifyErr != nil && agentErr == nil {
			agentErr = verifyErr
		}
		if i == 0 && len(turns) > 1 {
			sessionID = parsed.sessionID
			if sessionID == "" && agentErr == nil {
				agentErr = errors.New("multi-turn first turn did not expose a thread id")
			}
		}
	}

	metrics := aggregateMetrics(turnResults)
	verification := aggregateVerification(currentScenario, turnResults)
	timings.Total = roundSeconds(time.Since(totalStart).Seconds())
	exitCode := aggregateExitCode(turnResults)
	rawLogRef := ""
	if len(turnResults) > 0 {
		rawLogRef = turnResults[len(turnResults)-1].RawLogArtifactReference
	}
	result := runResult{
		Variant:                 currentVariant.ID,
		Scenario:                currentScenario.ID,
		ScenarioTitle:           currentScenario.Title,
		Passed:                  agentErr == nil && verification.Passed,
		ExitCode:                exitCode,
		WallSeconds:             roundSeconds(sumTurnWallSeconds(turnResults)),
		PhaseTimings:            timings.rounded(),
		Metrics:                 metrics,
		Verification:            verification,
		Turns:                   turnResults,
		PromptSummary:           promptSummary(currentScenario),
		RawLogArtifactReference: rawLogRef,
	}
	runSummaryPath := filepath.Join(runDir, "run-summary.json")
	_ = writeJSON(runSummaryPath, result)
	return result, nil
}

func timedPhase(target *float64, fn func() error) error {
	start := time.Now()
	err := fn()
	*target += roundSeconds(time.Since(start).Seconds())
	return err
}

func evalDatabasePath(runRepo string) string {
	return filepath.Join(runRepo, "openhealth.db")
}

func (p phaseTimings) rounded() phaseTimings {
	return phaseTimings{
		PrepareRunDir:  roundSeconds(p.PrepareRunDir),
		CopyRepo:       roundSeconds(p.CopyRepo),
		InstallVariant: roundSeconds(p.InstallVariant),
		WarmCache:      roundSeconds(p.WarmCache),
		SeedDB:         roundSeconds(p.SeedDB),
		AgentRun:       roundSeconds(p.AgentRun),
		ParseMetrics:   roundSeconds(p.ParseMetrics),
		Verify:         roundSeconds(p.Verify),
		Total:          roundSeconds(p.Total),
	}
}

func aggregatePhaseTimings(results []runResult) phaseTimings {
	total := phaseTimings{}
	for _, result := range results {
		total.PrepareRunDir += result.PhaseTimings.PrepareRunDir
		total.CopyRepo += result.PhaseTimings.CopyRepo
		total.InstallVariant += result.PhaseTimings.InstallVariant
		total.WarmCache += result.PhaseTimings.WarmCache
		total.SeedDB += result.PhaseTimings.SeedDB
		total.AgentRun += result.PhaseTimings.AgentRun
		total.ParseMetrics += result.PhaseTimings.ParseMetrics
		total.Verify += result.PhaseTimings.Verify
		total.Total += result.PhaseTimings.Total
	}
	return total.rounded()
}

func totalAgentWallSeconds(results []runResult) float64 {
	total := 0.0
	for _, result := range results {
		total += result.WallSeconds
	}
	return total
}

func sumTurnWallSeconds(turns []turnResult) float64 {
	total := 0.0
	for _, turn := range turns {
		total += turn.WallSeconds
	}
	return total
}

func aggregateExitCode(turns []turnResult) int {
	for _, turn := range turns {
		if turn.ExitCode != 0 {
			return turn.ExitCode
		}
	}
	return 0
}

func aggregateMetrics(turns []turnResult) metrics {
	out := metrics{
		EventTypeCounts:          map[string]int{},
		CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not from OS-level tracing.",
	}
	allUsageExposed := len(turns) > 0
	inputTotal := 0
	cachedTotal := 0
	nonCachedTotal := 0
	outputTotal := 0
	for _, turn := range turns {
		current := turn.Metrics
		out.AssistantCalls += current.AssistantCalls
		out.ToolCalls += current.ToolCalls
		out.CommandExecutions += current.CommandExecutions
		out.FileInspectionCommands += current.FileInspectionCommands
		out.GeneratedFileInspected = out.GeneratedFileInspected || current.GeneratedFileInspected
		out.GeneratedPathFromBroadSearch = out.GeneratedPathFromBroadSearch || current.GeneratedPathFromBroadSearch
		out.BroadRepoSearch = out.BroadRepoSearch || current.BroadRepoSearch
		out.ModuleCacheInspected = out.ModuleCacheInspected || current.ModuleCacheInspected
		out.CLIUsed = out.CLIUsed || current.CLIUsed
		out.DirectSQLiteAccess = out.DirectSQLiteAccess || current.DirectSQLiteAccess
		out.GeneratedFileEvidence = append(out.GeneratedFileEvidence, current.GeneratedFileEvidence...)
		out.GeneratedPathFromBroadSearchEvidence = append(out.GeneratedPathFromBroadSearchEvidence, current.GeneratedPathFromBroadSearchEvidence...)
		out.BroadRepoSearchEvidence = append(out.BroadRepoSearchEvidence, current.BroadRepoSearchEvidence...)
		out.ModuleCacheEvidence = append(out.ModuleCacheEvidence, current.ModuleCacheEvidence...)
		out.CLIUsageEvidence = append(out.CLIUsageEvidence, current.CLIUsageEvidence...)
		out.DirectSQLiteEvidence = append(out.DirectSQLiteEvidence, current.DirectSQLiteEvidence...)
		for eventType, count := range current.EventTypeCounts {
			out.EventTypeCounts[eventType] += count
		}
		if !current.UsageExposed || current.InputTokens == nil || current.CachedInputTokens == nil || current.NonCachedInputTokens == nil || current.OutputTokens == nil {
			allUsageExposed = false
			continue
		}
		inputTotal += *current.InputTokens
		cachedTotal += *current.CachedInputTokens
		nonCachedTotal += *current.NonCachedInputTokens
		outputTotal += *current.OutputTokens
	}
	if allUsageExposed {
		out.UsageExposed = true
		out.InputTokens = &inputTotal
		out.CachedInputTokens = &cachedTotal
		out.NonCachedInputTokens = &nonCachedTotal
		out.OutputTokens = &outputTotal
	}
	return out
}

func aggregateVerification(sc scenario, turns []turnResult) verificationResult {
	out := verificationResult{
		DatabasePass:  true,
		AssistantPass: true,
		Passed:        true,
	}
	details := []string{}
	for _, turn := range turns {
		verification := turn.Verification
		if !verification.DatabasePass {
			out.DatabasePass = false
		}
		if !verification.AssistantPass {
			out.AssistantPass = false
		}
		if !verification.Passed {
			out.Passed = false
		}
		if verification.Details != "" {
			details = append(details, fmt.Sprintf("turn %d: %s", turn.Index, verification.Details))
		}
		out.Weights = verification.Weights
		out.BloodPressures = verification.BloodPressures
	}
	if len(details) > 0 {
		out.Details = strings.Join(details, "; ")
	}
	if len(turns) == 0 {
		out.Passed = false
		out.DatabasePass = false
		out.AssistantPass = false
		out.Details = fmt.Sprintf("scenario %s did not run any turns", sc.ID)
	}
	return out
}

func countSingleTurnJobs(jobs []evalJob) int {
	count := 0
	for _, job := range jobs {
		if !isMultiTurnScenario(job.Scenario) {
			count++
		}
	}
	return count
}

func countMultiTurnJobs(jobs []evalJob) int {
	count := 0
	for _, job := range jobs {
		if isMultiTurnScenario(job.Scenario) {
			count++
		}
	}
	return count
}

func countMultiTurnPersistedTurns(jobs []evalJob) int {
	count := 0
	for _, job := range jobs {
		if isMultiTurnScenario(job.Scenario) {
			count += len(scenarioTurns(job.Scenario))
		}
	}
	return count
}

func historyIsolationStatus(newSessionFiles int, expectedPersistedSessions int) (string, string) {
	if newSessionFiles == expectedPersistedSessions {
		return "passed", ""
	}
	if expectedPersistedSessions == 0 {
		return "review", "Session-file count changed even though only single-turn ephemeral scenarios ran; this may be from another Codex process."
	}
	if newSessionFiles < expectedPersistedSessions {
		return "review", "Fewer session files referenced <run-root> than expected for persisted multi-turn eval sessions."
	}
	return "review", "More session files referenced <run-root> than expected for persisted multi-turn eval sessions."
}

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
	env := os.Environ()
	paths := evalPathsFor(runDir, cache)
	env = append(env,
		"OPENHEALTH_DATABASE_PATH="+dbPath,
		"OPENHEALTH_DATA_DIR=",
		"GOCACHE="+paths.GoCache,
		"GOMODCACHE="+paths.GoModCache,
		"TMPDIR="+paths.Temp,
	)
	return env
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

func variants() []variant {
	return []variant{
		{ID: "production", Title: "Production AgentOps skill"},
		{ID: "cli", Title: "CLI-oriented skill"},
	}
}

func scenarios() []scenario {
	return []scenario{
		{
			ID:     "add-two",
			Title:  "Add two weights and verify newest-first output",
			Prompt: "I need to update my local OpenHealth weight history. Use the configured local OpenHealth data path, and use year 2026 for these short dates: 03/29 152.2 lbs and 03/30 151.6. Then tell me the newest-first entries you see.",
		},
		{
			ID:     "repeat-add",
			Title:  "Repeat the same add request without creating duplicates",
			Prompt: "I need to record these OpenHealth weights again in the configured local data path, using year 2026 for the short dates: 03/29 152.2 lbs and 03/30 151.6. Tell me what is stored newest-first after you apply the request.",
		},
		{
			ID:     "update-existing",
			Title:  "Correct an existing same-date weight value",
			Prompt: "I need to correct my local OpenHealth weight for 03/29/2026. It should be 151.6 lbs, not 152.2 lbs. Update the configured local data and tell me what is stored now.",
		},
		{
			ID:     "bounded-range",
			Title:  "List a bounded date range newest-first",
			Prompt: "Please show every local OpenHealth weight history entry dated 03/29/2026 or 03/30/2026 only, newest first. Use the configured local data path. Include each entry dated 03/30/2026 and 03/29/2026 if present. In the final answer, mention only entries in that requested range; do not mention excluded dates at all.",
		},
		{
			ID:     "bounded-range-natural",
			Title:  "List a naturally phrased bounded date range",
			Prompt: "Can you show my OpenHealth weights for Mar 29 and Mar 30, 2026, newest first? Use the configured local data path.",
		},
		{
			ID:     "latest-only",
			Title:  "List only the latest weight entry",
			Prompt: "What is my latest local OpenHealth weight entry? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "history-limit-two",
			Title:  "List the two most recent weight entries",
			Prompt: "Please show my two most recent local OpenHealth weight entries, newest first. Use the configured local data path and mention only those two rows.",
		},
		{
			ID:     "ambiguous-short-date",
			Title:  "Clarify an ambiguous short date without writing",
			Prompt: "Please add this local OpenHealth weight: 03/29 152.2 lbs. There is no year context in this conversation or my request.",
		},
		{
			ID:     "invalid-input",
			Title:  "Reject invalid unit and value without writing",
			Prompt: "Please add this local OpenHealth weight entry: 03/31/2026 -5 stone.",
		},
		{
			ID:     "non-iso-date-reject",
			Title:  "Reject non-ISO full date without writing",
			Prompt: "Please add this local OpenHealth weight entry exactly as written: 2026/03/31 152.2 lbs. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "bp-add-two",
			Title:  "Record two blood-pressure readings and verify newest-first output",
			Prompt: "I need to update my local OpenHealth blood pressure history. Use the configured local OpenHealth data path, and use year 2026 for these short dates: 03/29 122/78 pulse 64 and 03/30 118/76. Then tell me the newest-first entries you see.",
		},
		{
			ID:     "bp-latest-only",
			Title:  "List only the latest blood-pressure reading",
			Prompt: "What is my latest local OpenHealth blood pressure reading? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "bp-history-limit-two",
			Title:  "List the two most recent blood-pressure readings",
			Prompt: "Please show my two most recent local OpenHealth blood pressure readings, newest first. Use the configured local data path and mention only those two rows.",
		},
		{
			ID:     "bp-bounded-range",
			Title:  "List a bounded blood-pressure date range newest-first",
			Prompt: "Please show every local OpenHealth blood pressure reading dated 03/29/2026 or 03/30/2026 only, newest first. Use the configured local data path. Include each reading dated 03/30/2026 and 03/29/2026 if present. In the final answer, mention only readings in that requested range; do not mention excluded dates at all.",
		},
		{
			ID:     "bp-bounded-range-natural",
			Title:  "List a naturally phrased bounded blood-pressure date range",
			Prompt: "Can you show my OpenHealth blood pressure readings for Mar 29 and Mar 30, 2026, newest first? Use the configured local data path.",
		},
		{
			ID:     "bp-invalid-input",
			Title:  "Reject invalid blood-pressure values without writing",
			Prompt: "Please add this local OpenHealth blood pressure reading: 03/31/2026 0/-5 pulse 0.",
		},
		{
			ID:     "bp-non-iso-date-reject",
			Title:  "Reject non-ISO blood-pressure date without writing",
			Prompt: "Please add this local OpenHealth blood pressure reading exactly as written: 2026/03/31 122/78. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "bp-correct-existing",
			Title:  "Correct an existing same-date blood-pressure reading",
			Prompt: "I need to correct my local OpenHealth blood pressure reading for 03/29/2026. It should be 121/77 pulse 63, not 122/78 pulse 64. Update the configured local data and tell me what is stored now.",
		},
		{
			ID:     "bp-correct-missing-reject",
			Title:  "Reject a blood-pressure correction for a missing date",
			Prompt: "Please correct my local OpenHealth blood pressure reading for 03/31/2026 to 121/77. If there is no reading for that date, do not create a new one; tell me why it was not updated.",
		},
		{
			ID:     "bp-correct-ambiguous-reject",
			Title:  "Reject an ambiguous same-date blood-pressure correction",
			Prompt: "Please correct my local OpenHealth blood pressure reading for 03/29/2026 to 121/77. If more than one reading exists for that date, do not guess; tell me why it was not updated.",
		},
		{
			ID:     "mixed-add-latest",
			Title:  "Record weight and blood-pressure readings, then report latest for both",
			Prompt: "Use the configured local OpenHealth data path. Record weight 150.8 lbs and blood pressure 119/77 pulse 62 for 03/31/2026. Then tell me the latest weight and latest blood-pressure entries.",
		},
		{
			ID:     "mixed-bounded-range",
			Title:  "List bounded weight and blood-pressure ranges newest-first",
			Prompt: "Use the configured local OpenHealth data path. Show my OpenHealth weights and blood pressure readings for Mar 29 and Mar 30, 2026 only, newest first in each domain. Do not mention entries outside that requested range.",
		},
		{
			ID:     "mixed-invalid-direct-reject",
			Title:  "Reject invalid mixed-domain values without writing",
			Prompt: "Please add these local OpenHealth entries: weight 03/31/2026 -5 stone and blood pressure 03/31/2026 0/-5 pulse 0.",
		},
		{
			ID:    "mt-weight-clarify-then-add",
			Title: "Clarify missing year, then add weight in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Please add this local OpenHealth weight: 03/29 152.2 lbs. There is no year context in this conversation or my request."},
				{Prompt: "Use 2026 as the year for that weight entry."},
			},
		},
		{
			ID:    "mt-mixed-latest-then-correct",
			Title: "Read latest mixed-domain entries, then correct both in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenHealth data path. What are my latest weight and blood-pressure entries? Mention only the latest row from each domain."},
				{Prompt: "Correct both latest entries for that same date: weight should be 151.0 lbs and blood pressure should be 117/75 pulse 63. Tell me what is stored now."},
			},
		},
		{
			ID:    "mt-bp-latest-then-correct",
			Title: "Read latest blood pressure, then correct it in a resumed turn",
			Turns: []scenarioTurn{
				{Prompt: "Use the configured local OpenHealth data path. What is my latest blood-pressure reading? Mention only the latest row."},
				{Prompt: "Correct that latest reading to 117/75 pulse 63. Tell me what is stored now."},
			},
		},
	}
}

func selectVariants(filter string) ([]variant, error) {
	all := variants()
	if strings.TrimSpace(filter) == "" {
		return all, nil
	}
	selected := []variant{}
	for _, id := range splitFilterIDs(filter) {
		found := false
		for _, candidate := range all {
			if candidate.ID == id {
				selected = append(selected, candidate)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown variant %q", id)
		}
	}
	return selected, nil
}

func selectScenarios(filter string) ([]scenario, error) {
	all := scenarios()
	if strings.TrimSpace(filter) == "" {
		return all, nil
	}
	selected := []scenario{}
	for _, id := range splitFilterIDs(filter) {
		found := false
		for _, candidate := range all {
			if candidate.ID == id {
				selected = append(selected, candidate)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown scenario %q", id)
		}
	}
	return selected, nil
}

func splitFilterIDs(filter string) []string {
	ids := []string{}
	for _, raw := range strings.Split(filter, ",") {
		id := strings.TrimSpace(raw)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func scenarioTurns(sc scenario) []scenarioTurn {
	if len(sc.Turns) > 0 {
		return sc.Turns
	}
	return []scenarioTurn{{Prompt: sc.Prompt}}
}

func isMultiTurnScenario(sc scenario) bool {
	return len(scenarioTurns(sc)) > 1
}

func scenarioByID(id string) (scenario, bool) {
	for _, sc := range scenarios() {
		if sc.ID == id {
			return sc, true
		}
	}
	return scenario{}, false
}

func seedScenario(dbPath string, sc scenario) error {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return err
	}
	defer func() {
		_ = api.Close()
	}()

	ctx := context.Background()
	switch sc.ID {
	case "mixed-bounded-range":
		if err := upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		}); err != nil {
			return err
		}
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "mt-mixed-latest-then-correct":
		if err := upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		}); err != nil {
			return err
		}
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "mt-bp-latest-then-correct":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "repeat-add":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "update-existing":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		})
	case "bounded-range", "bounded-range-natural", "latest-only":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "history-limit-two":
		return upsertWeights(ctx, api, []weightState{
			{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		})
	case "bp-latest-only", "bp-bounded-range", "bp-bounded-range-natural":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "bp-history-limit-two":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		})
	case "bp-correct-existing", "bp-correct-missing-reject":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		})
	case "bp-correct-ambiguous-reject":
		return recordBloodPressures(ctx, api, []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
		})
	default:
		return nil
	}
}

func upsertWeights(ctx context.Context, api *client.LocalClient, weights []weightState) error {
	for _, weight := range weights {
		recordedAt, err := parseDate(weight.Date)
		if err != nil {
			return err
		}
		if _, err := api.UpsertWeight(ctx, client.WeightRecordInput{
			RecordedAt: recordedAt,
			Value:      weight.Value,
			Unit:       client.WeightUnit(weight.Unit),
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordBloodPressures(ctx context.Context, api *client.LocalClient, readings []bloodPressureState) error {
	for _, reading := range readings {
		recordedAt, err := parseDate(reading.Date)
		if err != nil {
			return err
		}
		if _, err := api.RecordBloodPressure(ctx, client.BloodPressureRecordInput{
			RecordedAt: recordedAt,
			Systolic:   reading.Systolic,
			Diastolic:  reading.Diastolic,
			Pulse:      reading.Pulse,
		}); err != nil {
			return err
		}
	}
	return nil
}

func verifyScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	return verifyScenarioTurn(dbPath, sc, len(scenarioTurns(sc)), finalMessage)
}

func verifyScenarioTurn(dbPath string, sc scenario, turnIndex int, finalMessage string) (verificationResult, error) {
	if isMixedScenario(sc.ID) || strings.HasPrefix(sc.ID, "mt-") {
		return verifyMixedOrMultiTurnScenario(dbPath, sc, turnIndex, finalMessage)
	}
	if isBloodPressureScenario(sc.ID) {
		return verifyBloodPressureScenario(dbPath, sc, finalMessage)
	}

	weights, err := listWeights(dbPath)
	var states []weightState
	listErrorDetail := ""
	if err != nil {
		rawStates, rawErr := listRawWeights(dbPath)
		if rawErr != nil {
			return verificationResult{}, fmt.Errorf("list weights: %w; raw fallback: %w", err, rawErr)
		}
		states = rawStates
		listErrorDetail = fmt.Sprintf(" typed list error: %v.", err)
	} else {
		states = weightStates(weights)
	}

	result := verificationResult{
		Weights: states,
	}
	switch sc.ID {
	case "add-two", "repeat-add":
		expected := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected exactly two newest-first rows; observed %s%s", describeWeights(states), listErrorDetail)
	case "update-existing":
		expected := []weightState{
			{Date: "2026-03-29", Value: 151.6, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected one updated row; observed %s%s", describeWeights(states), listErrorDetail)
	case "bounded-range", "bounded-range-natural":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = boundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed %s%s", describeWeights(states), listErrorDetail)
	case "latest-only":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = latestOnlyAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest row 2026-03-30; observed %s%s", describeWeights(states), listErrorDetail)
	case "history-limit-two":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
			{Date: "2026-03-27", Value: 154.1, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = historyLimitTwoAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed %s%s", describeWeights(states), listErrorDetail)
	case "ambiguous-short-date":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"year", "which year", "clarify", "ambiguous"})
		result.Details = fmt.Sprintf("expected no write and a year clarification; observed %s%s", describeWeights(states), listErrorDetail)
	case "invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "unsupported", "positive", "cannot", "can't", "unit", "value", "lb", "pounds"})
		result.Details = fmt.Sprintf("expected no write and an invalid input rejection; observed %s%s", describeWeights(states), listErrorDetail)
	case "non-iso-date-reject":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and a strict YYYY-MM-DD date rejection; observed %s%s", describeWeights(states), listErrorDetail)
	default:
		return verificationResult{}, fmt.Errorf("unknown scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyMixedOrMultiTurnScenario(dbPath string, sc scenario, turnIndex int, finalMessage string) (verificationResult, error) {
	weights, err := listWeights(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list weights: %w", err)
	}
	bloodPressures, err := listBloodPressures(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list blood pressures: %w", err)
	}
	weightStates := weightStates(weights)
	bloodPressureStates := bloodPressureStates(bloodPressures)
	result := verificationResult{
		Weights:        weightStates,
		BloodPressures: bloodPressureStates,
	}

	switch sc.ID {
	case "mixed-add-latest":
		expectedWeights := []weightState{{Date: "2026-03-31", Value: 150.8, Unit: "lb"}}
		expectedReadings := []bloodPressureState{{Date: "2026-03-31", Systolic: 119, Diastolic: 77, Pulse: intPointer(62)}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = mixedLatestAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected latest weight and blood-pressure rows for 2026-03-31; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mixed-bounded-range":
		expectedWeights := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = mixedBoundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged mixed seed rows and output limited to 2026-03-29..2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mixed-invalid-direct-reject":
		result.DatabasePass = len(weightStates) == 0 && len(bloodPressureStates) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "positive", "unsupported", "cannot", "can't", "reject", "stone", "systolic", "diastolic"})
		result.Details = fmt.Sprintf("expected no mixed-domain writes and a direct invalid input rejection; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mt-weight-clarify-then-add":
		if turnIndex == 1 {
			result.DatabasePass = len(weightStates) == 0 && len(bloodPressureStates) == 0
			result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"year", "which year", "clarify", "ambiguous"})
			result.Details = fmt.Sprintf("expected no first-turn writes and a year clarification; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 152.2, Unit: "lb"}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "152.2"})
		result.Details = fmt.Sprintf("expected second-turn weight write after year clarification with no blood-pressure writes; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mt-mixed-latest-then-correct":
		if turnIndex == 1 {
			expectedWeights := []weightState{
				{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
				{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			}
			expectedReadings := []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			}
			result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
			result.AssistantPass = mixedLatestSeedAssistantPass(finalMessage)
			result.Details = fmt.Sprintf("expected unchanged seed rows and latest mixed-domain answer; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedWeights := []weightState{
			{Date: "2026-03-30", Value: 151.0, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-30", "151", "117/75"})
		result.Details = fmt.Sprintf("expected latest mixed-domain corrections on 2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	case "mt-bp-latest-then-correct":
		if turnIndex == 1 {
			expectedReadings := []bloodPressureState{
				{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
				{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			}
			result.DatabasePass = len(weightStates) == 0 && bloodPressuresEqual(bloodPressureStates, expectedReadings)
			result.AssistantPass = bloodPressureLatestOnlyAssistantPass(finalMessage)
			result.Details = fmt.Sprintf("expected unchanged seed rows and latest blood-pressure answer; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
			break
		}
		expectedReadings := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 117, Diastolic: 75, Pulse: intPointer(63)},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = len(weightStates) == 0 && bloodPressuresEqual(bloodPressureStates, expectedReadings)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-30", "117/75"})
		result.Details = fmt.Sprintf("expected latest blood-pressure correction on 2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	default:
		return verificationResult{}, fmt.Errorf("unknown mixed or multi-turn scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyBloodPressureScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	readings, err := listBloodPressures(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list blood pressures: %w", err)
	}
	states := bloodPressureStates(readings)
	result := verificationResult{
		BloodPressures: states,
	}

	switch sc.ID {
	case "bp-add-two":
		expected := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expected)
		result.AssistantPass = true
		result.Details = fmt.Sprintf("expected exactly two newest-first blood-pressure rows; observed %s", describeBloodPressures(states))
	case "bp-latest-only":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureLatestOnlyAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest blood-pressure row 2026-03-30; observed %s", describeBloodPressures(states))
	case "bp-history-limit-two":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
			{Date: "2026-03-27", Systolic: 126, Diastolic: 82},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureHistoryLimitTwoAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-30 and 2026-03-29 newest-first; observed %s", describeBloodPressures(states))
	case "bp-bounded-range", "bp-bounded-range-natural":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
			{Date: "2026-03-28", Systolic: 124, Diastolic: 80},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = bloodPressureBoundedRangeAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed %s", describeBloodPressures(states))
	case "bp-correct-existing":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 121, Diastolic: 77, Pulse: intPointer(63)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "121/77"})
		result.Details = fmt.Sprintf("expected corrected 2026-03-29 blood-pressure row with no duplicate; observed %s", describeBloodPressures(states))
	case "bp-correct-missing-reject":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"no existing", "missing", "not found", "cannot", "can't", "did not", "not updated"})
		result.Details = fmt.Sprintf("expected unchanged seed row and missing-date correction rejection; observed %s", describeBloodPressures(states))
	case "bp-correct-ambiguous-reject":
		expectedDB := []bloodPressureState{
			{Date: "2026-03-29", Systolic: 120, Diastolic: 76},
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64)},
		}
		result.DatabasePass = bloodPressuresEqual(states, expectedDB)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"multiple", "ambiguous", "more than one", "cannot", "can't", "did not", "not updated"})
		result.Details = fmt.Sprintf("expected unchanged duplicate same-date rows and ambiguous correction rejection; observed %s", describeBloodPressures(states))
	case "bp-invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "positive", "cannot", "can't", "systolic", "diastolic", "pulse", "reject"})
		result.Details = fmt.Sprintf("expected no write and an invalid blood-pressure rejection; observed %s", describeBloodPressures(states))
	case "bp-non-iso-date-reject":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and a strict YYYY-MM-DD date rejection; observed %s", describeBloodPressures(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown blood-pressure scenario %q", sc.ID)
	}

	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func listWeights(dbPath string) ([]client.WeightEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListWeights(context.Background(), client.WeightListOptions{Limit: 100})
}

func listBloodPressures(dbPath string) ([]client.BloodPressureEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListBloodPressure(context.Background(), client.BloodPressureListOptions{Limit: 100})
}

func listRawWeights(dbPath string) ([]weightState, error) {
	db, err := storagesqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(context.Background(), `
SELECT recorded_at, value, unit
FROM health_weight_entry
WHERE deleted_at IS NULL
ORDER BY recorded_at DESC, id DESC
`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	states := []weightState{}
	for rows.Next() {
		var state weightState
		if err := rows.Scan(&state.Date, &state.Value, &state.Unit); err != nil {
			return nil, err
		}
		state.Value = roundWeight(state.Value)
		states = append(states, state)
	}
	return states, rows.Err()
}

func weightStates(weights []client.WeightEntry) []weightState {
	states := make([]weightState, 0, len(weights))
	for _, weight := range weights {
		states = append(states, weightState{
			Date:  weight.RecordedAt.Format(time.DateOnly),
			Value: roundWeight(weight.Value),
			Unit:  string(weight.Unit),
		})
	}
	return states
}

func bloodPressureStates(readings []client.BloodPressureEntry) []bloodPressureState {
	states := make([]bloodPressureState, 0, len(readings))
	for _, reading := range readings {
		states = append(states, bloodPressureState{
			Date:      reading.RecordedAt.Format(time.DateOnly),
			Systolic:  reading.Systolic,
			Diastolic: reading.Diastolic,
			Pulse:     reading.Pulse,
		})
	}
	return states
}

type parsedMetrics struct {
	metrics      metrics
	finalMessage string
	sessionID    string
}

func parseMetrics(path string) (parsedMetrics, error) {
	file, err := os.Open(path)
	if err != nil {
		return parsedMetrics{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	out := parsedMetrics{
		metrics: metrics{
			EventTypeCounts:          map[string]int{},
			CommandMetricLimitations: "Command/file inspection metrics are inferred from codex exec JSON command events, not from OS-level tracing.",
		},
	}
	commandIDs := map[string]struct{}{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		out.metrics.EventTypeCounts[event.Type]++
		if event.Type == "thread.started" && event.ThreadID != "" && out.sessionID == "" {
			out.sessionID = event.ThreadID
		}
		switch event.Item.Type {
		case "agent_message":
			if event.Type == "item.completed" {
				out.metrics.AssistantCalls++
				out.finalMessage = event.Item.Text
			}
		case "command_execution":
			if event.Item.ID != "" {
				commandIDs[event.Item.ID] = struct{}{}
			}
			if event.Type == "item.completed" {
				out.metrics.CommandExecutions++
				if isFileInspectionCommand(event.Item.Command) {
					out.metrics.FileInspectionCommands++
				}
				if isBroadRepoSearchCommand(event.Item.Command) {
					out.metrics.BroadRepoSearch = true
					addMetricEvidence(&out.metrics.BroadRepoSearchEvidence, event.Item.Command)
					if mentionsGeneratedPath(event.Item.AggregatedOutput) {
						out.metrics.GeneratedPathFromBroadSearch = true
						addMetricEvidence(&out.metrics.GeneratedPathFromBroadSearchEvidence, event.Item.Command)
					}
				}
				if inspectsGeneratedFileCommand(event.Item.Command, event.Item.AggregatedOutput) {
					out.metrics.GeneratedFileInspected = true
					addMetricEvidence(&out.metrics.GeneratedFileEvidence, event.Item.Command)
				}
				if inspectsModuleCache(event.Item.Command) {
					out.metrics.ModuleCacheInspected = true
					addMetricEvidence(&out.metrics.ModuleCacheEvidence, event.Item.Command)
				}
				if usesOpenHealthCLI(event.Item.Command) {
					out.metrics.CLIUsed = true
					addMetricEvidence(&out.metrics.CLIUsageEvidence, event.Item.Command)
				}
				if usesDirectSQLite(event.Item.Command) {
					out.metrics.DirectSQLiteAccess = true
					addMetricEvidence(&out.metrics.DirectSQLiteEvidence, event.Item.Command)
				}
			}
		}
		if event.Usage != nil {
			input := event.Usage.InputTokens
			cached := event.Usage.CachedInputTokens
			nonCached := input - cached
			output := event.Usage.OutputTokens
			out.metrics.UsageExposed = true
			out.metrics.InputTokens = &input
			out.metrics.CachedInputTokens = &cached
			out.metrics.NonCachedInputTokens = &nonCached
			out.metrics.OutputTokens = &output
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	out.metrics.ToolCalls = len(commandIDs)
	return out, nil
}

func copyRepo(src string, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if shouldSkipCopy(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func shouldSkipCopy(rel string, entry fs.DirEntry) bool {
	name := entry.Name()
	if shouldSkipEvalPath(filepath.ToSlash(rel)) {
		return true
	}
	if entry.IsDir() {
		switch name {
		case ".git", ".beads", ".dolt", ".agents":
			return true
		}
	}
	if strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".db-shm") || strings.HasSuffix(name, ".db-wal") {
		return true
	}
	return false
}

func shouldSkipEvalPath(rel string) bool {
	switch rel {
	case "AGENTS.md", "docs/agent-evals.md", "docs/agent-eval-assets", "docs/agent-eval-results", "scripts/agent-eval":
		return true
	}
	return strings.HasPrefix(rel, "docs/agent-eval-assets/") ||
		strings.HasPrefix(rel, "docs/agent-eval-results/") ||
		strings.HasPrefix(rel, "scripts/agent-eval/")
}

func copyFile(src string, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func installVariant(repoRoot string, runRepo string, currentVariant variant) error {
	dest := filepath.Join(runRepo, ".agents", "skills", "openhealth")
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	var src string
	switch currentVariant.ID {
	case "production":
		src = filepath.Join(repoRoot, "skills", "openhealth")
	case "cli":
		src = filepath.Join(repoRoot, "docs", "agent-eval-assets", "variants", "cli")
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	if err := copyDir(src, dest); err != nil {
		return err
	}
	if currentVariant.ID == "cli" {
		if err := normalizeSkillName(filepath.Join(dest, "SKILL.md"), "openhealth"); err != nil {
			return err
		}
	}
	return installVariantAgentsFile(runRepo, currentVariant)
}

func installVariantAgentsFile(runRepo string, currentVariant variant) error {
	var content string
	switch currentVariant.ID {
	case "production":
		content = `# OpenHealth Eval Instructions

For direct local OpenHealth weight or blood-pressure requests, act as a product data agent, not a repo maintainer. Do not inspect .agents, source/generated files, go.mod, the Go module cache, or SQLite, or search the repo before the first runner call.

Reject final-answer-only, with exactly one assistant answer and no tools or DB check, for ambiguous short dates with no year, year-first slash dates like 2026/03/31, non-positive values, or unsupported units. Do not first announce skill use or process. 03/29/2026 may become 2026-03-29.

For valid tasks, pipe JSON to go run ./cmd/openhealth-agentops weight or go run ./cmd/openhealth-agentops blood-pressure. Use one call per domain for mixed requests and answer from JSON only; entries are newest-first. Use history with limit:2 for "two most recent"; latest returns one row.

Every request JSON must include action. Exact one-line shapes:
{"action":"upsert_weights","weights":[{"date":"2026-03-29","value":152.2,"unit":"lb"}]}
{"action":"list_weights","list_mode":"latest"}
{"action":"list_weights","list_mode":"history","limit":2}
{"action":"list_weights","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
{"action":"record_blood_pressure","readings":[{"date":"2026-03-29","systolic":122,"diastolic":78,"pulse":64}]}
{"action":"correct_blood_pressure","readings":[{"date":"2026-03-29","systolic":121,"diastolic":77,"pulse":63}]}
{"action":"list_blood_pressure","list_mode":"latest"}
{"action":"list_blood_pressure","list_mode":"history","limit":2}
{"action":"list_blood_pressure","list_mode":"range","from_date":"2026-03-29","to_date":"2026-03-30"}
`
	case "cli":
		content = `# OpenHealth CLI Eval Instructions

For direct local OpenHealth weight or blood-pressure data requests, use the installed OpenHealth CLI baseline skill contract.

Reject ambiguous short dates with no year context, year-first slash dates such as 2026/03/31, non-positive values, and unsupported units directly in the final answer when the request is clearly invalid. Do not inspect files or run commands for those clearly invalid requests.

For valid supported tasks, use go run ./cmd/openhealth weight or go run ./cmd/openhealth blood-pressure commands. Use blood-pressure correct, not add, when the user asks to correct an existing same-date reading. Results are newest-first.
`
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	return os.WriteFile(filepath.Join(runRepo, "AGENTS.md"), []byte(content), 0o644)
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func normalizeSkillName(path string, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := bytes.Split(data, []byte("\n"))
	inFrontmatter := len(lines) > 0 && string(lines[0]) == "---"
	for i := 1; i < len(lines); i++ {
		line := string(lines[i])
		if inFrontmatter && strings.HasPrefix(line, "name:") {
			lines[i] = []byte("name: " + name)
			break
		}
		if i > 0 && line == "---" {
			break
		}
	}
	return os.WriteFile(path, bytes.Join(lines, []byte("\n")), 0o644)
}

func writeJSON(path string, value any) error {
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return os.WriteFile(path, data.Bytes(), 0o644)
}

func baselineReport(repoRoot string, outDir string, targetPath string, explicitPath string) (*report, string, error) {
	if explicitPath != "" {
		path := explicitPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoRoot, path)
		}
		value, err := readReport(path)
		if err != nil {
			return nil, "", err
		}
		return value, repoRelativePath(repoRoot, path), nil
	}

	if fileExists(targetPath) {
		value, err := readReport(targetPath)
		if err != nil {
			return nil, "", err
		}
		return value, repoRelativePath(repoRoot, targetPath), nil
	}

	matches, err := filepath.Glob(filepath.Join(outDir, issueID+"-*.json"))
	if err != nil {
		return nil, "", err
	}
	var latest string
	for _, match := range matches {
		if match == targetPath {
			continue
		}
		if latest == "" || match > latest {
			latest = match
		}
	}
	if latest == "" {
		return nil, "", nil
	}
	value, err := readReport(latest)
	if err != nil {
		return nil, "", err
	}
	return value, repoRelativePath(repoRoot, latest), nil
}

func readReport(path string) (*report, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	var value report
	if err := json.NewDecoder(file).Decode(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func compareReports(baseline report, current report, baselineRef string) *comparisonSummary {
	baselineByKey := map[string]runResult{}
	for _, result := range baseline.Results {
		baselineByKey[resultKey(result.Variant, result.Scenario)] = result
	}
	entries := make([]comparisonEntry, 0, len(current.Results))
	for _, currentResult := range current.Results {
		baselineResult, ok := baselineByKey[resultKey(currentResult.Variant, currentResult.Scenario)]
		entry := comparisonEntry{
			Variant:  currentResult.Variant,
			Scenario: currentResult.Scenario,
		}
		if !ok {
			entry.Result = "new"
			entries = append(entries, entry)
			continue
		}
		entry.Result = resultDelta(baselineResult.Passed, currentResult.Passed)
		entry.ToolCallsDelta = intDelta(baselineResult.Metrics.ToolCalls, currentResult.Metrics.ToolCalls)
		entry.AssistantCallsDelta = intDelta(baselineResult.Metrics.AssistantCalls, currentResult.Metrics.AssistantCalls)
		entry.WallSecondsDelta = floatDelta(baselineResult.WallSeconds, currentResult.WallSeconds)
		entry.NonCachedInputTokensDelta = optionalIntDelta(baselineResult.Metrics.NonCachedInputTokens, currentResult.Metrics.NonCachedInputTokens)
		entry.GeneratedFileInspectionChange = boolChange(baselineResult.Metrics.GeneratedFileInspected, currentResult.Metrics.GeneratedFileInspected)
		entry.GeneratedPathFromBroadSearchChange = boolChange(baselineResult.Metrics.GeneratedPathFromBroadSearch, currentResult.Metrics.GeneratedPathFromBroadSearch)
		entry.BroadRepoSearchChange = boolChange(baselineResult.Metrics.BroadRepoSearch, currentResult.Metrics.BroadRepoSearch)
		entry.ModuleCacheInspectionChange = boolChange(baselineResult.Metrics.ModuleCacheInspected, currentResult.Metrics.ModuleCacheInspected)
		entry.CLIUsageChange = boolChange(baselineResult.Metrics.CLIUsed, currentResult.Metrics.CLIUsed)
		entry.DirectSQLiteAccessChange = boolChange(baselineResult.Metrics.DirectSQLiteAccess, currentResult.Metrics.DirectSQLiteAccess)
		entries = append(entries, entry)
	}
	return &comparisonSummary{
		BaselineReport: baselineRef,
		Entries:        entries,
	}
}

func metricNotes(reportDate string, results []runResult) []string {
	productionGenerated := []string{}
	productionGeneratedFromBroad := []string{}
	productionBroadSearch := []string{}
	productionModuleCache := []string{}
	productionCLIUsage := []string{}
	productionDirectSQLite := []string{}
	for _, result := range results {
		if result.Variant != "production" {
			continue
		}
		if result.Metrics.GeneratedFileInspected {
			productionGenerated = append(productionGenerated, result.Scenario)
		}
		if result.Metrics.GeneratedPathFromBroadSearch {
			productionGeneratedFromBroad = append(productionGeneratedFromBroad, result.Scenario)
		}
		if result.Metrics.BroadRepoSearch {
			productionBroadSearch = append(productionBroadSearch, result.Scenario)
		}
		if result.Metrics.ModuleCacheInspected {
			productionModuleCache = append(productionModuleCache, result.Scenario)
		}
		if result.Metrics.CLIUsed {
			productionCLIUsage = append(productionCLIUsage, result.Scenario)
		}
		if result.Metrics.DirectSQLiteAccess {
			productionDirectSQLite = append(productionDirectSQLite, result.Scenario)
		}
	}

	notes := []string{}
	if strings.Contains(reportDate, "oh-23a") {
		notes = append(notes, "oh-23a intentionally keeps agent-facing readiness scoped to weight and blood pressure; labs and medications remain a separate AgentOps expansion tracked in oh-bng.")
	}
	if len(productionGenerated) > 0 {
		notes = append(notes, fmt.Sprintf("Production direct generated-file inspection remained in %s in the %s run.", strings.Join(productionGenerated, ", "), reportDate))
	}
	if len(productionGeneratedFromBroad) > 0 {
		notes = append(notes, fmt.Sprintf("Production generated paths surfaced from broad repo searches in %s in the %s run; this is tracked separately from direct generated-file inspection.", strings.Join(productionGeneratedFromBroad, ", "), reportDate))
	}
	if len(productionBroadSearch) > 0 {
		notes = append(notes, fmt.Sprintf("Production broad repo search remained in %s in the %s run.", strings.Join(productionBroadSearch, ", "), reportDate))
	}
	if len(productionModuleCache) > 0 {
		notes = append(notes, fmt.Sprintf("Production module-cache inspection remained in %s in the %s run.", strings.Join(productionModuleCache, ", "), reportDate))
	}
	if len(productionCLIUsage) > 0 {
		notes = append(notes, fmt.Sprintf("Production CLI usage remained in %s in the %s run.", strings.Join(productionCLIUsage, ", "), reportDate))
	}
	if len(productionDirectSQLite) > 0 {
		notes = append(notes, fmt.Sprintf("Production direct SQLite access remained in %s in the %s run.", strings.Join(productionDirectSQLite, ", "), reportDate))
	}
	return notes
}

func productionStopLoss(results []runResult) *stopLossSummary {
	productionByScenario := map[string]runResult{}
	cliByScenario := map[string]runResult{}
	for _, result := range results {
		switch result.Variant {
		case "production":
			productionByScenario[result.Scenario] = result
		case "cli":
			cliByScenario[result.Scenario] = result
		}
	}
	if len(productionByScenario) == 0 {
		return nil
	}

	triggers := []string{}
	failed := []string{}
	directGenerated := []string{}
	moduleCache := []string{}
	broadSearch := []string{}
	cliUsage := []string{}
	directSQLite := []string{}
	for _, result := range productionByScenario {
		if !result.Passed || !result.Verification.DatabasePass || !result.Verification.AssistantPass {
			failed = append(failed, result.Scenario)
		}
		if result.Metrics.GeneratedFileInspected {
			directGenerated = append(directGenerated, result.Scenario)
		}
		if result.Metrics.ModuleCacheInspected {
			moduleCache = append(moduleCache, result.Scenario)
		}
		if result.Metrics.BroadRepoSearch && isRoutineScenario(result.Scenario) {
			broadSearch = append(broadSearch, result.Scenario)
		}
		if result.Metrics.CLIUsed {
			cliUsage = append(cliUsage, result.Scenario)
		}
		if result.Metrics.DirectSQLiteAccess {
			directSQLite = append(directSQLite, result.Scenario)
		}
	}
	if len(failed) > 0 {
		triggers = append(triggers, fmt.Sprintf("production correctness below 100%% in %s", sortedJoin(failed)))
	}
	if len(directGenerated) > 0 {
		triggers = append(triggers, fmt.Sprintf("direct generated-file inspection in %s", sortedJoin(directGenerated)))
	}
	if len(moduleCache) > 0 {
		triggers = append(triggers, fmt.Sprintf("module-cache inspection in %s", sortedJoin(moduleCache)))
	}
	if len(broadSearch) > 1 {
		triggers = append(triggers, fmt.Sprintf("broad repo search in more than one routine scenario: %s", sortedJoin(broadSearch)))
	}
	if len(cliUsage) > 0 {
		triggers = append(triggers, fmt.Sprintf("production used the openhealth CLI in %s", sortedJoin(cliUsage)))
	}
	if len(directSQLite) > 0 {
		triggers = append(triggers, fmt.Sprintf("production used direct SQLite access in %s", sortedJoin(directSQLite)))
	}

	toolThresholds := map[string]int{
		"add-two":         10,
		"update-existing": 12,
		"bounded-range":   8,
	}
	for scenario, threshold := range toolThresholds {
		if result, ok := productionByScenario[scenario]; ok && result.Metrics.ToolCalls > threshold {
			triggers = append(triggers, fmt.Sprintf("production %s used %d tools, above threshold %d", scenario, result.Metrics.ToolCalls, threshold))
		}
	}

	moreThanDoubleCLI := []string{}
	for scenario, production := range productionByScenario {
		cli, ok := cliByScenario[scenario]
		if !ok || cli.Metrics.ToolCalls == 0 {
			continue
		}
		if production.Metrics.ToolCalls > 2*cli.Metrics.ToolCalls {
			moreThanDoubleCLI = append(moreThanDoubleCLI, scenario)
		}
	}
	if len(moreThanDoubleCLI) >= 3 {
		triggers = append(triggers, fmt.Sprintf("production used more than 2x CLI tools in %s", sortedJoin(moreThanDoubleCLI)))
	}

	recommendation := "continue_production_hardening"
	if len(triggers) > 0 {
		recommendation = "continue_cli_baseline_for_agent_operations"
	}
	return &stopLossSummary{
		Policy:         "After production hardening, keep CLI as the baseline recommendation if production loses correctness, directly inspects generated files, inspects the module cache, uses broad repo search in more than one routine scenario, uses the openhealth CLI, uses direct SQLite access, exceeds core tool thresholds, or uses more than 2x CLI tools in at least three comparable scenarios.",
		Triggered:      len(triggers) > 0,
		Recommendation: recommendation,
		Triggers:       triggers,
	}
}

func codeFirstSummaryFor(results []runResult, candidateVariant string) *codeFirstSummary {
	const baselineVariant = "cli"

	if strings.TrimSpace(candidateVariant) == "" {
		candidateVariant = "production"
	}

	candidateByScenario := map[string]runResult{}
	cliByScenario := map[string]runResult{}
	for _, result := range results {
		switch result.Variant {
		case candidateVariant:
			candidateByScenario[result.Scenario] = result
		case baselineVariant:
			cliByScenario[result.Scenario] = result
		}
	}
	if len(candidateByScenario) == 0 {
		return nil
	}

	entries := []codeFirstComparisonRow{}
	candidatePassedAll := true
	noGenerated := true
	noModuleCache := true
	noRoutineBroadSearch := true
	noCLIUsage := true
	noDirectSQLite := true
	validationFinalAnswerOnly := true
	validationFinalAnswerFailures := []string{}
	totalCandidateTools := 0
	totalCLITools := 0
	totalCandidateNonCached := 0
	totalCLINonCached := 0
	tokenTotalComparable := true
	tokenMajorityWins := 0
	tokenMajorityScenarios := 0
	tokenMissing := []string{}
	scenariosAtOrBelowCLI := 0
	routineExceedsCLIByMoreThanOne := []string{}
	missingCLI := []string{}

	candidateScenarioIDs := []string{}
	for _, scenario := range scenarioIDs() {
		candidate, ok := candidateByScenario[scenario]
		if !ok {
			continue
		}
		candidateScenarioIDs = append(candidateScenarioIDs, scenario)
		cli, hasCLI := cliByScenario[scenario]
		if !hasCLI {
			missingCLI = append(missingCLI, scenario)
		}
		if !candidate.Passed {
			candidatePassedAll = false
		}
		if candidate.Metrics.GeneratedFileInspected {
			noGenerated = false
		}
		if candidate.Metrics.ModuleCacheInspected {
			noModuleCache = false
		}
		if candidate.Metrics.BroadRepoSearch && isRoutineScenario(candidate.Scenario) {
			noRoutineBroadSearch = false
		}
		if candidate.Metrics.CLIUsed {
			noCLIUsage = false
		}
		if candidate.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
		}
		if isFinalAnswerOnlyValidationScenario(candidate.Scenario) &&
			(candidate.Metrics.ToolCalls != 0 || candidate.Metrics.CommandExecutions != 0 || candidate.Metrics.AssistantCalls > 1) {
			validationFinalAnswerOnly = false
			validationFinalAnswerFailures = append(validationFinalAnswerFailures, candidate.Scenario)
		}
		totalCandidateTools += candidate.Metrics.ToolCalls
		row := codeFirstComparisonRow{
			Scenario:       scenario,
			CandidatePass:  candidate.Passed,
			CandidateTools: candidate.Metrics.ToolCalls,
		}
		if hasCLI {
			totalCLITools += cli.Metrics.ToolCalls
			row.CLIPass = cli.Passed
			row.CLITools = cli.Metrics.ToolCalls
			delta := candidate.Metrics.ToolCalls - cli.Metrics.ToolCalls
			row.ToolDelta = &delta
			if candidate.Metrics.ToolCalls <= cli.Metrics.ToolCalls {
				scenariosAtOrBelowCLI++
			}
			tokenMajorityScenarios++
			candidateTokens, candidateHasTokens := nonCachedTokens(candidate)
			cliTokens, cliHasTokens := nonCachedTokens(cli)
			if !candidateHasTokens || !cliHasTokens {
				tokenTotalComparable = false
				tokenMissing = append(tokenMissing, scenario)
			} else {
				totalCandidateNonCached += candidateTokens
				totalCLINonCached += cliTokens
				if candidateTokens < cliTokens {
					tokenMajorityWins++
				}
			}
			if isRoutineScenario(scenario) && candidate.Metrics.ToolCalls > cli.Metrics.ToolCalls+1 {
				routineExceedsCLIByMoreThanOne = append(routineExceedsCLIByMoreThanOne, scenario)
			}
		}
		entries = append(entries, row)
	}
	requiredAtOrBelowCLI := requiredScenariosAtOrBelowCLI(len(candidateScenarioIDs))
	requiredTokenWins := strictMajority(tokenMajorityScenarios)

	criteria := []codeFirstCriterion{
		{
			Name:    "candidate_passes_all_scenarios",
			Passed:  candidatePassedAll,
			Details: fmt.Sprintf("%d/%d candidate scenarios passed", countPassed(candidateByScenario), len(candidateScenarioIDs)),
		},
		{
			Name:    "no_direct_generated_file_inspection",
			Passed:  noGenerated,
			Details: fmt.Sprintf("%s must not directly inspect generated files", candidateVariant),
		},
		{
			Name:    "no_module_cache_inspection",
			Passed:  noModuleCache,
			Details: fmt.Sprintf("%s must not inspect the Go module cache", candidateVariant),
		},
		{
			Name:    "no_routine_broad_repo_search",
			Passed:  noRoutineBroadSearch,
			Details: fmt.Sprintf("%s must not use broad repo search in routine scenarios", candidateVariant),
		},
		{
			Name:    "no_openhealth_cli_usage",
			Passed:  noCLIUsage,
			Details: fmt.Sprintf("%s must not use the openhealth CLI", candidateVariant),
		},
		{
			Name:    "no_direct_sqlite_access",
			Passed:  noDirectSQLite,
			Details: fmt.Sprintf("%s must not use direct SQLite access", candidateVariant),
		},
		{
			Name:    "validation_scenarios_are_final_answer_only",
			Passed:  validationFinalAnswerOnly,
			Details: validationFinalAnswerDetails(validationFinalAnswerFailures),
		},
		{
			Name:    "total_tools_less_than_or_equal_cli",
			Passed:  len(missingCLI) == 0 && totalCandidateTools <= totalCLITools,
			Details: fmt.Sprintf("%s tools %d vs cli tools %d", candidateVariant, totalCandidateTools, totalCLITools),
		},
		{
			Name:    "minimum_scenarios_at_or_below_cli",
			Passed:  scenariosAtOrBelowCLI >= requiredAtOrBelowCLI,
			Details: fmt.Sprintf("%d scenarios at or below CLI tools; required %d of %d", scenariosAtOrBelowCLI, requiredAtOrBelowCLI, len(candidateScenarioIDs)),
		},
		{
			Name:    "no_routine_scenario_exceeds_cli_by_more_than_one_tool",
			Passed:  len(missingCLI) == 0 && len(routineExceedsCLIByMoreThanOne) == 0,
			Details: missingAwareDetails(missingCLI, routineExceedsCLIByMoreThanOne),
		},
		{
			Name:    "non_cached_token_majority",
			Passed:  len(missingCLI) == 0 && tokenMajorityWins >= requiredTokenWins,
			Details: fmt.Sprintf("%d scenarios with lower non-cached input tokens; required %d of %d; missing usage: %s", tokenMajorityWins, requiredTokenWins, tokenMajorityScenarios, missingTokenDetails(tokenMissing)),
		},
		{
			Name:    "non_cached_token_total_less_than_or_equal_cli",
			Passed:  len(missingCLI) == 0 && tokenTotalComparable && totalCandidateNonCached <= totalCLINonCached,
			Details: fmt.Sprintf("%s non-cached input tokens %d vs cli %d; missing usage: %s", candidateVariant, totalCandidateNonCached, totalCLINonCached, missingTokenDetails(tokenMissing)),
		},
	}

	beatsCLI := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			beatsCLI = false
			break
		}
	}
	recommendation := "continue_cli_for_routine_openhealth_operations"
	if beatsCLI {
		recommendation = recommendationForCandidate(candidateVariant)
	}

	return &codeFirstSummary{
		CandidateVariant: candidateVariant,
		BaselineVariant:  baselineVariant,
		BeatsCLI:         beatsCLI,
		Recommendation:   recommendation,
		Criteria:         criteria,
		Entries:          entries,
	}
}

func missingAwareDetails(missingCLI []string, exceeded []string) string {
	if len(missingCLI) > 0 {
		return fmt.Sprintf("missing cli scenarios: %s", sortedJoin(missingCLI))
	}
	if len(exceeded) > 0 {
		return fmt.Sprintf("routine scenarios over CLI by more than one tool: %s", sortedJoin(exceeded))
	}
	return "no routine scenario exceeded CLI by more than one tool"
}

func countPassed(results map[string]runResult) int {
	count := 0
	for _, result := range results {
		if result.Passed {
			count++
		}
	}
	return count
}

func requiredScenariosAtOrBelowCLI(candidateScenarios int) int {
	return int(math.Ceil(float64(candidateScenarios) * 0.8))
}

func strictMajority(count int) int {
	return count/2 + 1
}

func nonCachedTokens(result runResult) (int, bool) {
	if result.Metrics.NonCachedInputTokens == nil {
		return 0, false
	}
	return *result.Metrics.NonCachedInputTokens, true
}

func isFinalAnswerOnlyValidationScenario(id string) bool {
	switch id {
	case "ambiguous-short-date", "invalid-input", "non-iso-date-reject",
		"bp-invalid-input", "bp-non-iso-date-reject", "mixed-invalid-direct-reject":
		return true
	default:
		return false
	}
}

func validationFinalAnswerDetails(failures []string) string {
	if len(failures) == 0 {
		return "validation scenarios used no tools, no command executions, and at most one assistant answer"
	}
	return fmt.Sprintf("validation scenarios were not final-answer-only: %s", sortedJoin(failures))
}

func missingTokenDetails(missing []string) string {
	if len(missing) == 0 {
		return "none"
	}
	return sortedJoin(missing)
}

func recommendationForCandidate(candidateVariant string) string {
	if candidateVariant == "production" {
		return "prefer_agentops_production_for_routine_openhealth_operations"
	}
	return "prefer_agentops_for_routine_openhealth_operations"
}

func scenarioIDs() []string {
	ids := []string{}
	for _, scenario := range scenarios() {
		ids = append(ids, scenario.ID)
	}
	return ids
}

func isRoutineScenario(id string) bool {
	switch id {
	case "add-two", "repeat-add", "update-existing", "bounded-range", "bounded-range-natural", "latest-only", "history-limit-two",
		"bp-add-two", "bp-latest-only", "bp-history-limit-two", "bp-bounded-range", "bp-bounded-range-natural",
		"bp-correct-existing", "bp-correct-missing-reject", "bp-correct-ambiguous-reject",
		"mixed-add-latest", "mixed-bounded-range", "mt-bp-latest-then-correct", "mt-mixed-latest-then-correct":
		return true
	default:
		return false
	}
}

func isMixedScenario(id string) bool {
	return strings.HasPrefix(id, "mixed-")
}

func isBloodPressureScenario(id string) bool {
	return strings.HasPrefix(id, "bp-")
}

func sortedJoin(values []string) string {
	values = append([]string{}, values...)
	sort.Strings(values)
	return strings.Join(values, ", ")
}

func resultKey(variant string, scenario string) string {
	return variant + "/" + scenario
}

func resultDelta(was bool, now bool) string {
	switch {
	case was && now:
		return "same_pass"
	case !was && !now:
		return "same_fail"
	case was && !now:
		return "regressed"
	default:
		return "fixed"
	}
}

func intDelta(was int, now int) *int {
	delta := now - was
	return &delta
}

func optionalIntDelta(was *int, now *int) *int {
	if was == nil || now == nil {
		return nil
	}
	return intDelta(*was, *now)
}

func floatDelta(was float64, now float64) *float64 {
	delta := roundSeconds(now - was)
	return &delta
}

func boolChange(was bool, now bool) string {
	switch {
	case was == now && now:
		return "same_yes"
	case was == now && !now:
		return "same_no"
	case was && !now:
		return "improved_to_no"
	default:
		return "regressed_to_yes"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func repoRelativePath(repoRoot string, path string) string {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func writeMarkdown(path string, value report) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# oh-5yr Agent Eval Results\n\n")
	fmt.Fprintf(&b, "Date: %s\n\n", value.Date)
	fmt.Fprintf(&b, "Harness: `%s`\n\n", value.Harness)
	fmt.Fprintf(&b, "Model: `%s`, reasoning effort `%s`\n\n", value.Model, value.ReasoningEffort)
	fmt.Fprintf(&b, "Codex CLI: `%s`\n\n", value.CodexVersion)
	if value.Parallelism > 0 {
		fmt.Fprintf(&b, "Parallelism: `%d`\n\n", value.Parallelism)
	}
	if value.CacheMode != "" {
		fmt.Fprintf(&b, "Cache mode: `%s`\n\n", value.CacheMode)
	}
	if value.CachePrewarmSeconds > 0 {
		fmt.Fprintf(&b, "Cache prewarm seconds: `%.2f`\n\n", value.CachePrewarmSeconds)
	}
	if value.HarnessElapsedSeconds > 0 {
		fmt.Fprintf(&b, "Harness elapsed seconds: `%.2f`\n\n", value.HarnessElapsedSeconds)
	}
	if value.EffectiveSpeedup > 0 {
		fmt.Fprintf(&b, "Effective parallel speedup: `%.2fx`\n\n", value.EffectiveSpeedup)
	}
	if value.ParallelEfficiency > 0 {
		fmt.Fprintf(&b, "Parallel efficiency: `%.2f`\n\n", value.ParallelEfficiency)
	}
	fmt.Fprintf(&b, "Reduced JSON artifact: `docs/agent-eval-results/%s-%s.json`\n\n", value.Issue, value.Date)
	fmt.Fprintf(&b, "Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.\n\n")

	fmt.Fprintf(&b, "## History Isolation\n\n")
	fmt.Fprintf(&b, "- Status: `%s`\n", value.HistoryIsolation.Status)
	fmt.Fprintf(&b, "- Single-turn ephemeral runs: `%d`.\n", value.HistoryIsolation.SingleTurnEphemeralRuns)
	fmt.Fprintf(&b, "- Multi-turn persisted sessions: `%d` sessions / `%d` turns.\n", value.HistoryIsolation.MultiTurnPersistedSessions, value.HistoryIsolation.MultiTurnPersistedTurns)
	fmt.Fprintf(&b, "- The Codex desktop app was not used for eval prompts.\n")
	fmt.Fprintf(&b, "- New Codex session files referencing `<run-root>`: `%d`.\n", value.HistoryIsolation.NewSessionFilesAfterRun)
	if value.HistoryIsolation.VerificationLimitation != "" {
		fmt.Fprintf(&b, "- Limitation: %s\n", value.HistoryIsolation.VerificationLimitation)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "## Results\n\n")
	fmt.Fprintf(&b, "| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, result := range value.Results {
		tokenSummary := "not_exposed"
		if result.Metrics.UsageExposed {
			tokenSummary = fmt.Sprintf("in %d / cached %d / out %d", deref(result.Metrics.InputTokens), deref(result.Metrics.CachedInputTokens), deref(result.Metrics.OutputTokens))
		}
		fmt.Fprintf(
			&b,
			"| `%s` | `%s` | %s | %s | %s | %d | %d | %.2f | %s | %s | %s | %s | %s | %s | %s |\n",
			result.Variant,
			result.Scenario,
			passText(result.Passed),
			passText(result.Verification.DatabasePass),
			passText(result.Verification.AssistantPass),
			result.Metrics.ToolCalls,
			result.Metrics.AssistantCalls,
			result.WallSeconds,
			tokenSummary,
			yesNo(result.Metrics.GeneratedFileInspected),
			yesNo(result.Metrics.BroadRepoSearch),
			yesNo(result.Metrics.GeneratedPathFromBroadSearch),
			yesNo(result.Metrics.ModuleCacheInspected),
			yesNo(result.Metrics.CLIUsed),
			yesNo(result.Metrics.DirectSQLiteAccess),
		)
	}

	fmt.Fprintf(&b, "\n## Phase Timings\n\n")
	fmt.Fprintf(&b, "| Phase | Seconds |\n")
	fmt.Fprintf(&b, "| --- | ---: |\n")
	fmt.Fprintf(&b, "| prepare_run_dir | %.2f |\n", value.PhaseTotals.PrepareRunDir)
	fmt.Fprintf(&b, "| copy_repo | %.2f |\n", value.PhaseTotals.CopyRepo)
	fmt.Fprintf(&b, "| install_variant | %.2f |\n", value.PhaseTotals.InstallVariant)
	fmt.Fprintf(&b, "| warm_cache | %.2f |\n", value.PhaseTotals.WarmCache)
	fmt.Fprintf(&b, "| seed_db | %.2f |\n", value.PhaseTotals.SeedDB)
	fmt.Fprintf(&b, "| agent_run | %.2f |\n", value.PhaseTotals.AgentRun)
	fmt.Fprintf(&b, "| parse_metrics | %.2f |\n", value.PhaseTotals.ParseMetrics)
	fmt.Fprintf(&b, "| verify | %.2f |\n", value.PhaseTotals.Verify)
	fmt.Fprintf(&b, "| total_job_time | %.2f |\n", value.PhaseTotals.Total)

	if value.Comparison != nil {
		fmt.Fprintf(&b, "\n## Comparison\n\n")
		fmt.Fprintf(&b, "Baseline: `%s`.\n\n", value.Comparison.BaselineReport)
		fmt.Fprintf(&b, "| Variant | Scenario | Result | Tools Δ | Assistant Calls Δ | Wall Seconds Δ | Non-cache Tokens Δ | Direct Generated Files | Broad Search | Generated From Broad | Module Cache | CLI Used | Direct SQLite |\n")
		fmt.Fprintf(&b, "| --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |\n")
		for _, entry := range value.Comparison.Entries {
			fmt.Fprintf(
				&b,
				"| `%s` | `%s` | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				entry.Variant,
				entry.Scenario,
				entry.Result,
				formatIntDelta(entry.ToolCallsDelta),
				formatIntDelta(entry.AssistantCallsDelta),
				formatFloatDelta(entry.WallSecondsDelta),
				formatIntDelta(entry.NonCachedInputTokensDelta),
				entry.GeneratedFileInspectionChange,
				entry.BroadRepoSearchChange,
				entry.GeneratedPathFromBroadSearchChange,
				entry.ModuleCacheInspectionChange,
				entry.CLIUsageChange,
				entry.DirectSQLiteAccessChange,
			)
		}
	}

	if len(value.MetricNotes) > 0 {
		fmt.Fprintf(&b, "\n## Metric Notes\n\n")
		for _, note := range value.MetricNotes {
			fmt.Fprintf(&b, "- %s\n", note)
		}
	}

	if value.StopLoss != nil {
		fmt.Fprintf(&b, "\n## Production Stop-Loss\n\n")
		fmt.Fprintf(&b, "- Triggered: `%s`\n", yesNo(value.StopLoss.Triggered))
		fmt.Fprintf(&b, "- Recommendation: `%s`\n", value.StopLoss.Recommendation)
		if len(value.StopLoss.Triggers) > 0 {
			for _, trigger := range value.StopLoss.Triggers {
				fmt.Fprintf(&b, "- Trigger: %s\n", trigger)
			}
		}
	}

	if value.CodeFirst != nil {
		fmt.Fprintf(&b, "\n## Code-First CLI Comparison\n\n")
		fmt.Fprintf(&b, "- Candidate: `%s`\n", value.CodeFirst.CandidateVariant)
		fmt.Fprintf(&b, "- Baseline: `%s`\n", value.CodeFirst.BaselineVariant)
		fmt.Fprintf(&b, "- Beats CLI: `%s`\n", yesNo(value.CodeFirst.BeatsCLI))
		fmt.Fprintf(&b, "- Recommendation: `%s`\n\n", value.CodeFirst.Recommendation)
		fmt.Fprintf(&b, "| Criterion | Result | Details |\n")
		fmt.Fprintf(&b, "| --- | --- | --- |\n")
		for _, criterion := range value.CodeFirst.Criteria {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", criterion.Name, passText(criterion.Passed), criterion.Details)
		}
		fmt.Fprintf(&b, "\n| Scenario | Candidate | CLI | Tools Δ |\n")
		fmt.Fprintf(&b, "| --- | ---: | ---: | ---: |\n")
		for _, entry := range value.CodeFirst.Entries {
			fmt.Fprintf(&b, "| `%s` | %d | %d | %s |\n", entry.Scenario, entry.CandidateTools, entry.CLITools, formatIntDelta(entry.ToolDelta))
		}
	}

	if hasMetricEvidence(value.Results) {
		fmt.Fprintf(&b, "\n## Metric Evidence\n\n")
		for _, result := range value.Results {
			writeMetricEvidence(&b, result)
		}
	}

	if hasMultiTurnResults(value.Results) {
		fmt.Fprintf(&b, "\n## Turn Details\n\n")
		for _, result := range value.Results {
			if len(result.Turns) <= 1 {
				continue
			}
			for _, turn := range result.Turns {
				fmt.Fprintf(&b, "- `%s/%s` turn %d: exit `%d`, tools `%d`, assistant calls `%d`, wall `%.2f`, raw `%s`.\n", result.Variant, result.Scenario, turn.Index, turn.ExitCode, turn.Metrics.ToolCalls, turn.Metrics.AssistantCalls, turn.WallSeconds, turn.RawLogArtifactReference)
			}
		}
	}

	fmt.Fprintf(&b, "\n## Scenario Notes\n\n")
	for _, result := range value.Results {
		fmt.Fprintf(&b, "- `%s/%s`: %s Raw event reference: `%s`.\n", result.Variant, result.Scenario, result.Verification.Details, result.RawLogArtifactReference)
	}

	fmt.Fprintf(&b, "\n## CLI-Oriented Variant\n\n")
	fmt.Fprintf(&b, "Status: `%s`.\n", value.CLIStatus)

	fmt.Fprintf(&b, "\n## App Server\n\n")
	fmt.Fprintf(&b, "%s.\n", value.AppServerFallback)

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func hasMetricEvidence(results []runResult) bool {
	for _, result := range results {
		if len(result.Metrics.GeneratedFileEvidence) > 0 ||
			len(result.Metrics.GeneratedPathFromBroadSearchEvidence) > 0 ||
			len(result.Metrics.BroadRepoSearchEvidence) > 0 ||
			len(result.Metrics.ModuleCacheEvidence) > 0 ||
			len(result.Metrics.CLIUsageEvidence) > 0 ||
			len(result.Metrics.DirectSQLiteEvidence) > 0 {
			return true
		}
	}
	return false
}

func hasMultiTurnResults(results []runResult) bool {
	for _, result := range results {
		if len(result.Turns) > 1 {
			return true
		}
	}
	return false
}

func writeMetricEvidence(b *strings.Builder, result runResult) {
	if len(result.Metrics.GeneratedFileEvidence) == 0 &&
		len(result.Metrics.GeneratedPathFromBroadSearchEvidence) == 0 &&
		len(result.Metrics.BroadRepoSearchEvidence) == 0 &&
		len(result.Metrics.ModuleCacheEvidence) == 0 &&
		len(result.Metrics.CLIUsageEvidence) == 0 &&
		len(result.Metrics.DirectSQLiteEvidence) == 0 {
		return
	}
	prefix := fmt.Sprintf("- `%s/%s`", result.Variant, result.Scenario)
	if len(result.Metrics.GeneratedFileEvidence) > 0 {
		fmt.Fprintf(b, "%s direct generated-file: `%s`.\n", prefix, strings.Join(result.Metrics.GeneratedFileEvidence, "`; `"))
	}
	if len(result.Metrics.BroadRepoSearchEvidence) > 0 {
		fmt.Fprintf(b, "%s broad repo search: `%s`.\n", prefix, strings.Join(result.Metrics.BroadRepoSearchEvidence, "`; `"))
	}
	if len(result.Metrics.GeneratedPathFromBroadSearchEvidence) > 0 {
		fmt.Fprintf(b, "%s generated path from broad search: `%s`.\n", prefix, strings.Join(result.Metrics.GeneratedPathFromBroadSearchEvidence, "`; `"))
	}
	if len(result.Metrics.ModuleCacheEvidence) > 0 {
		fmt.Fprintf(b, "%s module cache: `%s`.\n", prefix, strings.Join(result.Metrics.ModuleCacheEvidence, "`; `"))
	}
	if len(result.Metrics.CLIUsageEvidence) > 0 {
		fmt.Fprintf(b, "%s openhealth CLI: `%s`.\n", prefix, strings.Join(result.Metrics.CLIUsageEvidence, "`; `"))
	}
	if len(result.Metrics.DirectSQLiteEvidence) > 0 {
		fmt.Fprintf(b, "%s direct SQLite: `%s`.\n", prefix, strings.Join(result.Metrics.DirectSQLiteEvidence, "`; `"))
	}
}

func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return filepath.Abs(strings.TrimSpace(string(out)))
}

func commandOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(out))
}

func countNewSessionFiles(marker time.Time, runRoot string) int {
	home, err := os.UserHomeDir()
	if err != nil {
		return -1
	}
	sessionsDir := filepath.Join(home, ".codex", "sessions")
	count := 0
	_ = filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		if info.ModTime().After(marker) && fileContains(path, runRoot) {
			count++
		}
		return nil
	})
	return count
}

func fileContains(path string, needle string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			return true
		}
	}
	return false
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func isWithin(path string, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC), nil
}

func intPointer(value int) *int {
	return &value
}

func equalIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func weightsEqual(got []weightState, want []weightState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date || got[i].Unit != want[i].Unit || math.Abs(got[i].Value-want[i].Value) > 0.001 {
			return false
		}
	}
	return true
}

func bloodPressuresEqual(got []bloodPressureState, want []bloodPressureState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Systolic != want[i].Systolic ||
			got[i].Diastolic != want[i].Diastolic ||
			!equalIntPointer(got[i].Pulse, want[i].Pulse) {
			return false
		}
	}
	return true
}

func describeWeights(weights []weightState) string {
	if len(weights) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(weights))
	for _, weight := range weights {
		parts = append(parts, fmt.Sprintf("%s %.1f %s", weight.Date, weight.Value, weight.Unit))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeBloodPressures(readings []bloodPressureState) string {
	if len(readings) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(readings))
	for _, reading := range readings {
		pulse := ""
		if reading.Pulse != nil {
			pulse = fmt.Sprintf(" pulse %d", *reading.Pulse)
		}
		parts = append(parts, fmt.Sprintf("%s %d/%d%s", reading.Date, reading.Systolic, reading.Diastolic, pulse))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func containsAll(value string, needles []string) bool {
	normalized := strings.ToLower(value)
	for _, needle := range needles {
		if !strings.Contains(normalized, strings.ToLower(needle)) {
			return false
		}
	}
	return true
}

func mentionsDate(message string, date string) bool {
	return dateMentionIndex(message, date) >= 0
}

func boundedRangeAssistantPass(message string) bool {
	previous := -1
	for _, date := range []string{"2026-03-30", "2026-03-29"} {
		index := includedDateResultLineIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedDate(message, "2026-03-28")
}

func latestOnlyAssistantPass(message string) bool {
	return mentionsIncludedDate(message, "2026-03-30") &&
		!mentionsIncludedDate(message, "2026-03-29") &&
		!mentionsIncludedDate(message, "2026-03-28")
}

func historyLimitTwoAssistantPass(message string) bool {
	previous := -1
	for _, date := range []string{"2026-03-30", "2026-03-29"} {
		index := includedDateResultLineIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedDate(message, "2026-03-28") &&
		!mentionsIncludedDate(message, "2026-03-27")
}

func bloodPressureBoundedRangeAssistantPass(message string) bool {
	previous := -1
	for _, expected := range []bloodPressureState{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	} {
		index := includedBloodPressureResultLineIndex(message, expected)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func bloodPressureLatestOnlyAssistantPass(message string) bool {
	return includedBloodPressureResultLineIndex(message, bloodPressureState{Date: "2026-03-30", Systolic: 118, Diastolic: 76}) >= 0 &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-29") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func bloodPressureHistoryLimitTwoAssistantPass(message string) bool {
	previous := -1
	for _, expected := range []bloodPressureState{
		{Date: "2026-03-30", Systolic: 118, Diastolic: 76},
		{Date: "2026-03-29", Systolic: 122, Diastolic: 78},
	} {
		index := includedBloodPressureResultLineIndex(message, expected)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return !mentionsIncludedBloodPressureDate(message, "2026-03-28") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-27")
}

func mixedLatestAssistantPass(message string) bool {
	return mentionsIncludedDate(message, "2026-03-31") &&
		containsAll(message, []string{"150.8", "119/77"}) &&
		containsAny(strings.ToLower(message), []string{"pulse 62", "62"})
}

func mixedLatestSeedAssistantPass(message string) bool {
	return mentionsIncludedDate(message, "2026-03-30") &&
		containsAll(message, []string{"151.6", "118/76"}) &&
		!mentionsIncludedDate(message, "2026-03-29")
}

func mixedBoundedRangeAssistantPass(message string) bool {
	return boundedRangeAssistantPass(message) &&
		bloodPressureBoundedRangeAssistantPass(message) &&
		containsAll(message, []string{"151.6", "152.2"}) &&
		!mentionsIncludedDate(message, "2026-03-28") &&
		!mentionsIncludedBloodPressureDate(message, "2026-03-28")
}

func nonISODateRejectAssistantPass(message string) bool {
	return containsAny(strings.ToLower(message), []string{"yyyy-mm-dd", "iso", "invalid", "cannot", "can't", "reject", "unsupported", "format"})
}

func mentionsIncludedDate(message string, date string) bool {
	return includedDateResultLineIndex(message, date) >= 0
}

func includedDateResultLineIndex(message string, date string) int {
	offset := 0
	for _, line := range strings.SplitAfter(message, "\n") {
		dateIndex := dateMentionIndex(line, date)
		if dateIndex < 0 {
			offset += len(line)
			continue
		}
		if lineMentionsExclusion(line) {
			offset += len(line)
			continue
		}
		if lineLooksLikeResult(line) {
			return offset + dateIndex
		}
		offset += len(line)
	}
	return -1
}

func includedBloodPressureResultLineIndex(message string, expected bloodPressureState) int {
	offset := 0
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		dateIndex := dateMentionIndex(line, expected.Date)
		if dateIndex < 0 {
			offset += len(line)
			continue
		}
		if lineMentionsExclusion(line) {
			offset += len(line)
			continue
		}
		if lineMentionsBloodPressure(line, expected.Systolic, expected.Diastolic) {
			return offset + dateIndex
		}
		if followingBloodPressureResultLine(lines, i, func(candidate string) bool {
			return lineMentionsBloodPressure(candidate, expected.Systolic, expected.Diastolic)
		}) {
			return offset + dateIndex
		}
		offset += len(line)
	}
	return -1
}

func mentionsIncludedBloodPressureDate(message string, date string) bool {
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		if dateMentionIndex(line, date) >= 0 && !lineMentionsExclusion(line) {
			if lineLooksLikeBloodPressureResult(line) {
				return true
			}
			if followingBloodPressureResultLine(lines, i, lineLooksLikeBloodPressureResult) {
				return true
			}
		}
	}
	return false
}

func followingBloodPressureResultLine(lines []string, start int, matches func(string) bool) bool {
	for i := start + 1; i < len(lines) && i <= start+3; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if lineMentionsExclusion(line) {
			return false
		}
		return matches(line)
	}
	return false
}

func lineMentionsBloodPressure(line string, systolic int, diastolic int) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, fmt.Sprintf("%d/%d", systolic, diastolic)) ||
		(strings.Contains(lower, fmt.Sprintf("%d", systolic)) && strings.Contains(lower, fmt.Sprintf("%d", diastolic)))
}

func lineLooksLikeBloodPressureResult(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return strings.Contains(lower, "/") ||
		strings.Contains(lower, "systolic") ||
		strings.Contains(lower, "diastolic") ||
		strings.Contains(lower, "blood pressure")
}

func lineMentionsExclusion(line string) bool {
	lower := strings.ToLower(line)
	return containsAny(lower, []string{"no entries", "not included", "not include", "excluded", "outside", "do not include"})
}

func lineLooksLikeResult(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
		return true
	}
	if len(trimmed) >= 2 && unicode.IsDigit(rune(trimmed[0])) && (trimmed[1] == '.' || trimmed[1] == ')') {
		return true
	}
	lower := strings.ToLower(trimmed)
	return strings.Contains(lower, " lb") ||
		(strings.Contains(lower, `"date"`) &&
			strings.Contains(lower, `"value"`) &&
			(strings.Contains(lower, `"unit":"lb"`) || strings.Contains(lower, `"unit": "lb"`)))
}

func mentionsDatesInOrder(message string, dates ...string) bool {
	previous := -1
	for _, date := range dates {
		index := dateMentionIndex(message, date)
		if index < 0 || index <= previous {
			return false
		}
		previous = index
	}
	return true
}

func dateMentionIndex(message string, date string) int {
	lower := strings.ToLower(message)
	best := -1
	for _, pattern := range dateMentionPatterns(date) {
		index := strings.Index(lower, pattern)
		if index >= 0 && (best < 0 || index < best) {
			best = index
		}
	}
	return best
}

func dateMentionPatterns(date string) []string {
	patterns := []string{strings.ToLower(date)}
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return patterns
	}
	return append(patterns,
		strings.ToLower(parsed.Format("01/02")),
		strings.ToLower(parsed.Format("January 2, 2006")),
		strings.ToLower(parsed.Format("January 2")),
		strings.ToLower(parsed.Format("Jan 2")),
	)
}

func promptSummary(sc scenario) string {
	switch sc.ID {
	case "add-two":
		return "add 2026-03-29 152.2 lb and 2026-03-30 151.6 lb"
	case "repeat-add":
		return "reapply the same two weights against preseeded rows"
	case "update-existing":
		return "correct 2026-03-29 from 152.2 lb to 151.6 lb"
	case "bounded-range":
		return "list only 2026-03-29 through 2026-03-30 from preseeded rows"
	case "bounded-range-natural":
		return "naturally ask for 2026-03-29 and 2026-03-30 from preseeded rows"
	case "latest-only":
		return "list only the latest row from preseeded rows"
	case "history-limit-two":
		return "list only the two most recent rows from preseeded rows"
	case "ambiguous-short-date":
		return "ask for year before writing 03/29 152.2 lb"
	case "invalid-input":
		return "reject -5 stone for 2026-03-31"
	case "non-iso-date-reject":
		return "reject non-ISO date 2026/03/31"
	case "bp-add-two":
		return "record 2026-03-29 122/78 pulse 64 and 2026-03-30 118/76"
	case "bp-latest-only":
		return "list only the latest blood-pressure row from preseeded rows"
	case "bp-history-limit-two":
		return "list only the two most recent blood-pressure rows from preseeded rows"
	case "bp-bounded-range":
		return "list only 2026-03-29 through 2026-03-30 blood-pressure rows from preseeded rows"
	case "bp-bounded-range-natural":
		return "naturally ask for 2026-03-29 and 2026-03-30 blood-pressure rows from preseeded rows"
	case "bp-invalid-input":
		return "reject invalid 0/-5 pulse 0 blood-pressure reading"
	case "bp-non-iso-date-reject":
		return "reject non-ISO blood-pressure date 2026/03/31"
	case "bp-correct-existing":
		return "correct 2026-03-29 blood pressure from 122/78 pulse 64 to 121/77 pulse 63"
	case "bp-correct-missing-reject":
		return "reject correction for missing 2026-03-31 blood-pressure reading without creating one"
	case "bp-correct-ambiguous-reject":
		return "reject ambiguous correction when multiple 2026-03-29 blood-pressure rows exist"
	case "mixed-add-latest":
		return "record one weight and one blood-pressure reading, then report latest for both"
	case "mixed-bounded-range":
		return "list only 2026-03-29 through 2026-03-30 for both domains"
	case "mixed-invalid-direct-reject":
		return "reject invalid mixed weight and blood-pressure values without tools"
	case "mt-weight-clarify-then-add":
		return "ask for missing year, then add 2026-03-29 152.2 lb in a resumed turn"
	case "mt-bp-latest-then-correct":
		return "read latest blood pressure, then correct it in a resumed turn"
	case "mt-mixed-latest-then-correct":
		return "read latest weight and blood pressure, then correct both in a resumed turn"
	default:
		return sc.Title
	}
}

func isFileInspectionCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "sed", "cat", "find", "ls", "awk", "head", "tail", "nl")
}

func inspectsGeneratedFileCommand(command string, output string) bool {
	if !isFileInspectionCommand(command) || isBroadRepoSearchCommand(command) {
		return false
	}
	if onlyAgentSkillTargets(commandPathTargets(commandFields(command))) {
		return false
	}
	if mentionsGeneratedPath(command) {
		return true
	}
	return isContentSearchCommand(command) && outputHasGeneratedResultPath(output)
}

func isBroadRepoSearchCommand(command string) bool {
	fields := commandFields(command)
	normalized := " " + strings.Join(fields, " ") + " "
	switch {
	case commandHasExecutable(command, "rg"):
		return rgBroadRepoSearch(fields, normalized)
	case commandHasExecutable(command, "grep"):
		return grepBroadRepoSearch(fields)
	case commandHasExecutable(command, "find"):
		return hasRepoRootTarget(fields)
	default:
		return false
	}
}

func rgBroadRepoSearch(fields []string, normalized string) bool {
	args := commandArgsAfterExecutable(fields, "rg")
	filesMode := strings.Contains(normalized, " --files ")
	targets := rgPathTargets(args, filesMode)
	return len(targets) == 0 || hasRepoRootTarget(targets) || onlyRepoRootTargets(targets)
}

func rgPathTargets(args []string, filesMode bool) []string {
	targets := []string{}
	patternSeen := filesMode
	for i := 0; i < len(args); i++ {
		field := args[i]
		if field == "" {
			continue
		}
		if field == "--" {
			for _, rest := range args[i+1:] {
				if filesMode || patternSeen {
					targets = append(targets, rest)
					continue
				}
				patternSeen = true
			}
			break
		}
		if strings.HasPrefix(field, "--") {
			if longOptionTakesValue(field) && !strings.Contains(field, "=") {
				i++
			}
			continue
		}
		if strings.HasPrefix(field, "-") && field != "-" {
			if shortOptionTakesValue(field) && len(field) == 2 {
				i++
			}
			continue
		}
		if filesMode || patternSeen {
			targets = append(targets, field)
			continue
		}
		patternSeen = true
	}
	return targets
}

func longOptionTakesValue(field string) bool {
	switch field {
	case "--glob", "--iglob", "--type", "--type-not", "--type-add", "--regexp", "--file", "--max-count", "--after-context", "--before-context", "--context", "--colors", "--engine", "--sort", "--sortr", "--path-separator":
		return true
	default:
		return false
	}
}

func shortOptionTakesValue(field string) bool {
	switch field {
	case "-g", "-e", "-f", "-t", "-T", "-m", "-A", "-B", "-C", "-E", "-M":
		return true
	default:
		return false
	}
}

func commandArgsAfterExecutable(fields []string, names ...string) []string {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for i, field := range fields {
		if _, ok := nameSet[field]; ok {
			return fields[i+1:]
		}
		if _, ok := nameSet[filepath.Base(field)]; ok {
			return fields[i+1:]
		}
	}
	return fields
}

func grepBroadRepoSearch(fields []string) bool {
	if !hasRepoRootTarget(fields) {
		return false
	}
	for _, field := range fields {
		if field == "-r" || field == "-R" || strings.Contains(field, "r") && strings.HasPrefix(field, "-") {
			return true
		}
	}
	return false
}

func commandPathTargets(fields []string) []string {
	targets := []string{}
	for _, field := range fields {
		if field == "" || strings.HasPrefix(field, "-") {
			continue
		}
		switch filepath.Base(field) {
		case "rg", "grep", "find", "awk", "zsh", "bash", "sh":
			continue
		}
		switch field {
		case "lc":
			continue
		}
		if field == "." || field == "./" || strings.HasPrefix(field, "./") || strings.HasPrefix(field, "../") || strings.Contains(field, "/") {
			targets = append(targets, field)
		}
	}
	return targets
}

func onlyRepoRootTargets(targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		if target != "." && target != "./" && target != "./." {
			return false
		}
	}
	return true
}

func onlyAgentSkillTargets(targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		cleaned := filepath.ToSlash(strings.Trim(target, `"'`))
		if !strings.HasPrefix(cleaned, ".agents/skills/") && !strings.HasPrefix(cleaned, "skills/openhealth/") && cleaned != "references/weights.md" {
			return false
		}
	}
	return true
}

func hasRepoRootTarget(fields []string) bool {
	for _, field := range fields {
		if field == "." || field == "./" || field == "./." {
			return true
		}
	}
	return false
}

func isContentSearchCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "awk")
}

func mentionsGeneratedPath(text string) bool {
	return strings.Contains(text, "client.gen.go") ||
		strings.Contains(text, "internal/api/generated") ||
		strings.Contains(text, "server.gen.go")
}

func outputHasGeneratedResultPath(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "./")
		if strings.HasPrefix(trimmed, "client/client.gen.go:") ||
			strings.HasPrefix(trimmed, "client.gen.go:") ||
			strings.HasPrefix(trimmed, "internal/api/generated/") ||
			strings.HasPrefix(trimmed, "server.gen.go:") {
			return true
		}
	}
	return false
}

func inspectsModuleCache(command string) bool {
	lower := strings.ToLower(command)
	return strings.Contains(lower, "gomodcache") ||
		strings.Contains(lower, "pkg/mod")
}

func usesOpenHealthCLI(command string) bool {
	for _, segment := range shellCommandSegments(shellCommandPayload(command)) {
		if segmentUsesOpenHealthCLI(segment) {
			return true
		}
	}
	return false
}

func segmentUsesOpenHealthCLI(segment string) bool {
	fields := commandFields(segment)
	if primaryExecutableIs(fields, "rg", "grep", "sed", "awk", "find", "cat", "head", "tail", "nl", "echo", "printf") {
		return false
	}
	for i, field := range fields {
		if filepath.Base(field) == "go" && i+2 < len(fields) && fields[i+1] == "run" {
			for _, candidate := range fields[i+2:] {
				trimmed := strings.Trim(candidate, `"'`)
				if trimmed == "./cmd/openhealth" || trimmed == "cmd/openhealth" || strings.HasSuffix(trimmed, "/cmd/openhealth") {
					return true
				}
			}
		}
		if (field == "openhealth" || strings.HasSuffix(field, "/openhealth")) && i+1 < len(fields) {
			switch fields[i+1] {
			case "weight", "blood-pressure", "migrate", "serve":
				return true
			}
		}
	}
	return false
}

func shellCommandPayload(command string) string {
	lower := strings.ToLower(command)
	for _, marker := range []string{" -lc ", " -c "} {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		return trimShellArgument(strings.TrimSpace(command[index+len(marker):]))
	}
	return command
}

func trimShellArgument(value string) string {
	if len(value) < 2 {
		return value
	}
	quote := value[0]
	if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
		return value[1 : len(value)-1]
	}
	return value
}

func shellCommandSegments(command string) []string {
	segments := []string{}
	start := 0
	var quote rune
	for i, r := range command {
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ';' || r == '|' || r == '&' || r == '\n' {
			if segment := strings.TrimSpace(command[start:i]); segment != "" {
				segments = append(segments, segment)
			}
			start = i + 1
		}
	}
	if segment := strings.TrimSpace(command[start:]); segment != "" {
		segments = append(segments, segment)
	}
	return segments
}

func primaryExecutableIs(fields []string, names ...string) bool {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for _, field := range fields {
		base := filepath.Base(field)
		if base == "zsh" || base == "bash" || base == "sh" || field == "lc" || field == "cd" || strings.Contains(field, "=") {
			continue
		}
		_, ok := nameSet[base]
		return ok
	}
	return false
}

func usesDirectSQLite(command string) bool {
	lower := strings.ToLower(command)
	return commandHasExecutable(command, "sqlite3") ||
		strings.Contains(lower, "import sqlite3") ||
		strings.Contains(lower, "modernc.org/sqlite") ||
		(strings.Contains(lower, "database/sql") && strings.Contains(lower, "sqlite"))
}

func addMetricEvidence(evidence *[]string, command string) {
	const maxEvidence = 5
	sanitized := sanitizeMetricEvidence(command)
	for _, existing := range *evidence {
		if existing == sanitized {
			return
		}
	}
	if len(*evidence) >= maxEvidence {
		return
	}
	*evidence = append(*evidence, sanitized)
}

func sanitizeMetricEvidence(command string) string {
	fields := strings.Fields(command)
	for i, field := range fields {
		if strings.Contains(field, "openhealth-oh") {
			fields[i] = "<run-root>"
		}
	}
	return strings.Join(fields, " ")
}

func commandHasExecutable(command string, names ...string) bool {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for _, field := range commandFields(command) {
		if _, ok := nameSet[field]; ok {
			return true
		}
		if _, ok := nameSet[filepath.Base(field)]; ok {
			return true
		}
	}
	return false
}

func commandFields(command string) []string {
	return strings.FieldsFunc(strings.ToLower(command), func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("'\"`;&|()", r)
	})
}

func normalizedCommandText(command string) string {
	return " " + strings.Join(commandFields(command), " ") + " "
}

func roundWeight(value float64) float64 {
	return math.Round(value*10) / 10
}

func roundSeconds(value float64) float64 {
	return math.Round(value*100) / 100
}

func passText(value bool) string {
	if value {
		return "pass"
	}
	return "fail"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func formatIntDelta(value *int) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+d", *value)
}

func formatFloatDelta(value *float64) string {
	if value == nil {
		return "n/a"
	}
	return fmt.Sprintf("%+.2f", *value)
}

func deref(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
