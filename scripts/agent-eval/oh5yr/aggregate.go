package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

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
