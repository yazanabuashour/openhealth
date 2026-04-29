package main

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
