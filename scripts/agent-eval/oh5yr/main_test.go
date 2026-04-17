package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareRunDirResetsAndCreatesRuntimeDirs(t *testing.T) {
	t.Parallel()

	runDir := filepath.Join(t.TempDir(), "production", "add-two")
	if err := os.MkdirAll(filepath.Join(runDir, "repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "openhealth.db"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := prepareRunDir(runDir); err != nil {
		t.Fatalf("prepareRunDir() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "openhealth.db")); !os.IsNotExist(err) {
		t.Fatalf("stale database stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "repo")); !os.IsNotExist(err) {
		t.Fatalf("stale repo stat error = %v, want not exist", err)
	}

	paths := evalPathsFor(runDir)
	for _, dir := range []string{runDir, paths.GoCache, paths.GoModCache, paths.Temp} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
}

func TestVariantsIncludeCLI(t *testing.T) {
	t.Parallel()

	ids := map[string]bool{}
	for _, variant := range variants() {
		ids[variant.ID] = true
	}
	for _, want := range []string{"production", "generated-client", "cli"} {
		if !ids[want] {
			t.Fatalf("variants() missing %q: %#v", want, variants())
		}
	}
}

func TestWriteMarkdownReportsRunnableCLIStatus(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.md")
	if err := writeMarkdown(path, report{
		Issue:           issueID,
		Date:            "2026-04-16",
		Harness:         "test",
		Model:           modelName,
		ReasoningEffort: reasoningEffort,
		CodexVersion:    "test",
		CLIStatus:       "runnable: cli variant uses go run ./cmd/openhealth weight add/list",
	}); err != nil {
		t.Fatalf("writeMarkdown: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Status: `runnable: cli variant uses go run ./cmd/openhealth weight add/list`.") {
		t.Fatalf("markdown = %q, want runnable CLI status", text)
	}
	if strings.Contains(text, "not run because") {
		t.Fatalf("markdown = %q, should not report CLI as skipped", text)
	}
}

func TestShouldSkipEvalPath(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"docs/agent-evals.md",
		"docs/agent-eval-assets",
		"docs/agent-eval-assets/variants/generated-client/SKILL.md",
		"docs/agent-eval-results",
		"docs/agent-eval-results/oh-5yr-2026-04-16.md",
		"scripts/agent-eval",
		"scripts/agent-eval/oh5yr/main.go",
	} {
		if !shouldSkipEvalPath(path) {
			t.Fatalf("shouldSkipEvalPath(%q) = false, want true", path)
		}
	}

	for _, path := range []string{
		"docs/maintainers.md",
		"scripts/validate-agent-skill.sh",
		"skills/openhealth/SKILL.md",
	} {
		if shouldSkipEvalPath(path) {
			t.Fatalf("shouldSkipEvalPath(%q) = true, want false", path)
		}
	}
}

func TestGeneratedFileInspectionIgnoresBroadListings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		output  string
		want    bool
	}{
		{
			name:    "rg files listing",
			command: "/bin/zsh -lc rg --files",
			output:  "client/client.gen.go\ninternal/api/generated/server.gen.go\n",
			want:    false,
		},
		{
			name:    "direct rg files listing",
			command: "rg --files",
			output:  "client/client.gen.go\ninternal/api/generated/server.gen.go\n",
			want:    false,
		},
		{
			name:    "find listing",
			command: "/bin/zsh -lc find . -type f",
			output:  "./client/client.gen.go\n",
			want:    false,
		},
		{
			name:    "direct find listing",
			command: "find . -type f",
			output:  "./client/client.gen.go\n",
			want:    false,
		},
		{
			name:    "direct read",
			command: "/bin/zsh -lc sed -n '1,40p' client/client.gen.go",
			output:  "package client\n",
			want:    true,
		},
		{
			name:    "content search with generated output",
			command: "/bin/zsh -lc rg 'CreateHealthWeight' .",
			output:  "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			want:    true,
		},
		{
			name:    "direct content search with generated output",
			command: "rg 'CreateHealthWeight' .",
			output:  "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			want:    true,
		},
		{
			name:    "direct grep with generated output",
			command: "grep -R CreateHealthWeight .",
			output:  "client/client.gen.go:func (c *Client) CreateHealthWeight(...)\n",
			want:    true,
		},
		{
			name:    "non inspection command",
			command: "/bin/zsh -lc go test ./...",
			output:  "ok github.com/yazanabuashour/openhealth/client\n",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inspectsGeneratedFileCommand(tt.command, tt.output); got != tt.want {
				t.Fatalf("inspectsGeneratedFileCommand(%q, %q) = %v, want %v", tt.command, tt.output, got, tt.want)
			}
		})
	}
}

func TestMentionsDatesInOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "newest first iso",
			message: "Stored rows: 2026-03-30 151.6 lb, 2026-03-29 152.2 lb.",
			want:    true,
		},
		{
			name:    "oldest first iso",
			message: "Stored rows: 2026-03-29 152.2 lb, 2026-03-30 151.6 lb.",
			want:    false,
		},
		{
			name:    "newest first short dates",
			message: "03/30: 151.6 lb; 03/29: 152.2 lb.",
			want:    true,
		},
		{
			name:    "missing date",
			message: "Only 2026-03-30 is present.",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mentionsDatesInOrder(tt.message, "2026-03-30", "2026-03-29")
			if got != tt.want {
				t.Fatalf("mentionsDatesInOrder(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestBoundedRangeAssistantPass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name: "allows excluded date in exclusion sentence",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"No entries from 03/28/2026 are included.",
			}, "\n"),
			want: true,
		},
		{
			name: "ignores natural date order in prose before result rows",
			message: strings.Join([]string{
				"Here are the weights for March 29 and March 30, 2026:",
				"",
				"- 2026-03-30 12:00:00Z: 151.6 lb",
				"- 2026-03-29 12:00:00Z: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name: "allows bounded prose with bullet result rows",
			message: strings.Join([]string{
				"Using the configured local database, the weights for March 29 and March 30, 2026, newest first:",
				"",
				"- 2026-03-30: 151.6 lb",
				"- 2026-03-29: 152.2 lb",
			}, "\n"),
			want: true,
		},
		{
			name:    "allows compact same-line result rows",
			message: "2026-03-30: 151.6 lb; 2026-03-29: 152.2 lb",
			want:    true,
		},
		{
			name: "rejects excluded date as result row",
			message: strings.Join([]string{
				"1. 2026-03-30 151.6 lb",
				"2. 2026-03-29 152.2 lb",
				"3. 2026-03-28 153.0 lb",
			}, "\n"),
			want: false,
		},
		{
			name:    "rejects missing newest date",
			message: "1. 2026-03-29 152.2 lb",
			want:    false,
		},
		{
			name: "rejects oldest first",
			message: strings.Join([]string{
				"1. 2026-03-29 152.2 lb",
				"2. 2026-03-30 151.6 lb",
			}, "\n"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := boundedRangeAssistantPass(tt.message); got != tt.want {
				t.Fatalf("boundedRangeAssistantPass(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestNaturalBoundedRangeScenarioSeedsExpectedRows(t *testing.T) {
	t.Parallel()

	sc, ok := scenarioByID("bounded-range-natural")
	if !ok {
		t.Fatal("missing bounded-range-natural scenario")
	}
	if !strings.Contains(sc.Prompt, "Mar 29") || !strings.Contains(sc.Prompt, "Mar 30") {
		t.Fatalf("prompt = %q, want natural Mar 29 and Mar 30 wording", sc.Prompt)
	}

	databasePath := filepath.Join(t.TempDir(), "openhealth.db")
	if err := seedScenario(databasePath, sc); err != nil {
		t.Fatalf("seedScenario: %v", err)
	}
	weights, err := listWeights(databasePath)
	if err != nil {
		t.Fatalf("listWeights: %v", err)
	}
	got := weightStates(weights)
	want := []weightState{
		{Date: "2026-03-30", Value: 151.6, Unit: "lb"},
		{Date: "2026-03-29", Value: 152.2, Unit: "lb"},
		{Date: "2026-03-28", Value: 153.0, Unit: "lb"},
	}
	if !weightsEqual(got, want) {
		t.Fatalf("seeded weights = %s, want %s", describeWeights(got), describeWeights(want))
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
