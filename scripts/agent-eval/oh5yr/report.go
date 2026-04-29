package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
