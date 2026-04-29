package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

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
