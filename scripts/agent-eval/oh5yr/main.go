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
	Results           []runResult             `json:"results"`
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
	AssistantCalls           int            `json:"assistant_calls"`
	ToolCalls                int            `json:"tool_calls"`
	CommandExecutions        int            `json:"command_executions"`
	FileInspectionCommands   int            `json:"file_inspection_commands"`
	GeneratedFileInspected   bool           `json:"generated_file_inspected"`
	ModuleCacheInspected     bool           `json:"module_cache_inspected"`
	UsageExposed             bool           `json:"usage_exposed"`
	InputTokens              *int           `json:"input_tokens,omitempty"`
	CachedInputTokens        *int           `json:"cached_input_tokens,omitempty"`
	NonCachedInputTokens     *int           `json:"non_cached_input_tokens,omitempty"`
	OutputTokens             *int           `json:"output_tokens,omitempty"`
	EventTypeCounts          map[string]int `json:"event_type_counts"`
	CommandMetricLimitations string         `json:"command_metric_limitations"`
}

type verificationResult struct {
	Passed        bool          `json:"passed"`
	DatabasePass  bool          `json:"database_pass"`
	AssistantPass bool          `json:"assistant_pass"`
	Details       string        `json:"details"`
	Weights       []weightState `json:"weights"`
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
	for _, currentVariant := range variants() {
		for _, currentScenario := range scenarios() {
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
			"GOMODCACHE=<run-root>/<variant>/<scenario>/gomodcache",
			"codex exec --json --ephemeral --full-auto --skip-git-repo-check --add-dir <run-root>/<variant>/<scenario> -C <run-root>/<variant>/<scenario>/repo -m gpt-5.4-mini -c model_reasoning_effort=\"medium\" -c shell_environment_policy.inherit=all <natural user prompt>",
		},
		CLIStatus:         "runnable: cli variant uses go run ./cmd/openhealth weight add/list",
		Results:           results,
		RawLogsCommitted:  false,
		RawLogsNote:       "Raw codex exec event logs and stderr files were retained under <run-root> during execution and intentionally not committed.",
		TokenUsageCaveat:  "Token metrics come from codex exec turn.completed usage events when exposed; unavailable usage must be recorded as not_exposed.",
		AppServerFallback: "not used: codex exec --json exposed enough event detail for this run",
	}

	outDir := filepath.Join(repoRoot, "docs", "agent-eval-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		failf("create result directory: %v", err)
	}
	jsonPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.json", issueID, *dateFlag))
	mdPath := filepath.Join(outDir, fmt.Sprintf("%s-%s.md", issueID, *dateFlag))
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
			Prompt: "Please show my local OpenHealth weight history from 03/29/2026 through 03/30/2026 only, newest first. Use the configured local data path.",
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
	}
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
	case "bounded-range":
		return upsertWeights(ctx, api, []weightState{
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
	case "bounded-range":
		expectedDB := []weightState{
			{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
			{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
			{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
		}
		result.DatabasePass = weightsEqual(states, expectedDB)
		result.AssistantPass = mentionsDatesInOrder(finalMessage, "2026-03-30", "2026-03-29") &&
			!mentionsDate(finalMessage, "2026-03-28")
		result.Details = fmt.Sprintf("expected unchanged seed rows and assistant output limited to 2026-03-29..2026-03-30 newest-first; observed %s%s", describeWeights(states), listErrorDetail)
	case "ambiguous-short-date":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"year", "which year", "clarify", "ambiguous"})
		result.Details = fmt.Sprintf("expected no write and a year clarification; observed %s%s", describeWeights(states), listErrorDetail)
	case "invalid-input":
		result.DatabasePass = len(states) == 0
		result.AssistantPass = containsAny(strings.ToLower(finalMessage), []string{"invalid", "unsupported", "positive", "cannot", "can't", "unit", "value", "lb", "pounds"})
		result.Details = fmt.Sprintf("expected no write and an invalid input rejection; observed %s%s", describeWeights(states), listErrorDetail)
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
				if inspectsGeneratedFileCommand(event.Item.Command, event.Item.AggregatedOutput) {
					out.metrics.GeneratedFileInspected = true
				}
				if inspectsModuleCache(event.Item.Command) {
					out.metrics.ModuleCacheInspected = true
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
	case "cli":
		src = filepath.Join(repoRoot, "docs", "agent-eval-assets", "variants", "cli")
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	if err := copyDir(src, dest); err != nil {
		return err
	}
	if currentVariant.ID == "generated-client" || currentVariant.ID == "cli" {
		return normalizeSkillName(filepath.Join(dest, "SKILL.md"), "openhealth")
	}
	return nil
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
	fmt.Fprintf(&b, "| Variant | Scenario | Result | DB | Assistant | Tools | Assistant Calls | Wall Seconds | Tokens | Generated Files | Module Cache |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- |\n")
	for _, result := range value.Results {
		tokenSummary := "not_exposed"
		if result.Metrics.UsageExposed {
			tokenSummary = fmt.Sprintf("in %d / cached %d / out %d", deref(result.Metrics.InputTokens), deref(result.Metrics.CachedInputTokens), deref(result.Metrics.OutputTokens))
		}
		fmt.Fprintf(
			&b,
			"| `%s` | `%s` | %s | %s | %s | %d | %d | %.2f | %s | %s | %s |\n",
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
			yesNo(result.Metrics.ModuleCacheInspected),
		)
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
	case "ambiguous-short-date":
		return "ask for year before writing 03/29 152.2 lb"
	case "invalid-input":
		return "reject -5 stone for 2026-03-31"
	default:
		return sc.Title
	}
}

func isFileInspectionCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "sed", "cat", "find", "ls", "awk", "head", "tail", "nl")
}

func inspectsGeneratedFileCommand(command string, output string) bool {
	if !isFileInspectionCommand(command) || isBroadFileListingCommand(command) {
		return false
	}
	if mentionsGeneratedPath(command) {
		return true
	}
	return isContentSearchCommand(command) && mentionsGeneratedPath(output)
}

func isBroadFileListingCommand(command string) bool {
	normalized := normalizedCommandText(command)
	return (commandHasExecutable(command, "rg") && strings.Contains(normalized, " --files ")) ||
		commandHasExecutable(command, "find", "ls")
}

func isContentSearchCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "awk")
}

func mentionsGeneratedPath(text string) bool {
	return strings.Contains(text, "client.gen.go") ||
		strings.Contains(text, "internal/api/generated") ||
		strings.Contains(text, "server.gen.go")
}

func inspectsModuleCache(command string) bool {
	lower := strings.ToLower(command)
	return strings.Contains(lower, "gomodcache") ||
		strings.Contains(lower, "pkg/mod") ||
		strings.Contains(lower, "go env")
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
