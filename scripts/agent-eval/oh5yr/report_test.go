package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteMarkdownOmitsCLIVariantSection(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, report{
		Issue:           issueID,
		Date:            "2026-04-16",
		Harness:         "test",
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		CodexVersion:    "test",
	}); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "CLI-Oriented Variant") || strings.Contains(text, "Code-First CLI Comparison") {
		t.Fatalf("markdown = %q, should not include retired CLI sections", text)
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
	if summary.Recommendation != "continue_production_hardening" {
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
	if summary.Recommendation != "ship_openhealth_runner_production" {
		t.Fatalf("recommendation = %q", summary.Recommendation)
	}
}
