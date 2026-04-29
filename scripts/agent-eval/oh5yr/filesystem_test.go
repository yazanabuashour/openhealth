package main

import (
	"bytes"
	"os"
	"path/filepath"
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

	cache := cacheConfig{Mode: cacheModeIsolated, RunRoot: t.TempDir()}
	if err := prepareRunDir(runDir, cache); err != nil {
		t.Fatalf("prepareRunDir() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "openhealth.db")); !os.IsNotExist(err) {
		t.Fatalf("stale database stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(runDir, "repo")); !os.IsNotExist(err) {
		t.Fatalf("stale repo stat error = %v, want not exist", err)
	}

	paths := evalPathsFor(runDir, cache)
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

func TestCopyRepoSkipsVariantContaminatingInstructions(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	src := filepath.Join(temp, "src")
	dst := filepath.Join(temp, "dst")
	for _, path := range []string{
		filepath.Join(src, ".agents", "skills", "openhealth"),
		filepath.Join(src, "docs", "agent-eval-results"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for path, content := range map[string]string{
		filepath.Join(src, "AGENTS.md"):                                     "repo agent instructions",
		filepath.Join(src, "README.md"):                                     "kept",
		filepath.Join(src, ".agents", "skills", "openhealth", "SKILL.md"):   "stale skill",
		filepath.Join(src, "docs", "agent-eval-results", "previous.md"):     "previous report",
		filepath.Join(src, "docs", "agent-evals.md"):                        "eval docs",
		filepath.Join(src, "scripts", "agent-eval", "oh5yr", "main.go"):     "harness",
		filepath.Join(src, "docs", "agent-eval-assets", "variants", "x.md"): "asset",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := copyRepo(src, dst); err != nil {
		t.Fatalf("copyRepo() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Fatalf("kept file stat error = %v", err)
	}
	for _, skipped := range []string{
		"AGENTS.md",
		filepath.Join(".agents", "skills", "openhealth", "SKILL.md"),
		filepath.Join("docs", "agent-eval-results", "previous.md"),
		filepath.Join("docs", "agent-evals.md"),
		filepath.Join("scripts", "agent-eval", "oh5yr", "main.go"),
		filepath.Join("docs", "agent-eval-assets", "variants", "x.md"),
	} {
		if _, err := os.Stat(filepath.Join(dst, skipped)); !os.IsNotExist(err) {
			t.Fatalf("copied skipped path %s: stat error = %v", skipped, err)
		}
	}
}

func TestInstallVariantInstallsExactProductionSkillWithoutAgentsFile(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	repoRoot := filepath.Join(temp, "src")
	runRepo := filepath.Join(temp, "run")
	sourceSkillDir := filepath.Join(repoRoot, "skills", "openhealth")
	if err := os.MkdirAll(sourceSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceSkill := []byte("---\nname: openhealth\ndescription: test\n---\n# Skill\n")
	if err := os.WriteFile(filepath.Join(sourceSkillDir, "SKILL.md"), sourceSkill, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installVariant(repoRoot, runRepo, variant{ID: "production"}); err != nil {
		t.Fatalf("installVariant: %v", err)
	}
	installed, err := os.ReadFile(filepath.Join(runRepo, ".agents", "skills", "openhealth", "SKILL.md"))
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	if !bytes.Equal(installed, sourceSkill) {
		t.Fatalf("installed skill = %q, want exact source skill", installed)
	}
	if _, err := os.Stat(filepath.Join(runRepo, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md stat error = %v, want not exist", err)
	}
}

func TestPromptInputPreflightFlagsOpenHealthAgentsInstructions(t *testing.T) {
	t.Parallel()

	clean := `{"text":"<skills_instructions>- openhealth: Use this skill. (file: /tmp/run/repo/.agents/skills/openhealth/SKILL.md)</skills_instructions>"}`
	if containsOpenHealthAgentsInstructions(clean) {
		t.Fatalf("clean rendered prompt flagged as contaminated")
	}
	contaminated := `{"text":"# AGENTS.md instructions for /tmp/run/repo\n\n<INSTRUCTIONS>\nFor valid tasks, pipe JSON to openhealth weight.\n{\"action\":\"upsert_weights\"}\n</INSTRUCTIONS>"}`
	if !containsOpenHealthAgentsInstructions(contaminated) {
		t.Fatalf("contaminated rendered prompt was not flagged")
	}
}

func TestShouldSkipEvalPath(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"docs/agent-evals.md",
		"docs/agent-eval-assets",
		"docs/agent-eval-assets/legacy/old.md",
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
