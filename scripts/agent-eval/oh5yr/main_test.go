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
