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
	"time"
	"unicode"

	"github.com/yazanabuashour/openhealth/client"
	storagesqlite "github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

const (
	issueID         = "oh-5yr"
	modelName       = "gpt-5.4-mini"
	reasoningEffort = "medium"
)

type scenario struct {
	ID     string
	Title  string
	Prompt string
}

type variant struct {
	ID    string
	Title string
}

type report struct {
	Issue             string                  `json:"issue"`
	Date              string                  `json:"date"`
	Model             string                  `json:"model"`
	ReasoningEffort   string                  `json:"reasoning_effort"`
	Harness           string                  `json:"harness"`
	CodexVersion      string                  `json:"codex_version"`
	HistoryIsolation  historyIsolationSummary `json:"history_isolation"`
	CommandTemplate   []string                `json:"command_template"`
	CLIStatus         string                  `json:"cli_status"`
	MetricNotes       []string                `json:"metric_notes,omitempty"`
	StopLoss          *stopLossSummary        `json:"stop_loss,omitempty"`
	CodeFirst         *codeFirstSummary       `json:"code_first,omitempty"`
	Results           []runResult             `json:"results"`
	Comparison        *comparisonSummary      `json:"comparison,omitempty"`
	RawLogsCommitted  bool                    `json:"raw_logs_committed"`
	RawLogsNote       string                  `json:"raw_logs_note"`
	TokenUsageCaveat  string                  `json:"token_usage_caveat"`
	AppServerFallback string                  `json:"app_server_fallback"`
}

type historyIsolationSummary struct {
	Status                  string `json:"status"`
	EphemeralFlagRequired   bool   `json:"ephemeral_flag_required"`
	RunDirectoryOutsideRepo bool   `json:"run_directory_outside_repo"`
	NewSessionFilesAfterRun int    `json:"new_session_files_after_run"`
	OpenHealthWorkspaceUsed bool   `json:"openhealth_workspace_used"`
	DesktopAppUsed          bool   `json:"desktop_app_used"`
	VerificationMethod      string `json:"verification_method"`
	VerificationLimitation  string `json:"verification_limitation,omitempty"`
}

type runResult struct {
	Variant                 string             `json:"variant"`
	Scenario                string             `json:"scenario"`
	ScenarioTitle           string             `json:"scenario_title"`
	Passed                  bool               `json:"passed"`
	ExitCode                int                `json:"exit_code"`
	WallSeconds             float64            `json:"wall_seconds"`
	Metrics                 metrics            `json:"metrics"`
	Verification            verificationResult `json:"verification"`
	PromptSummary           string             `json:"prompt_summary"`
	RawLogArtifactReference string             `json:"raw_log_artifact_reference"`
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
	Passed        bool          `json:"passed"`
	DatabasePass  bool          `json:"database_pass"`
	AssistantPass bool          `json:"assistant_pass"`
	Details       string        `json:"details"`
	Weights       []weightState `json:"weights"`
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

type codexEvent struct {
	Type  string `json:"type"`
	Item  item   `json:"item"`
	Usage *usage `json:"usage"`
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

func runCommand(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	runRootFlag := fs.String("run-root", "", "directory for raw run artifacts outside the repo")
	dateFlag := fs.String("date", time.Now().Format(time.DateOnly), "report date in YYYY-MM-DD form")
	compareToFlag := fs.String("compare-to", "", "optional baseline JSON report path for comparison")
	variantFilter := fs.String("variant", "", "optional comma-separated variant ids to run")
	scenarioFilter := fs.String("scenario", "", "optional comma-separated scenario ids to run")
	candidateVariantFlag := fs.String("candidate", "agentops-code", "candidate variant id to compare directly with cli")
	if err := fs.Parse(args); err != nil {
		failf("parse flags: %v", err)
	}
	if fs.NArg() != 0 {
		failf("run does not accept positional arguments")
	}

	repoRoot, err := repoRoot()
	if err != nil {
		failf("resolve repo root: %v", err)
	}

	runRoot := *runRootFlag
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
	selectedVariants, err := selectVariants(*variantFilter)
	if err != nil {
		failf("select variants: %v", err)
	}
	selectedScenarios, err := selectScenarios(*scenarioFilter)
	if err != nil {
		failf("select scenarios: %v", err)
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
	results := []runResult{}
	for _, currentVariant := range selectedVariants {
		for _, currentScenario := range selectedScenarios {
			result, err := runOne(repoRoot, runRoot, currentVariant, currentScenario)
			if err != nil {
				result = runResult{
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
			results = append(results, result)
		}
	}

	newSessionFiles := countNewSessionFiles(markerInfo.ModTime(), runRoot)
	historyStatus := "passed"
	limitation := ""
	if newSessionFiles != 0 {
		historyStatus = "review"
		limitation = "A session-file count changed while evals ran; this may be from another Codex process, because the harness uses --ephemeral and a throwaway cwd."
	}

	outDir := filepath.Join(repoRoot, "docs", "agent-eval-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		failf("create result directory: %v", err)
	}
	jsonPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.json", issueID, *dateFlag))
	mdPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.md", issueID, *dateFlag))

	outReport := report{
		Issue:           issueID,
		Date:            *dateFlag,
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		Harness:         "codex exec --json --ephemeral --full-auto from throwaway run directories",
		CodexVersion:    codexVersion,
		HistoryIsolation: historyIsolationSummary{
			Status:                  historyStatus,
			EphemeralFlagRequired:   true,
			RunDirectoryOutsideRepo: true,
			NewSessionFilesAfterRun: newSessionFiles,
			OpenHealthWorkspaceUsed: false,
			DesktopAppUsed:          false,
			VerificationMethod:      "The runner required --ephemeral, used -C <run-root>/.../repo, kept raw logs under <run-root>, and counted new Codex session files that referenced the throwaway run root.",
			VerificationLimitation:  limitation,
		},
		CommandTemplate: []string{
			"OPENHEALTH_DATABASE_PATH=<run-root>/<variant>/<scenario>/openhealth.db",
			"GOCACHE=<run-root>/<variant>/<scenario>/gocache",
			"GOMODCACHE=<run-root>/<variant>/<scenario>/gomodcache (prewarmed with go mod download before agent execution)",
			"codex exec --json --ephemeral --full-auto --skip-git-repo-check --add-dir <run-root>/<variant>/<scenario> -C <run-root>/<variant>/<scenario>/repo -m gpt-5.4-mini -c model_reasoning_effort=\"medium\" -c shell_environment_policy.inherit=all <natural user prompt>",
		},
		CLIStatus:         "runnable: cli variant uses go run ./cmd/openhealth weight add/list with a prewarmed per-scenario module cache",
		MetricNotes:       metricNotes(*dateFlag, results),
		StopLoss:          productionStopLoss(results),
		CodeFirst:         codeFirstSummaryFor(results, *candidateVariantFlag),
		Results:           results,
		RawLogsCommitted:  false,
		RawLogsNote:       "Raw codex exec event logs and stderr files were retained under <run-root> during execution and intentionally not committed.",
		TokenUsageCaveat:  "Token metrics come from codex exec turn.completed usage events when exposed; unavailable usage must be recorded as not_exposed.",
		AppServerFallback: "not used: codex exec --json exposed enough event detail for this run",
	}
	baseline, baselineRef, err := baselineReport(repoRoot, outDir, jsonPath, *compareToFlag)
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

func runOne(repoRoot string, runRoot string, currentVariant variant, currentScenario scenario) (runResult, error) {
	runDir := filepath.Join(runRoot, currentVariant.ID, currentScenario.ID)
	runRepo := filepath.Join(runDir, "repo")
	dbPath := filepath.Join(runDir, "openhealth.db")
	eventsPath := filepath.Join(runDir, "events.jsonl")
	stderrPath := filepath.Join(runDir, "stderr.log")

	if err := prepareRunDir(runDir); err != nil {
		return runResult{}, fmt.Errorf("prepare run dir: %w", err)
	}
	if err := copyRepo(repoRoot, runRepo); err != nil {
		return runResult{}, fmt.Errorf("copy repo: %w", err)
	}
	if err := installVariant(repoRoot, runRepo, currentVariant); err != nil {
		return runResult{}, fmt.Errorf("install variant: %w", err)
	}
	if err := warmGoModules(runRepo, runDir, dbPath); err != nil {
		return runResult{}, fmt.Errorf("warm go modules: %w", err)
	}
	if err := seedScenario(dbPath, currentScenario); err != nil {
		return runResult{}, fmt.Errorf("seed scenario: %w", err)
	}

	stdoutFile, err := os.Create(eventsPath)
	if err != nil {
		return runResult{}, err
	}
	defer func() {
		_ = stdoutFile.Close()
	}()
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return runResult{}, err
	}
	defer func() {
		_ = stderrFile.Close()
	}()

	args := []string{
		"exec",
		"--json",
		"--ephemeral",
		"--full-auto",
		"--skip-git-repo-check",
		"--add-dir", runDir,
		"-C", runRepo,
		"-m", modelName,
		"-c", fmt.Sprintf("model_reasoning_effort=%q", reasoningEffort),
		"-c", "shell_environment_policy.inherit=all",
		currentScenario.Prompt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Stdin = strings.NewReader("")
	cmd.Env = evalEnv(runDir, dbPath)

	start := time.Now()
	err = cmd.Run()
	wallSeconds := time.Since(start).Seconds()
	exitCode := commandExitCode(err)
	if ctx.Err() == context.DeadlineExceeded {
		exitCode = -1
	}

	parsedMetrics, parseErr := parseMetrics(eventsPath)
	if parseErr != nil {
		parsedMetrics.metrics.CommandMetricLimitations = fmt.Sprintf("failed to parse event log: %v", parseErr)
	}
	verification, verifyErr := verifyScenario(dbPath, currentScenario, parsedMetrics.finalMessage)
	if verifyErr != nil {
		verification = verificationResult{
			Passed:  false,
			Details: fmt.Sprintf("verification error: %v", verifyErr),
		}
	}

	result := runResult{
		Variant:                 currentVariant.ID,
		Scenario:                currentScenario.ID,
		ScenarioTitle:           currentScenario.Title,
		Passed:                  err == nil && verifyErr == nil && verification.Passed,
		ExitCode:                exitCode,
		WallSeconds:             roundSeconds(wallSeconds),
		Metrics:                 parsedMetrics.metrics,
		Verification:            verification,
		PromptSummary:           promptSummary(currentScenario),
		RawLogArtifactReference: fmt.Sprintf("<run-root>/%s/%s/events.jsonl", currentVariant.ID, currentScenario.ID),
	}
	runSummaryPath := filepath.Join(runDir, "run-summary.json")
	_ = writeJSON(runSummaryPath, result)
	return result, nil
}

func evalEnv(runDir string, dbPath string) []string {
	env := os.Environ()
	paths := evalPathsFor(runDir)
	env = append(env,
		"OPENHEALTH_DATABASE_PATH="+dbPath,
		"OPENHEALTH_DATA_DIR=",
		"GOCACHE="+paths.GoCache,
		"GOMODCACHE="+paths.GoModCache,
		"TMPDIR="+paths.Temp,
	)
	return env
}

func warmGoModules(runRepo string, runDir string, dbPath string) error {
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = runRepo
	cmd.Env = evalEnv(runDir, dbPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

type evalPaths struct {
	GoCache    string
	GoModCache string
	Temp       string
}

func evalPathsFor(runDir string) evalPaths {
	return evalPaths{
		GoCache:    filepath.Join(runDir, "gocache"),
		GoModCache: filepath.Join(runDir, "gomodcache"),
		Temp:       filepath.Join(runDir, "tmp"),
	}
}

func prepareRunDir(runDir string) error {
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
	paths := evalPathsFor(runDir)
	for _, dir := range []string{paths.GoCache, paths.GoModCache, paths.Temp} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func variants() []variant {
	return []variant{
		{ID: "production", Title: "Production SDK skill"},
		{ID: "generated-client", Title: "Generated-client baseline skill"},
		{ID: "agentops-code", Title: "Code-first AgentOps facade skill"},
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

func verifyScenario(dbPath string, sc scenario, finalMessage string) (verificationResult, error) {
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

type parsedMetrics struct {
	metrics      metrics
	finalMessage string
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
	case "docs/agent-evals.md", "docs/agent-eval-assets", "docs/agent-eval-results", "scripts/agent-eval":
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
	case "generated-client":
		src = filepath.Join(repoRoot, "docs", "agent-eval-assets", "variants", "generated-client")
	case "agentops-code":
		src = filepath.Join(repoRoot, "docs", "agent-eval-assets", "variants", "agentops-code")
	case "cli":
		src = filepath.Join(repoRoot, "docs", "agent-eval-assets", "variants", "cli")
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	if err := copyDir(src, dest); err != nil {
		return err
	}
	if currentVariant.ID == "generated-client" || currentVariant.ID == "agentops-code" || currentVariant.ID == "cli" {
		if err := normalizeSkillName(filepath.Join(dest, "SKILL.md"), "openhealth"); err != nil {
			return err
		}
	}
	if currentVariant.ID == "agentops-code" {
		return appendAgentOpsVariantInstructions(filepath.Join(runRepo, "AGENTS.md"))
	}
	return nil
}

func appendAgentOpsVariantInstructions(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.WriteString(agentOpsVariantInstructions())
	return err
}

func agentOpsVariantInstructions() string {
	return `

## AgentOps Code Eval Variant

For OpenHealth weight tasks in this isolated eval, use the code-first
agentops facade directly. Do not run bd prime. Do not inspect .agents files,
generated files, the Go module cache, SQLite, or repo-wide search/listing
commands. Do not use the openhealth CLI.

If the request has an ambiguous short date, a non-positive value, or a unit
other than lb/lbs/pound/pounds, reject it directly without running code.

For valid write/list requests, run a single shell command that creates a
temporary Go module outside the repository, uses:

  require github.com/yazanabuashour/openhealth v0.0.0
  replace github.com/yazanabuashour/openhealth => $repo

then imports github.com/yazanabuashour/openhealth/agentops and calls
agentops.RunWeightTask(context.Background(), client.LocalConfig{}, request).
Run the temporary module with exactly:

  (cd "$tmp" && GOPROXY=off GOSUMDB=off go run -mod=mod .)

Do not retry with module-cache inspection, repo search, or other Go command
shapes. For writes, use agentops.WeightTaskRequest{Action:
agentops.WeightTaskActionUpsert, Weights: []agentops.WeightInput{{Date:
"2026-03-29", Value: 152.2, Unit: "lb"}}}. For bounded ranges, use Action:
agentops.WeightTaskActionList, ListMode: agentops.WeightListModeRange,
FromDate: "2026-03-29", ToDate: "2026-03-30". Print JSON and answer only from
the JSON entries/writes/rejection fields. When reporting entries, convert them
to plain newest-first rows such as "2026-03-30 151.6 lb"; for bounded ranges,
include every JSON entry and do not mention excluded dates.
`
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
		if result.Metrics.BroadRepoSearch && isRoutineWeightScenario(result.Scenario) {
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
		recommendation = "pivot_to_cli_for_agent_operations"
	}
	return &stopLossSummary{
		Policy:         "After production hardening, pivot if production loses correctness, directly inspects generated files, inspects the module cache, uses broad repo search in more than one routine scenario, uses the openhealth CLI, uses direct SQLite access, exceeds core tool thresholds, or uses more than 2x CLI tools in at least three comparable scenarios.",
		Triggered:      len(triggers) > 0,
		Recommendation: recommendation,
		Triggers:       triggers,
	}
}

func codeFirstSummaryFor(results []runResult, candidateVariant string) *codeFirstSummary {
	const baselineVariant = "cli"

	if strings.TrimSpace(candidateVariant) == "" {
		candidateVariant = "agentops-code"
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
	totalCandidateTools := 0
	totalCLITools := 0
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
		if candidate.Metrics.BroadRepoSearch && isRoutineWeightScenario(candidate.Scenario) {
			noRoutineBroadSearch = false
		}
		if candidate.Metrics.CLIUsed {
			noCLIUsage = false
		}
		if candidate.Metrics.DirectSQLiteAccess {
			noDirectSQLite = false
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
			if isRoutineWeightScenario(scenario) && candidate.Metrics.ToolCalls > cli.Metrics.ToolCalls+1 {
				routineExceedsCLIByMoreThanOne = append(routineExceedsCLIByMoreThanOne, scenario)
			}
		}
		entries = append(entries, row)
	}
	requiredAtOrBelowCLI := requiredScenariosAtOrBelowCLI(len(candidateScenarioIDs))

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
			Details: fmt.Sprintf("%s must not use broad repo search in routine weight scenarios", candidateVariant),
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
	}

	beatsCLI := true
	for _, criterion := range criteria {
		if !criterion.Passed {
			beatsCLI = false
			break
		}
	}
	recommendation := "continue_cli_for_routine_weight_operations"
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
	if candidateScenarios <= 7 {
		return min(5, candidateScenarios)
	}
	return candidateScenarios - 2
}

func recommendationForCandidate(candidateVariant string) string {
	if candidateVariant == "production" {
		return "prefer_agentops_production_for_routine_weight_operations"
	}
	return "prefer_agentops_code_for_routine_weight_operations"
}

func scenarioIDs() []string {
	ids := []string{}
	for _, scenario := range scenarios() {
		ids = append(ids, scenario.ID)
	}
	return ids
}

func isRoutineWeightScenario(id string) bool {
	switch id {
	case "add-two", "repeat-add", "update-existing", "bounded-range", "bounded-range-natural", "latest-only", "history-limit-two":
		return true
	default:
		return false
	}
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
	fmt.Fprintf(&b, "Reduced JSON artifact: `docs/agent-eval-results/%s-%s.json`\n\n", value.Issue, value.Date)
	fmt.Fprintf(&b, "Raw logs: not committed. They were retained under `<run-root>` during execution and are referenced below only with neutral placeholders.\n\n")

	fmt.Fprintf(&b, "## History Isolation\n\n")
	fmt.Fprintf(&b, "- Status: `%s`\n", value.HistoryIsolation.Status)
	fmt.Fprintf(&b, "- Every agent run used `codex exec --ephemeral` from `<run-root>/<variant>/<scenario>/repo`.\n")
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

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
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
	return strings.Contains(strings.ToLower(trimmed), " lb")
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
			case "weight", "migrate", "serve":
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
		if strings.Contains(field, "openhealth-oh-5yr-") {
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
