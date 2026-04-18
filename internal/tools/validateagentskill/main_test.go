package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsValidSkillWithYAMLFrontmatter(t *testing.T) {
	t.Parallel()

	skillDir := writeSkill(t, "openhealth", `---
name: openhealth
description: >
  Use this skill for local OpenHealth data:
  weights, blood pressure, medications, and labs.
compatibility: "Requires local filesystem access and an installed openhealth binary on PATH."
license: MIT
---

# OpenHealth

Use:

- openhealth weight
- openhealth blood-pressure
- openhealth medications
- openhealth labs
`)

	var stdout bytes.Buffer
	if err := run([]string{skillDir}, &stdout); err != nil {
		t.Fatalf("run validator: %v", err)
	}
	if !strings.Contains(stdout.String(), "validated ") {
		t.Fatalf("stdout = %q, want validated message", stdout.String())
	}
}

func TestRunRejectsInvalidSkillPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{
			name: "extra file",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openhealth"),
				"notes.md": "extra",
			},
			wantErr: "must contain only SKILL.md",
		},
		{
			name: "missing opening delimiter",
			files: map[string]string{
				"SKILL.md": "name: openhealth\n",
			},
			wantErr: "must start with YAML frontmatter",
		},
		{
			name: "missing closing delimiter",
			files: map[string]string{
				"SKILL.md": "---\nname: openhealth\n",
			},
			wantErr: "must include a closing ---",
		},
		{
			name: "malformed yaml",
			files: map[string]string{
				"SKILL.md": "---\nname: openhealth\ndescription: [unterminated\n---\n",
			},
			wantErr: "must be valid YAML",
		},
		{
			name: "non string metadata",
			files: map[string]string{
				"SKILL.md": "---\nname: openhealth\ndescription: 42\n---\n",
			},
			wantErr: `field "description" must be a string`,
		},
		{
			name: "missing required name",
			files: map[string]string{
				"SKILL.md": "---\ndescription: Use OpenHealth locally.\n---\n",
			},
			wantErr: "must define a non-empty name",
		},
		{
			name: "description too long",
			files: map[string]string{
				"SKILL.md": "---\nname: openhealth\ndescription: " + strings.Repeat("a", 1025) + "\n---\n",
			},
			wantErr: "description must be 1024 characters or fewer",
		},
		{
			name: "missing referenced file",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openhealth") + "\n[Reference](references/foo.md)\n",
			},
			wantErr: "is not installed with the skill",
		},
		{
			name: "retired runner binary name",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openhealth") + "\nRun `openhealth-" + "agent" + "ops weight`.\n",
			},
			wantErr: "retired product binary name",
		},
		{
			name: "retired runner guidance",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openhealth") + "\nUse Agent" + "Ops.\n",
			},
			wantErr: "retired runner guidance",
		},
		{
			name: "retired human cli command",
			files: map[string]string{
				"SKILL.md": validSkillMarkdown("openhealth") + "\nRun `openhealth weight add --date 2026-03-29 --value 152.2`.\n",
			},
			wantErr: "retired human CLI command guidance",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			skillDir := writeSkillFiles(t, "openhealth", tt.files)
			var stdout bytes.Buffer
			err := run([]string{skillDir}, &stdout)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("run error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func validSkillMarkdown(name string) string {
	return `---
name: ` + name + `
description: Use OpenHealth locally.
compatibility: Requires local filesystem access and an installed openhealth binary on PATH.
---

# OpenHealth

Run ` + "`openhealth weight`" + `.
`
}

func writeSkill(t *testing.T, name string, content string) string {
	t.Helper()
	return writeSkillFiles(t, name, map[string]string{"SKILL.md": content})
}

func writeSkillFiles(t *testing.T, name string, files map[string]string) string {
	t.Helper()

	skillDir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	for fileName, content := range files {
		path := filepath.Join(skillDir, fileName)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	return skillDir
}
