package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
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
	"slices"
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

var prewarmCompilePackages = []string{"./cmd/openhealth", "./internal/runner"}

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
	MetricNotes           []string                `json:"metric_notes,omitempty"`
	StopLoss              *stopLossSummary        `json:"stop_loss,omitempty"`
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
	BuildAgentApp  float64 `json:"build_agent_app_seconds,omitempty"`
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
	Passed          bool                   `json:"passed"`
	DatabasePass    bool                   `json:"database_pass"`
	AssistantPass   bool                   `json:"assistant_pass"`
	Details         string                 `json:"details"`
	Weights         []weightState          `json:"weights,omitempty"`
	BodyComposition []bodyCompositionState `json:"body_composition,omitempty"`
	BloodPressures  []bloodPressureState   `json:"blood_pressures,omitempty"`
	Sleep           []sleepState           `json:"sleep,omitempty"`
	Medications     []medicationState      `json:"medications,omitempty"`
	Labs            []labCollectionState   `json:"labs,omitempty"`
	Imaging         []imagingState         `json:"imaging,omitempty"`
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
	Note  *string `json:"note,omitempty"`
}

type bloodPressureState struct {
	Date      string  `json:"date"`
	Systolic  int     `json:"systolic"`
	Diastolic int     `json:"diastolic"`
	Pulse     *int    `json:"pulse,omitempty"`
	Note      *string `json:"note,omitempty"`
}

type medicationState struct {
	Name       string  `json:"name"`
	DosageText *string `json:"dosage_text,omitempty"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
	Note       *string `json:"note,omitempty"`
}

type labCollectionState struct {
	Date    string           `json:"date"`
	Note    *string          `json:"note,omitempty"`
	Results []labResultState `json:"results"`
}

type bodyCompositionState struct {
	Date           string   `json:"date"`
	BodyFatPercent *float64 `json:"body_fat_percent,omitempty"`
	WeightValue    *float64 `json:"weight_value,omitempty"`
	WeightUnit     *string  `json:"weight_unit,omitempty"`
	Method         *string  `json:"method,omitempty"`
	Note           *string  `json:"note,omitempty"`
}

type sleepState struct {
	Date         string  `json:"date"`
	QualityScore int     `json:"quality_score"`
	WakeupCount  *int    `json:"wakeup_count,omitempty"`
	Note         *string `json:"note,omitempty"`
}

type imagingState struct {
	Date       string   `json:"date"`
	Modality   string   `json:"modality"`
	BodySite   *string  `json:"body_site,omitempty"`
	Title      *string  `json:"title,omitempty"`
	Summary    string   `json:"summary"`
	Impression *string  `json:"impression,omitempty"`
	Note       *string  `json:"note,omitempty"`
	Notes      []string `json:"notes,omitempty"`
}

type labResultState struct {
	TestName      string   `json:"test_name"`
	CanonicalSlug *string  `json:"canonical_slug,omitempty"`
	ValueText     string   `json:"value_text"`
	ValueNumeric  *float64 `json:"value_numeric,omitempty"`
	Units         *string  `json:"units,omitempty"`
	Notes         []string `json:"notes,omitempty"`
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
	RunRoot        string
	Date           string
	CompareTo      string
	VariantFilter  string
	ScenarioFilter string
	Parallelism    int
	CacheMode      string
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
	if err := timedPhase(&timings.BuildAgentApp, func() error { return buildProductionBinary(runRepo, runDir, dbPath, cache) }); err != nil {
		return runResult{}, fmt.Errorf("build agent app: %w", err)
	}
	if err := preflightEvalContext(repoRoot, runRepo, runDir, cache); err != nil {
		return runResult{}, fmt.Errorf("preflight eval context: %w", err)
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
		BuildAgentApp:  roundSeconds(p.BuildAgentApp),
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
		total.BuildAgentApp += result.PhaseTimings.BuildAgentApp
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
		return "review", "Fewer session files appeared under <run-root>/codex-home than expected for persisted multi-turn eval sessions."
	}
	return "review", "More session files appeared under <run-root>/codex-home than expected for persisted multi-turn eval sessions."
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
		{ID: "production", Title: "Production OpenHealth runner skill"},
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
			ID:     "body-composition-combined-weight-row",
			Title:  "Record combined weight and body-fat import row through two domains",
			Prompt: "Use the configured local OpenHealth data path. Import this row: 03/29/2026 weight 154.2 lb and body-fat percentage 18.7% measured by smart scale. Record the scale weight as weight data and the body-fat value as body-composition data, then tell me what is stored.",
		},
		{
			ID:     "bp-add-two",
			Title:  "Record two blood-pressure readings and verify newest-first output",
			Prompt: "I need to update my local OpenHealth blood pressure history. Use the configured local OpenHealth data path, and use year 2026 for these short dates: 03/29 122/78 pulse 64 with note home cuff, and 03/30 118/76. Then tell me the newest-first entries you see.",
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
			ID:     "bp-invalid-relation",
			Title:  "Reject blood-pressure systolic not greater than diastolic",
			Prompt: "Please add this local OpenHealth blood pressure reading: 03/31/2026 78/78. Do not write it if systolic must be greater than diastolic.",
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
			ID:     "sleep-upsert-natural",
			Title:  "Record a subjective sleep check-in with optional wakeups",
			Prompt: "Use the configured local OpenHealth data path. For my 03/29/2026 wake date, I slept good, woke up 2 times, and the note is woke up after storm. Record that sleep check-in and tell me what is stored.",
		},
		{
			ID:     "sleep-latest-only",
			Title:  "List only the latest sleep check-in",
			Prompt: "What is my latest local OpenHealth sleep check-in? Use the configured local data path and mention only the latest row.",
		},
		{
			ID:     "sleep-invalid-input",
			Title:  "Reject invalid sleep quality and wakeup count without writing",
			Prompt: "Please add this local OpenHealth sleep check-in for 03/31/2026: quality 6 out of 5 and woke up -1 times.",
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
			ID:     "medication-add-list",
			Title:  "Record medications and list active courses",
			Prompt: "Use the configured local OpenHealth data path. Record these medications: Levothyroxine 25 mcg starting 01/01/2026 and Vitamin D starting 02/01/2026 ending 03/01/2026. Then list my active medications only.",
		},
		{
			ID:     "medication-non-oral-dosage",
			Title:  "Record a non-oral medication dosage text",
			Prompt: "Use the configured local OpenHealth data path. Record Semaglutide 2.5 mg subcutaneous injection weekly starting 02/01/2026. Then list my active medications.",
		},
		{
			ID:     "medication-note",
			Title:  "Record medication course narrative note",
			Prompt: "Use the configured local OpenHealth data path. Record Semaglutide 0.25 mg subcutaneous injection weekly starting 02/01/2026 with this medication note: coverage approved after prior authorization. Then list my active medications.",
		},
		{
			ID:     "medication-correct",
			Title:  "Correct an existing medication course",
			Prompt: "Use the configured local OpenHealth data path. Correct my Levothyroxine medication that started 01/01/2026 so the dosage is 50 mcg. Tell me what is stored now.",
		},
		{
			ID:     "medication-delete",
			Title:  "Delete an existing medication course",
			Prompt: "Use the configured local OpenHealth data path. Delete the Vitamin D medication course that started 02/01/2026. Then list all medications.",
		},
		{
			ID:     "medication-invalid-date",
			Title:  "Reject an invalid medication date without writing",
			Prompt: "Please add this local OpenHealth medication exactly as written: Levothyroxine 25 mcg starting 2026/01/01. Do not normalize or rewrite the date if OpenHealth requires another date format.",
		},
		{
			ID:     "medication-end-before-start",
			Title:  "Reject a medication end date before start date",
			Prompt: "Please add this local OpenHealth medication: Levothyroxine 25 mcg starting 01/02/2026 and ending 01/01/2026.",
		},
		{
			ID:     "lab-record-list",
			Title:  "Record labs and list latest collection",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with a Metabolic panel containing Glucose 89 mg/dL, canonical analyte glucose, range 70-99. Then show my latest lab collection.",
		},
		{
			ID:     "lab-arbitrary-slug",
			Title:  "Record and list an arbitrary lab analyte slug",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with a Micronutrients panel containing Vitamin D 32 ng/mL, canonical analyte vitamin-d. Then show my latest Vitamin D lab result.",
		},
		{
			ID:     "lab-note",
			Title:  "Record lab collection with clinician note",
			Prompt: "Use the configured local OpenHealth data path. Record a lab collection for 03/29/2026 with collection note \"labs look stable, keep moving\" and a Metabolic panel containing Glucose 89 mg/dL, canonical analyte glucose, with result notes \"HIV 4th gen narrative\" and \"A1C context\". Then show my latest lab collection.",
		},
		{
			ID:     "lab-same-day-multiple",
			Title:  "Record multiple distinct same-day lab collections",
			Prompt: "Use the configured local OpenHealth data path. Record two lab collections for 03/29/2026: one Metabolic panel with Glucose 89 mg/dL canonical analyte glucose, and one Thyroid panel with TSH 3.1 uIU/mL canonical analyte tsh. Then list lab collections newest first.",
		},
		{
			ID:     "lab-range",
			Title:  "List a bounded lab date range",
			Prompt: "Use the configured local OpenHealth data path. Show my OpenHealth lab collections for Mar 29 and Mar 30, 2026 only, newest first. Do not mention entries outside that requested range.",
		},
		{
			ID:     "lab-latest-analyte",
			Title:  "List latest lab result for a canonical analyte",
			Prompt: "Use the configured local OpenHealth data path. What is my latest glucose lab result? Mention only the latest matching collection/result.",
		},
		{
			ID:     "lab-correct",
			Title:  "Correct an existing lab collection",
			Prompt: "Use the configured local OpenHealth data path. Correct my lab collection dated 03/29/2026 so it has a Thyroid panel with TSH 3.1 uIU/mL, canonical analyte tsh. Tell me what is stored now.",
		},
		{
			ID:     "lab-patch",
			Title:  "Patch one lab result while preserving sibling results",
			Prompt: "Use the configured local OpenHealth data path. In my existing 03/29/2026 lab collection, correct only the Glucose result to 92 mg/dL and preserve the other lab results. Tell me what is stored now.",
		},
		{
			ID:     "lab-delete",
			Title:  "Delete an existing lab collection",
			Prompt: "Use the configured local OpenHealth data path. Delete my lab collection dated 03/29/2026. Then list lab collections.",
		},
		{
			ID:     "lab-invalid-slug",
			Title:  "Reject an invalid lab analyte slug shape without writing",
			Prompt: "Please add this local OpenHealth lab for 03/29/2026: UnknownTest 1 mg/dL with canonical analyte bad/slug. Do not write it if the analyte slug shape is invalid.",
		},
		{
			ID:     "mixed-medication-lab",
			Title:  "Record medication and lab data, then report latest domain entries",
			Prompt: "Use the configured local OpenHealth data path. Record Levothyroxine 25 mcg starting 01/01/2026 and a 03/29/2026 glucose lab result of 89 mg/dL. Then tell me the active medication and latest lab result.",
		},
		{
			ID:     "imaging-record-list",
			Title:  "Record an imaging summary and list it",
			Prompt: "Use the configured local OpenHealth data path. Record imaging from 03/29/2026: modality X-ray, body site chest, title Chest X-ray, summary No acute cardiopulmonary abnormality, impression Normal chest radiograph, note ordered for cough, result notes \"XR TOE RIGHT narrative\" and \"US Head/Neck findings\". Then list my latest imaging records.",
		},
		{
			ID:     "imaging-correct",
			Title:  "Correct an existing imaging summary",
			Prompt: "Use the configured local OpenHealth data path. Correct my 03/29/2026 imaging record so the modality is CT, body site chest, and the summary is Stable small pulmonary nodule. Tell me what is stored now.",
		},
		{
			ID:     "imaging-delete",
			Title:  "Delete an existing imaging record",
			Prompt: "Use the configured local OpenHealth data path. Delete my imaging record dated 03/29/2026. Then list imaging records.",
		},
		{
			ID:     "mixed-import-file-coverage",
			Title:  "Import file-style data that previously risked skipped rows",
			Prompt: "Use the configured local OpenHealth data path. Import this health-file data and do not skip supported fields: weight 154.2 lb with note morning scale and body-fat 18.7% on 03/29/2026; lab collection 03/29/2026 Glucose 89 mg/dL canonical analyte glucose with collection note \"labs look stable, keep moving\" and result notes \"HIV 4th gen narrative\" and \"A1C context\"; chest X-ray 03/29/2026 summary No acute cardiopulmonary abnormality impression Normal chest radiograph with result notes \"XR TOE RIGHT narrative\" and \"US Head/Neck findings\"; medication Semaglutide 0.25 mg subcutaneous injection weekly starting 02/01/2026 with note coverage approved after prior authorization. Then summarize what is stored.",
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
	case "sleep-latest-only":
		return upsertSleep(ctx, api, []sleepState{
			{Date: "2026-03-28", QualityScore: 2, WakeupCount: intPointer(3)},
			{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(1)},
			{Date: "2026-03-30", QualityScore: 5},
		})
	case "medication-correct", "medication-delete":
		return recordMedications(ctx, api, []medicationState{
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
		})
	case "lab-range", "lab-latest-analyte", "lab-correct", "lab-delete":
		return recordLabs(ctx, api, []labCollectionState{
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
		})
	case "lab-patch":
		return recordLabs(ctx, api, []labCollectionState{
			{Date: "2026-03-29", Results: []labResultState{
				{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")},
				{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
			}},
		})
	case "imaging-correct", "imaging-delete":
		return recordImaging(ctx, api, []imagingState{
			{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough")},
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
			Note:       weight.Note,
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
			Note:       reading.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func upsertSleep(ctx context.Context, api *client.LocalClient, entries []sleepState) error {
	for _, entry := range entries {
		recordedAt, err := parseDate(entry.Date)
		if err != nil {
			return err
		}
		if _, err := api.UpsertSleep(ctx, client.SleepInput{
			RecordedAt:   recordedAt,
			QualityScore: entry.QualityScore,
			WakeupCount:  entry.WakeupCount,
			Note:         entry.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordMedications(ctx context.Context, api *client.LocalClient, medications []medicationState) error {
	for _, medication := range medications {
		if _, err := api.CreateMedicationCourse(ctx, client.MedicationCourseInput{
			Name:       medication.Name,
			DosageText: medication.DosageText,
			StartDate:  medication.StartDate,
			EndDate:    medication.EndDate,
			Note:       medication.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordLabs(ctx context.Context, api *client.LocalClient, collections []labCollectionState) error {
	for _, collection := range collections {
		collectedAt, err := parseDate(collection.Date)
		if err != nil {
			return err
		}
		results := make([]client.LabResultInput, 0, len(collection.Results))
		for _, result := range collection.Results {
			results = append(results, client.LabResultInput{
				TestName:      result.TestName,
				CanonicalSlug: clientAnalyteSlug(result.CanonicalSlug),
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				Notes:         append([]string(nil), result.Notes...),
			})
		}
		if _, err := api.CreateLabCollection(ctx, client.LabCollectionInput{
			CollectedAt: collectedAt,
			Note:        collection.Note,
			Panels:      []client.LabPanelInput{{PanelName: "Panel", Results: results}},
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordBodyComposition(ctx context.Context, api *client.LocalClient, records []bodyCompositionState) error {
	for _, record := range records {
		recordedAt, err := parseDate(record.Date)
		if err != nil {
			return err
		}
		var unit *client.WeightUnit
		if record.WeightUnit != nil {
			value := client.WeightUnit(*record.WeightUnit)
			unit = &value
		}
		if _, err := api.CreateBodyComposition(ctx, client.BodyCompositionInput{
			RecordedAt:     recordedAt,
			BodyFatPercent: record.BodyFatPercent,
			WeightValue:    record.WeightValue,
			WeightUnit:     unit,
			Method:         record.Method,
			Note:           record.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

func recordImaging(ctx context.Context, api *client.LocalClient, records []imagingState) error {
	for _, record := range records {
		performedAt, err := parseDate(record.Date)
		if err != nil {
			return err
		}
		if _, err := api.CreateImaging(ctx, client.ImagingRecordInput{
			PerformedAt: performedAt,
			Modality:    record.Modality,
			BodySite:    record.BodySite,
			Title:       record.Title,
			Summary:     record.Summary,
			Impression:  record.Impression,
			Note:        record.Note,
			Notes:       append([]string(nil), record.Notes...),
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
	if isBodyCompositionScenario(sc.ID) {
		return verifyBodyCompositionScenario(dbPath, sc, finalMessage)
	}
	if isBloodPressureScenario(sc.ID) {
		return verifyBloodPressureScenario(dbPath, sc, finalMessage)
	}
	if isSleepScenario(sc.ID) {
		return verifySleepScenario(dbPath, sc, finalMessage)
	}
	if isMedicationScenario(sc.ID) {
		return verifyMedicationScenario(dbPath, sc, finalMessage)
	}
	if isLabScenario(sc.ID) {
		return verifyLabScenario(dbPath, sc, finalMessage)
	}
	if isImagingScenario(sc.ID) {
		return verifyImagingScenario(dbPath, sc, finalMessage)
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
	medications, err := listMedications(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list medications: %w", err)
	}
	labs, err := listLabs(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list labs: %w", err)
	}
	bodyComposition, err := listBodyComposition(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list body composition: %w", err)
	}
	imaging, err := listImaging(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list imaging: %w", err)
	}
	weightStates := weightStates(weights)
	bloodPressureStates := bloodPressureStates(bloodPressures)
	medicationStates := medicationStates(medications)
	labStates := labCollectionStates(labs)
	bodyCompositionStates := bodyCompositionStates(bodyComposition)
	imagingStates := imagingStates(imaging)
	result := verificationResult{
		Weights:         weightStates,
		BodyComposition: bodyCompositionStates,
		BloodPressures:  bloodPressureStates,
		Medications:     medicationStates,
		Labs:            labStates,
		Imaging:         imagingStates,
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
	case "mixed-medication-lab":
		expectedMedications := []medicationState{{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"}}
		expectedLabs := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}}}
		result.DatabasePass = medicationsEqual(medicationStates, expectedMedications) && labsEqual(labStates, expectedLabs) && len(weightStates) == 0 && len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "25 mcg", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected one medication and one glucose lab; observed medications %s and labs %s", describeMedications(medicationStates), describeLabs(labStates))
	case "mixed-import-file-coverage":
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb", Note: stringPointer("morning scale")}}
		expectedBody := []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb")}}
		expectedMedications := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}
		expectedLabs := []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}}
		expectedImaging := []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality", Impression: stringPointer("Normal chest radiograph"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) &&
			bodyCompositionEqual(bodyCompositionStates, expectedBody) &&
			medicationsEqual(medicationStates, expectedMedications) &&
			labsEqual(labStates, expectedLabs) &&
			imagingEqual(imagingStates, expectedImaging) &&
			len(bloodPressureStates) == 0
		result.AssistantPass = containsAll(finalMessage, []string{"154.2", "18.7", "Glucose", "89", "Semaglutide", "X-ray", "narrative"})
		result.Details = fmt.Sprintf("expected no skipped import-file rows; observed weights %s, body composition %s, medications %s, labs %s, imaging %s", describeWeights(weightStates), describeBodyComposition(bodyCompositionStates), describeMedications(medicationStates), describeLabs(labStates), describeImaging(imagingStates))
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
		result.AssistantPass = includedBloodPressureResultLineIndex(finalMessage, bloodPressureState{Date: "2026-03-30", Systolic: 117, Diastolic: 75}) >= 0
		result.Details = fmt.Sprintf("expected latest blood-pressure correction on 2026-03-30; observed weights %s and blood pressures %s", describeWeights(weightStates), describeBloodPressures(bloodPressureStates))
	default:
		return verificationResult{}, fmt.Errorf("unknown mixed or multi-turn scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyBodyCompositionScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	weights, err := listWeights(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list weights: %w", err)
	}
	bodyComposition, err := listBodyComposition(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list body composition: %w", err)
	}
	weightStates := weightStates(weights)
	bodyStates := bodyCompositionStates(bodyComposition)
	result := verificationResult{
		Weights:         weightStates,
		BodyComposition: bodyStates,
	}

	switch sc.ID {
	case "body-composition-combined-weight-row":
		expectedWeights := []weightState{{Date: "2026-03-29", Value: 154.2, Unit: "lb"}}
		expectedBody := []bodyCompositionState{{Date: "2026-03-29", BodyFatPercent: floatPointer(18.7), WeightValue: floatPointer(154.2), WeightUnit: stringPointer("lb"), Method: stringPointer("smart scale")}}
		result.DatabasePass = weightsEqual(weightStates, expectedWeights) && bodyCompositionEqual(bodyStates, expectedBody)
		result.AssistantPass = containsAll(finalMessage, []string{"154.2", "18.7"})
		result.Details = fmt.Sprintf("expected combined import row split into weight and body-composition domains; observed weights %s and body composition %s", describeWeights(weightStates), describeBodyComposition(bodyStates))
	default:
		return verificationResult{}, fmt.Errorf("unknown body-composition scenario %q", sc.ID)
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
			{Date: "2026-03-29", Systolic: 122, Diastolic: 78, Pulse: intPointer(64), Note: stringPointer("home cuff")},
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
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"no existing", "no local", "missing", "not found", "cannot", "can't", "did not", "not updated", "no update"})
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
	case "bp-invalid-relation":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "systolic", "diastolic", "greater", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and a systolic-greater-than-diastolic rejection; observed %s", describeBloodPressures(states))
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

func verifySleepScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	entries, err := listSleep(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list sleep: %w", err)
	}
	states := sleepStates(entries)
	result := verificationResult{Sleep: states}

	switch sc.ID {
	case "sleep-upsert-natural":
		expected := []sleepState{{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(2), Note: stringPointer("woke up after storm")}}
		result.DatabasePass = sleepEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "4"}) && sleepWakeupCountAssistantPass(finalMessage, 2)
		result.Details = fmt.Sprintf("expected one sleep check-in with quality 4 and two wakeups; observed %s", describeSleep(states))
	case "sleep-latest-only":
		expectedDB := []sleepState{
			{Date: "2026-03-30", QualityScore: 5},
			{Date: "2026-03-29", QualityScore: 4, WakeupCount: intPointer(1)},
			{Date: "2026-03-28", QualityScore: 2, WakeupCount: intPointer(3)},
		}
		result.DatabasePass = sleepEqual(states, expectedDB)
		result.AssistantPass = mentionsIncludedDate(finalMessage, "2026-03-30") &&
			!mentionsIncludedDate(finalMessage, "2026-03-29") &&
			!mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to latest sleep row 2026-03-30; observed %s", describeSleep(states))
	case "sleep-invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "quality", "1-5", "between 1 and 5", "wakeup", "negative", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and an invalid sleep rejection; observed %s", describeSleep(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown sleep scenario %q", sc.ID)
	}

	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyMedicationScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	medications, err := listMedications(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list medications: %w", err)
	}
	states := medicationStates(medications)
	result := verificationResult{Medications: states}

	switch sc.ID {
	case "medication-add-list":
		expected := []medicationState{
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
			{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"},
		}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "25 mcg"}) && !strings.Contains(strings.ToLower(finalMessage), "vitamin d")
		result.Details = fmt.Sprintf("expected two stored medications and active output limited to Levothyroxine; observed %s", describeMedications(states))
	case "medication-non-oral-dosage":
		expected := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("2.5 mg subcutaneous injection weekly"), StartDate: "2026-02-01"}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Semaglutide", "subcutaneous", "weekly"})
		result.Details = fmt.Sprintf("expected Semaglutide non-oral dosage text; observed %s", describeMedications(states))
	case "medication-note":
		expected := []medicationState{{Name: "Semaglutide", DosageText: stringPointer("0.25 mg subcutaneous injection weekly"), StartDate: "2026-02-01", Note: stringPointer("coverage approved after prior authorization")}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Semaglutide", "subcutaneous", "coverage approved"})
		result.Details = fmt.Sprintf("expected Semaglutide medication note; observed %s", describeMedications(states))
	case "medication-correct":
		expected := []medicationState{
			{Name: "Vitamin D", StartDate: "2026-02-01", EndDate: stringPointer("2026-03-01")},
			{Name: "Levothyroxine", DosageText: stringPointer("50 mcg"), StartDate: "2026-01-01"},
		}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"Levothyroxine", "50 mcg"})
		result.Details = fmt.Sprintf("expected Levothyroxine dosage correction; observed %s", describeMedications(states))
	case "medication-delete":
		expected := []medicationState{{Name: "Levothyroxine", DosageText: stringPointer("25 mcg"), StartDate: "2026-01-01"}}
		result.DatabasePass = medicationsEqual(states, expected)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no vitamin d"}) && containsAll(finalMessage, []string{"Levothyroxine"})
		result.Details = fmt.Sprintf("expected Vitamin D deleted and Levothyroxine retained; observed %s", describeMedications(states))
	case "medication-invalid-date":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = nonISODateRejectAssistantPass(finalMessage)
		result.Details = fmt.Sprintf("expected no write and strict medication date rejection; observed %s", describeMedications(states))
	case "medication-end-before-start":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"end", "start", "before", "invalid", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and end-before-start rejection; observed %s", describeMedications(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown medication scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyLabScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	labs, err := listLabs(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list labs: %w", err)
	}
	states := labCollectionStates(labs)
	result := verificationResult{Labs: states}

	switch sc.ID {
	case "lab-record-list":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected one glucose lab collection; observed %s", describeLabs(states))
	case "lab-arbitrary-slug":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{{TestName: "Vitamin D", CanonicalSlug: stringPointer("vitamin-d"), ValueText: "32", ValueNumeric: floatPointer(32), Units: stringPointer("ng/mL")}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Vitamin D", "32"})
		result.Details = fmt.Sprintf("expected one Vitamin D lab collection with arbitrary slug; observed %s", describeLabs(states))
	case "lab-note":
		expected := []labCollectionState{{Date: "2026-03-29", Note: stringPointer("labs look stable, keep moving"), Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL"), Notes: []string{"HIV 4th gen narrative", "A1C context"}}}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89", "stable", "A1C"})
		result.Details = fmt.Sprintf("expected lab collection note; observed %s", describeLabs(states))
	case "lab-same-day-multiple":
		expected := []labCollectionState{
			{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "TSH", "3.1", "Glucose", "89"})
		result.Details = fmt.Sprintf("expected two distinct same-day lab collections; observed %s", describeLabs(states))
	case "lab-range":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = mentionsDatesInOrder(finalMessage, "2026-03-30", "2026-03-29") && !mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged lab seed rows and output limited to 2026-03-29..2026-03-30; observed %s", describeLabs(states))
	case "lab-latest-analyte":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "89", ValueNumeric: floatPointer(89), Units: stringPointer("mg/dL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "89"}) && !mentionsIncludedDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged lab seed rows and latest glucose answer; observed %s", describeLabs(states))
	case "lab-correct":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-29", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.1", ValueNumeric: floatPointer(3.1), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "TSH", "3.1"})
		result.Details = fmt.Sprintf("expected 2026-03-29 lab correction; observed %s", describeLabs(states))
	case "lab-patch":
		expected := []labCollectionState{{Date: "2026-03-29", Results: []labResultState{
			{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "92", ValueNumeric: floatPointer(92), Units: stringPointer("mg/dL")},
			{TestName: "HDL", CanonicalSlug: stringPointer("hdl"), ValueText: "51", ValueNumeric: floatPointer(51), Units: stringPointer("mg/dL")},
		}}}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "Glucose", "92", "HDL", "51"})
		result.Details = fmt.Sprintf("expected one patched glucose result with sibling HDL preserved; observed %s", describeLabs(states))
	case "lab-delete":
		expected := []labCollectionState{
			{Date: "2026-03-30", Results: []labResultState{{TestName: "TSH", CanonicalSlug: stringPointer("tsh"), ValueText: "3.4", ValueNumeric: floatPointer(3.4), Units: stringPointer("uIU/mL")}}},
			{Date: "2026-03-28", Results: []labResultState{{TestName: "Glucose", CanonicalSlug: stringPointer("glucose"), ValueText: "91", ValueNumeric: floatPointer(91), Units: stringPointer("mg/dL")}}},
		}
		result.DatabasePass = labsEqual(states, expected)
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no lab"}) && !mentionsIncludedDate(finalMessage, "2026-03-29")
		result.Details = fmt.Sprintf("expected 2026-03-29 lab deleted; observed %s", describeLabs(states))
	case "lab-invalid-slug":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"analyte", "slug", "invalid", "cannot", "can't", "reject"})
		result.Details = fmt.Sprintf("expected no write and invalid analyte slug rejection; observed %s", describeLabs(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown lab scenario %q", sc.ID)
	}
	result.Passed = result.DatabasePass && result.AssistantPass
	return result, nil
}

func verifyImagingScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
	imaging, err := listImaging(dbPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("list imaging: %w", err)
	}
	states := imagingStates(imaging)
	result := verificationResult{Imaging: states}

	switch sc.ID {
	case "imaging-record-list":
		expected := []imagingState{{Date: "2026-03-29", Modality: "X-ray", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "No acute cardiopulmonary abnormality", Impression: stringPointer("Normal chest radiograph"), Note: stringPointer("ordered for cough"), Notes: []string{"XR TOE RIGHT narrative", "US Head/Neck findings"}}}
		result.DatabasePass = imagingEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "X-ray", "chest", "narrative"})
		result.Details = fmt.Sprintf("expected one chest X-ray imaging record; observed %s", describeImaging(states))
	case "imaging-correct":
		expected := []imagingState{{Date: "2026-03-29", Modality: "CT", BodySite: stringPointer("chest"), Title: stringPointer("Chest X-ray"), Summary: "Stable small pulmonary nodule.", Impression: stringPointer("Normal chest radiograph."), Note: stringPointer("ordered for cough")}}
		result.DatabasePass = imagingEqual(states, expected)
		result.AssistantPass = containsAll(finalMessage, []string{"2026-03-29", "CT", "Stable"})
		result.Details = fmt.Sprintf("expected corrected CT imaging record; observed %s", describeImaging(states))
	case "imaging-delete":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"deleted", "removed", "no imaging"})
		result.Details = fmt.Sprintf("expected imaging record deleted; observed %s", describeImaging(states))
	default:
		return verificationResult{}, fmt.Errorf("unknown imaging scenario %q", sc.ID)
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

func listMedications(dbPath string) ([]client.MedicationCourse, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListMedicationCourses(context.Background(), client.MedicationListOptions{Status: client.MedicationStatusAll})
}

func listLabs(dbPath string) ([]client.LabCollection, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListLabCollections(context.Background())
}

func listBodyComposition(dbPath string) ([]client.BodyCompositionEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListBodyComposition(context.Background(), client.BodyCompositionListOptions{Limit: 100})
}

func listSleep(dbPath string) ([]client.SleepEntry, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListSleep(context.Background(), client.SleepListOptions{Limit: 100})
}

func listImaging(dbPath string) ([]client.ImagingRecord, error) {
	api, err := client.OpenLocal(client.LocalConfig{DatabasePath: dbPath})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = api.Close()
	}()
	return api.ListImaging(context.Background(), client.ImagingListOptions{Limit: 100})
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
SELECT recorded_at, value, unit, note
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
		var note sql.NullString
		if err := rows.Scan(&state.Date, &state.Value, &state.Unit, &note); err != nil {
			return nil, err
		}
		state.Value = roundWeight(state.Value)
		if note.Valid {
			state.Note = &note.String
		}
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
			Note:  weight.Note,
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
			Note:      reading.Note,
		})
	}
	return states
}

func medicationStates(medications []client.MedicationCourse) []medicationState {
	states := make([]medicationState, 0, len(medications))
	for _, medication := range medications {
		states = append(states, medicationState{
			Name:       medication.Name,
			DosageText: medication.DosageText,
			StartDate:  medication.StartDate,
			EndDate:    medication.EndDate,
			Note:       medication.Note,
		})
	}
	return states
}

func labCollectionStates(collections []client.LabCollection) []labCollectionState {
	states := make([]labCollectionState, 0, len(collections))
	for _, collection := range collections {
		state := labCollectionState{
			Date: collection.CollectedAt.Format(time.DateOnly),
			Note: collection.Note,
		}
		for _, panel := range collection.Panels {
			for _, result := range panel.Results {
				var slug *string
				if result.CanonicalSlug != nil {
					value := string(*result.CanonicalSlug)
					slug = &value
				}
				state.Results = append(state.Results, labResultState{
					TestName:      result.TestName,
					CanonicalSlug: slug,
					ValueText:     result.ValueText,
					ValueNumeric:  result.ValueNumeric,
					Units:         result.Units,
					Notes:         append([]string(nil), result.Notes...),
				})
			}
		}
		states = append(states, state)
	}
	return states
}

func bodyCompositionStates(records []client.BodyCompositionEntry) []bodyCompositionState {
	states := make([]bodyCompositionState, 0, len(records))
	for _, record := range records {
		var weightUnit *string
		if record.WeightUnit != nil {
			value := string(*record.WeightUnit)
			weightUnit = &value
		}
		states = append(states, bodyCompositionState{
			Date:           record.RecordedAt.Format(time.DateOnly),
			BodyFatPercent: record.BodyFatPercent,
			WeightValue:    record.WeightValue,
			WeightUnit:     weightUnit,
			Method:         record.Method,
			Note:           record.Note,
		})
	}
	return states
}

func sleepStates(entries []client.SleepEntry) []sleepState {
	states := make([]sleepState, 0, len(entries))
	for _, entry := range entries {
		states = append(states, sleepState{
			Date:         entry.RecordedAt.Format(time.DateOnly),
			QualityScore: entry.QualityScore,
			WakeupCount:  entry.WakeupCount,
			Note:         entry.Note,
		})
	}
	return states
}

func imagingStates(records []client.ImagingRecord) []imagingState {
	states := make([]imagingState, 0, len(records))
	for _, record := range records {
		states = append(states, imagingState{
			Date:       record.PerformedAt.Format(time.DateOnly),
			Modality:   record.Modality,
			BodySite:   record.BodySite,
			Title:      record.Title,
			Summary:    record.Summary,
			Impression: record.Impression,
			Note:       record.Note,
			Notes:      append([]string(nil), record.Notes...),
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
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	if err := copyDir(src, dest); err != nil {
		return err
	}
	return nil
}

func preflightEvalContext(repoRoot string, runRepo string, runDir string, cache cacheConfig) error {
	sourceSkill := filepath.Join(repoRoot, "skills", "openhealth", "SKILL.md")
	installedSkill := filepath.Join(runRepo, ".agents", "skills", "openhealth", "SKILL.md")
	sourceBytes, err := os.ReadFile(sourceSkill)
	if err != nil {
		return err
	}
	installedBytes, err := os.ReadFile(installedSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, installedBytes) {
		return errors.New("installed production skill does not match shipped SKILL.md")
	}
	if _, err := os.Stat(filepath.Join(runRepo, "AGENTS.md")); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("production eval repo must not contain AGENTS.md")
		}
		return err
	}

	cmd := exec.Command("codex", "debug", "prompt-input", "Use OpenHealth to list latest weight.")
	cmd.Dir = runRepo
	cmd.Env = evalEnv(runDir, evalDatabasePath(runRepo), cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	rendered := string(output)
	if !strings.Contains(rendered, "- openhealth:") {
		return errors.New("rendered prompt is missing openhealth skill discovery")
	}
	if !strings.Contains(rendered, ".agents/skills/openhealth/SKILL.md") {
		return errors.New("rendered prompt does not point openhealth to the installed project skill")
	}
	if containsOpenHealthAgentsInstructions(rendered) {
		return errors.New("rendered prompt contains OpenHealth product instructions from AGENTS.md")
	}
	return nil
}

func containsOpenHealthAgentsInstructions(rendered string) bool {
	const marker = "# AGENTS.md instructions"
	index := strings.Index(rendered, marker)
	if index < 0 {
		return false
	}
	agentsText := rendered[index:]
	for _, forbidden := range []string{
		"openhealth",
		"upsert_weights",
		"record_blood_pressure",
		"record_medications",
		"record_labs",
		"ambiguous short date",
		"product data agent",
	} {
		if strings.Contains(agentsText, forbidden) {
			return true
		}
	}
	return false
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
		notes = append(notes, "oh-23a intentionally keeps agent-facing readiness scoped to weight and blood pressure; labs and medications remain a separate runner expansion tracked in oh-bng.")
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
	for _, result := range results {
		if result.Variant == "production" {
			productionByScenario[result.Scenario] = result
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

	recommendation := "ship_openhealth_runner_production"
	if len(triggers) > 0 {
		recommendation = "continue_production_hardening"
	}
	return &stopLossSummary{
		Policy:         "Production OpenHealth runner must pass every scenario without direct generated-file inspection, module-cache inspection, direct SQLite access, retired human OpenHealth CLI usage, broad repo search in more than one routine scenario, or excessive core tool counts.",
		Triggered:      len(triggers) > 0,
		Recommendation: recommendation,
		Triggers:       triggers,
	}
}

func isRoutineScenario(id string) bool {
	switch id {
	case "add-two", "repeat-add", "update-existing", "bounded-range", "bounded-range-natural", "latest-only", "history-limit-two",
		"body-composition-combined-weight-row",
		"bp-add-two", "bp-latest-only", "bp-history-limit-two", "bp-bounded-range", "bp-bounded-range-natural",
		"bp-correct-existing", "bp-correct-missing-reject", "bp-correct-ambiguous-reject",
		"sleep-upsert-natural", "sleep-latest-only",
		"mixed-add-latest", "mixed-bounded-range", "mt-bp-latest-then-correct", "mt-mixed-latest-then-correct",
		"medication-add-list", "medication-note", "medication-correct", "medication-delete",
		"lab-record-list", "lab-note", "lab-range", "lab-latest-analyte", "lab-correct", "lab-delete",
		"mixed-medication-lab", "mixed-import-file-coverage",
		"imaging-record-list", "imaging-correct", "imaging-delete":
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

func isSleepScenario(id string) bool {
	return strings.HasPrefix(id, "sleep-")
}

func isBodyCompositionScenario(id string) bool {
	return strings.HasPrefix(id, "body-composition-")
}

func isMedicationScenario(id string) bool {
	return strings.HasPrefix(id, "medication-")
}

func isLabScenario(id string) bool {
	return strings.HasPrefix(id, "lab-")
}

func isImagingScenario(id string) bool {
	return strings.HasPrefix(id, "imaging-")
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
	fmt.Fprintf(&b, "- New Codex session files under `<run-root>/codex-home`: `%d`.\n", value.HistoryIsolation.NewSessionFilesAfterRun)
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
	fmt.Fprintf(&b, "| build_agent_app | %.2f |\n", value.PhaseTotals.BuildAgentApp)
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

func commandOutputWithEnv(env []string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(out))
}

func countNewSessionFiles(marker time.Time, runRoot string) int {
	sessionsDir := filepath.Join(evalCodexHome(runRoot), "sessions")
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

func stringPointer(value string) *string {
	return &value
}

func floatPointer(value float64) *float64 {
	return &value
}

func clientAnalyteSlug(value *string) *client.AnalyteSlug {
	if value == nil {
		return nil
	}
	slug := client.AnalyteSlug(*value)
	return &slug
}

func equalIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalExpectedStringPointer(got *string, want *string) bool {
	if want == nil {
		return true
	}
	return equalClinicalSentencePointer(got, want)
}

func equalExpectedStringSlice(got []string, want []string) bool {
	if len(want) == 0 {
		return true
	}
	return slices.Equal(got, want)
}

func equalClinicalSentence(left string, right string) bool {
	return normalizeClinicalSentence(left) == normalizeClinicalSentence(right)
}

func equalClinicalSentencePointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return equalClinicalSentence(*left, *right)
}

func normalizeClinicalSentence(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), ".")
}

func equalFloatPointer(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return math.Abs(*left-*right) < 0.001
}

func weightsEqual(got []weightState, want []weightState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Unit != want[i].Unit ||
			math.Abs(got[i].Value-want[i].Value) > 0.001 ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
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
			!equalIntPointer(got[i].Pulse, want[i].Pulse) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func medicationsEqual(got []medicationState, want []medicationState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Name != want[i].Name ||
			got[i].StartDate != want[i].StartDate ||
			!equalStringPointer(got[i].DosageText, want[i].DosageText) ||
			!equalStringPointer(got[i].EndDate, want[i].EndDate) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func labsEqual(got []labCollectionState, want []labCollectionState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) ||
			len(got[i].Results) != len(want[i].Results) {
			return false
		}
		for j := range got[i].Results {
			if got[i].Results[j].TestName != want[i].Results[j].TestName ||
				!equalStringPointer(got[i].Results[j].CanonicalSlug, want[i].Results[j].CanonicalSlug) ||
				!labValueTextEqual(got[i].Results[j].ValueText, want[i].Results[j].ValueText, want[i].Results[j].Units) ||
				!equalFloatPointer(got[i].Results[j].ValueNumeric, want[i].Results[j].ValueNumeric) ||
				!equalStringPointer(got[i].Results[j].Units, want[i].Results[j].Units) ||
				!equalExpectedStringSlice(got[i].Results[j].Notes, want[i].Results[j].Notes) {
				return false
			}
		}
	}
	return true
}

func labValueTextEqual(got string, want string, units *string) bool {
	if got == want {
		return true
	}
	if units == nil {
		return false
	}
	return got == strings.TrimSpace(want+" "+*units)
}

func bodyCompositionEqual(got []bodyCompositionState, want []bodyCompositionState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			!equalFloatPointer(got[i].BodyFatPercent, want[i].BodyFatPercent) ||
			!equalFloatPointer(got[i].WeightValue, want[i].WeightValue) ||
			!equalStringPointer(got[i].WeightUnit, want[i].WeightUnit) ||
			!equalStringPointer(got[i].Method, want[i].Method) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func sleepEqual(got []sleepState, want []sleepState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].QualityScore != want[i].QualityScore ||
			!equalIntPointer(got[i].WakeupCount, want[i].WakeupCount) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) {
			return false
		}
	}
	return true
}

func imagingEqual(got []imagingState, want []imagingState) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].Date != want[i].Date ||
			got[i].Modality != want[i].Modality ||
			!equalClinicalSentence(got[i].Summary, want[i].Summary) ||
			!equalStringPointer(got[i].BodySite, want[i].BodySite) ||
			!equalStringPointer(got[i].Title, want[i].Title) ||
			!equalClinicalSentencePointer(got[i].Impression, want[i].Impression) ||
			!equalExpectedStringPointer(got[i].Note, want[i].Note) ||
			!equalExpectedStringSlice(got[i].Notes, want[i].Notes) {
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
		note := ""
		if weight.Note != nil {
			note = " note " + *weight.Note
		}
		parts = append(parts, fmt.Sprintf("%s %.1f %s%s", weight.Date, weight.Value, weight.Unit, note))
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
		note := ""
		if reading.Note != nil {
			note = " note " + *reading.Note
		}
		parts = append(parts, fmt.Sprintf("%s %d/%d%s%s", reading.Date, reading.Systolic, reading.Diastolic, pulse, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeMedications(medications []medicationState) string {
	if len(medications) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(medications))
	for _, medication := range medications {
		dosage := ""
		if medication.DosageText != nil {
			dosage = " " + *medication.DosageText
		}
		end := ""
		if medication.EndDate != nil {
			end = " to " + *medication.EndDate
		}
		note := ""
		if medication.Note != nil {
			note = " note " + *medication.Note
		}
		parts = append(parts, fmt.Sprintf("%s%s from %s%s%s", medication.Name, dosage, medication.StartDate, end, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeLabs(collections []labCollectionState) string {
	if len(collections) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(collections))
	for _, collection := range collections {
		results := make([]string, 0, len(collection.Results))
		for _, result := range collection.Results {
			note := ""
			if len(result.Notes) > 0 {
				note = " notes " + strings.Join(result.Notes, " / ")
			}
			results = append(results, fmt.Sprintf("%s %s%s", result.TestName, result.ValueText, note))
		}
		note := ""
		if collection.Note != nil {
			note = " note " + *collection.Note
		}
		parts = append(parts, fmt.Sprintf("%s%s (%s)", collection.Date, note, strings.Join(results, ", ")))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeBodyComposition(records []bodyCompositionState) string {
	if len(records) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(records))
	for _, record := range records {
		values := []string{}
		if record.BodyFatPercent != nil {
			values = append(values, fmt.Sprintf("body fat %.1f%%", *record.BodyFatPercent))
		}
		if record.WeightValue != nil && record.WeightUnit != nil {
			values = append(values, fmt.Sprintf("weight %.1f %s", *record.WeightValue, *record.WeightUnit))
		}
		if record.Method != nil {
			values = append(values, "method "+*record.Method)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", record.Date, strings.Join(values, ", ")))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeSleep(entries []sleepState) string {
	if len(entries) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		wakeups := ""
		if entry.WakeupCount != nil {
			wakeups = fmt.Sprintf(" wakeups %d", *entry.WakeupCount)
		}
		note := ""
		if entry.Note != nil {
			note = " note " + *entry.Note
		}
		parts = append(parts, fmt.Sprintf("%s quality %d%s%s", entry.Date, entry.QualityScore, wakeups, note))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeImaging(records []imagingState) string {
	if len(records) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(records))
	for _, record := range records {
		site := ""
		if record.BodySite != nil {
			site = " " + *record.BodySite
		}
		note := ""
		if len(record.Notes) > 0 {
			note = " notes " + strings.Join(record.Notes, " / ")
		}
		parts = append(parts, fmt.Sprintf("%s %s%s: %s%s", record.Date, record.Modality, site, record.Summary, note))
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
	return dateMentionIndex(message, "2026-03-31") >= 0 &&
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

func sleepWakeupCountAssistantPass(message string, count int) bool {
	lower := strings.ToLower(message)
	digits := fmt.Sprintf("%d", count)
	words := numberWord(count)
	needles := []string{
		"woke up " + digits,
		digits + " wake",
		"wakeups " + digits,
		"wakeups: " + digits,
		"wakeup count " + digits,
		"wakeup_count\":" + digits,
		"wakeup_count\": " + digits,
		digits + " times",
		digits + " time",
	}
	if words != "" {
		needles = append(needles,
			"woke up "+words,
			words+" wake",
			words+" times",
			words+" time",
		)
	}
	if count == 2 {
		needles = append(needles, "twice")
	}
	return containsAny(lower, needles)
}

func numberWord(value int) string {
	switch value {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	default:
		return ""
	}
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
		strings.Contains(lower, "sleep quality") ||
		strings.Contains(lower, "quality_score") ||
		strings.Contains(lower, "wakeups") ||
		strings.Contains(lower, "wakeup_count") ||
		strings.Contains(lower, "glucose") ||
		strings.Contains(lower, "tsh") ||
		strings.Contains(lower, "mg/dl") ||
		strings.Contains(lower, "uiu/ml") ||
		(strings.Contains(lower, `"date"`) &&
			strings.Contains(lower, `"value"`) &&
			(strings.Contains(lower, `"unit":"lb"`) || strings.Contains(lower, `"unit": "lb"`)))
}

func mentionsDatesInOrder(message string, dates ...string) bool {
	previous := -1
	for _, date := range dates {
		index := includedDateResultLineIndex(message, date)
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
	case "body-composition-combined-weight-row":
		return "record combined weight and body-fat import row through weight and body-composition"
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
	case "bp-invalid-relation":
		return "reject systolic not greater than diastolic without tools"
	case "bp-non-iso-date-reject":
		return "reject non-ISO blood-pressure date 2026/03/31"
	case "bp-correct-existing":
		return "correct 2026-03-29 blood pressure from 122/78 pulse 64 to 121/77 pulse 63"
	case "bp-correct-missing-reject":
		return "reject correction for missing 2026-03-31 blood-pressure reading without creating one"
	case "bp-correct-ambiguous-reject":
		return "reject ambiguous correction when multiple 2026-03-29 blood-pressure rows exist"
	case "sleep-upsert-natural":
		return "record subjective sleep quality and optional wakeup count"
	case "sleep-latest-only":
		return "list only the latest sleep check-in from preseeded rows"
	case "sleep-invalid-input":
		return "reject invalid sleep quality and wakeup count"
	case "mixed-add-latest":
		return "record one weight and one blood-pressure reading, then report latest for both"
	case "mixed-bounded-range":
		return "list only 2026-03-29 through 2026-03-30 for both domains"
	case "mixed-invalid-direct-reject":
		return "reject invalid mixed weight and blood-pressure values without tools"
	case "medication-add-list":
		return "record two medications, then list active medications"
	case "medication-non-oral-dosage":
		return "record Semaglutide subcutaneous injection dosage text"
	case "medication-note":
		return "record medication course narrative note"
	case "medication-correct":
		return "correct Levothyroxine dosage from 25 mcg to 50 mcg"
	case "medication-delete":
		return "delete Vitamin D and list all medications"
	case "medication-invalid-date":
		return "reject medication date 2026/01/01 without tools"
	case "medication-end-before-start":
		return "reject medication end date before start date without tools"
	case "lab-record-list":
		return "record one glucose lab collection and list latest labs"
	case "lab-arbitrary-slug":
		return "record one Vitamin D lab collection with arbitrary slug"
	case "lab-note":
		return "record one glucose lab collection with a clinician note"
	case "lab-same-day-multiple":
		return "record two distinct same-day lab collections"
	case "lab-range":
		return "list only 2026-03-29 through 2026-03-30 lab collections"
	case "lab-latest-analyte":
		return "list only the latest glucose lab result"
	case "lab-correct":
		return "correct 2026-03-29 lab collection to TSH"
	case "lab-patch":
		return "patch one glucose lab result while preserving HDL"
	case "lab-delete":
		return "delete 2026-03-29 lab collection"
	case "lab-invalid-slug":
		return "reject invalid lab analyte slug without tools"
	case "mixed-medication-lab":
		return "record one medication and one lab result, then report both"
	case "imaging-record-list":
		return "record one chest X-ray imaging summary and list it"
	case "imaging-correct":
		return "correct one seeded imaging summary"
	case "imaging-delete":
		return "delete one seeded imaging record"
	case "mixed-import-file-coverage":
		return "record mixed import-file data that previously risked skipped rows"
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
		if !strings.HasPrefix(cleaned, ".agents/skills/") &&
			!strings.HasPrefix(cleaned, "skills/openhealth/") {
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
	return strings.Contains(text, "internal/storage/sqlite/sqlc")
}

func outputHasGeneratedResultPath(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "./")
		if strings.HasPrefix(trimmed, "internal/storage/sqlite/sqlc/") {
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
			if fields[i+1] == "migrate" || fields[i+1] == "serve" {
				return true
			}
			if i+2 < len(fields) {
				switch fields[i+1] {
				case "weight":
					if fields[i+2] == "add" || fields[i+2] == "list" {
						return true
					}
				case "blood-pressure":
					if fields[i+2] == "add" || fields[i+2] == "list" || fields[i+2] == "correct" {
						return true
					}
				}
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
