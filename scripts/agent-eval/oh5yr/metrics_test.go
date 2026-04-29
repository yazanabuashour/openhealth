package main

import (
	"strings"
	"testing"
)

func TestSanitizeMetricEvidenceRedactsCustomRunRoots(t *testing.T) {
	t.Parallel()

	command := "/bin/zsh -lc 'go run ./cmd/openhealth weight list --db /tmp/openhealth-oh743-final-r1/cli/latest-only/openhealth.db'"
	got := sanitizeMetricEvidence(command)
	if strings.Contains(got, "/tmp/") || strings.Contains(got, "openhealth-oh743") {
		t.Fatalf("sanitizeMetricEvidence() = %q, want run root redacted", got)
	}
	if !strings.Contains(got, "<run-root>") {
		t.Fatalf("sanitizeMetricEvidence() = %q, want <run-root>", got)
	}
}

func TestAggregateMetricsRequiresEveryTurnUsage(t *testing.T) {
	t.Parallel()

	first := turnResult{Metrics: testMetrics(1, 50)}
	second := turnResult{Metrics: testMetrics(2, 70)}
	got := aggregateMetrics([]turnResult{first, second})
	if got.ToolCalls != 3 || got.CommandExecutions != 3 || !got.UsageExposed || got.NonCachedInputTokens == nil || *got.NonCachedInputTokens != 120 {
		t.Fatalf("aggregateMetrics = %#v, want summed tools and tokens", got)
	}

	second.Metrics.UsageExposed = false
	got = aggregateMetrics([]turnResult{first, second})
	if got.UsageExposed || got.NonCachedInputTokens != nil {
		t.Fatalf("aggregateMetrics with missing usage = %#v, want aggregate usage hidden", got)
	}
}

func TestGeneratedFileInspectionIgnoresBroadListings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		command         string
		output          string
		wantDirect      bool
		wantBroadSearch bool
		wantBroadGen    bool
	}{
		{
			name:            "rg files listing",
			command:         "/bin/zsh -lc rg --files",
			output:          "internal/storage/sqlite/sqlc/health.sql.go\ninternal/storage/sqlite/sqlc/db.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "direct rg files listing",
			command:         "rg --files",
			output:          "internal/storage/sqlite/sqlc/health.sql.go\ninternal/storage/sqlite/sqlc/db.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "mixed targeted and root rg files listing",
			command:         "rg --files .agents/skills/openhealth repo .",
			output:          ".agents/skills/openhealth/SKILL.md\ninternal/storage/sqlite/sqlc/health.sql.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "find listing",
			command:         "/bin/zsh -lc find . -type f",
			output:          "./internal/storage/sqlite/sqlc/health.sql.go\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "targeted skill file listing",
			command:         "rg --files .agents/skills/openhealth",
			output:          ".agents/skills/openhealth/SKILL.md\n",
			wantDirect:      false,
			wantBroadSearch: false,
			wantBroadGen:    false,
		},
		{
			name:       "direct read",
			command:    "/bin/zsh -lc sed -n '1,40p' internal/storage/sqlite/sqlc/health.sql.go",
			output:     "package sqlc\n",
			wantDirect: true,
		},
		{
			name:       "skill guidance mentions generated file",
			command:    "/bin/zsh -lc sed -n '1,220p' .agents/skills/openhealth/SKILL.md",
			output:     "Do not inspect generated database code\n",
			wantDirect: false,
		},
		{
			name:            "broad content search with generated output",
			command:         "/bin/zsh -lc rg 'ListWeightEntries' .",
			output:          "internal/storage/sqlite/sqlc/health.sql.go:func (q *Queries) ListWeightEntries(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:            "implicit broad content search with generated output",
			command:         "/bin/zsh -lc rg -n 'ListWeightEntries'",
			output:          "internal/storage/sqlite/sqlc/health.sql.go:func (q *Queries) ListWeightEntries(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:       "targeted content search with generated output",
			command:    "rg 'ListWeightEntries' internal/storage/sqlite/sqlc",
			output:     "internal/storage/sqlite/sqlc/health.sql.go:func (q *Queries) ListWeightEntries(...)\n",
			wantDirect: true,
		},
		{
			name:            "direct grep with generated output",
			command:         "grep -R ListWeightEntries .",
			output:          "internal/storage/sqlite/sqlc/health.sql.go:func (q *Queries) ListWeightEntries(...)\n",
			wantDirect:      false,
			wantBroadSearch: true,
			wantBroadGen:    true,
		},
		{
			name:       "non inspection command",
			command:    "/bin/zsh -lc go test ./...",
			output:     "ok github.com/yazanabuashour/openhealth/client\n",
			wantDirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inspectsGeneratedFileCommand(tt.command, tt.output); got != tt.wantDirect {
				t.Fatalf("inspectsGeneratedFileCommand(%q, %q) = %v, want %v", tt.command, tt.output, got, tt.wantDirect)
			}
			if got := isBroadRepoSearchCommand(tt.command); got != tt.wantBroadSearch {
				t.Fatalf("isBroadRepoSearchCommand(%q) = %v, want %v", tt.command, got, tt.wantBroadSearch)
			}
			gotBroadGen := isBroadRepoSearchCommand(tt.command) && mentionsGeneratedPath(tt.output)
			if gotBroadGen != tt.wantBroadGen {
				t.Fatalf("broad generated path metric = %v, want %v", gotBroadGen, tt.wantBroadGen)
			}
		})
	}
}

func TestCLIAndDirectSQLiteMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		command    string
		wantCLI    bool
		wantSQLite bool
	}{
		{
			name:    "go run cli",
			command: "go run ./cmd/openhealth weight list --limit 2",
			wantCLI: true,
		},
		{
			name:    "go run cli after shell setup",
			command: `/bin/zsh -lc 'cd repo && go run ./cmd/openhealth weight list --limit 2'`,
			wantCLI: true,
		},
		{
			name:    "installed cli",
			command: "/usr/local/bin/openhealth weight add --date 2026-03-29 --value 152.2",
			wantCLI: true,
		},
		{
			name:    "search for go run cli text",
			command: `/bin/zsh -lc 'rg -n "go run ./cmd/openhealth" skills/openhealth'`,
		},
		{
			name:    "grep for installed cli text",
			command: `grep -R "openhealth weight" skills/openhealth`,
		},
		{
			name: "temporary Go runner",
			command: `tmp="$(mktemp -d)" && repo="$(pwd)" && cat > "$tmp/go.mod" <<EOF
require github.com/yazanabuashour/openhealth v0.0.0
replace github.com/yazanabuashour/openhealth => $repo
EOF
(cd "$tmp" && GOPROXY=off GOSUMDB=off go run -mod=mod .)`,
		},
		{
			name:    "openhealth json runner",
			command: `openhealth weight <<'EOF'{"action":"list_weights"}EOF`,
		},
		{
			name:       "sqlite executable",
			command:    `sqlite3 "$OPENHEALTH_DATABASE_PATH" "select * from health_weight_entry"`,
			wantSQLite: true,
		},
		{
			name: "python sqlite import",
			command: `python - <<'PY'
import sqlite3
sqlite3.connect("openhealth.db")
PY`,
			wantSQLite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := usesOpenHealthCLI(tt.command); got != tt.wantCLI {
				t.Fatalf("usesOpenHealthCLI(%q) = %v, want %v", tt.command, got, tt.wantCLI)
			}
			if got := usesDirectSQLite(tt.command); got != tt.wantSQLite {
				t.Fatalf("usesDirectSQLite(%q) = %v, want %v", tt.command, got, tt.wantSQLite)
			}
		})
	}
}
