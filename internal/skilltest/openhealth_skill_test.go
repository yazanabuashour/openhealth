package skilltest_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestOpenHealthSkillPayloadContainsOnlySkillMarkdown(t *testing.T) {
	t.Parallel()

	skillDir := openHealthSkillDir(t)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		t.Fatalf("read skill dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "SKILL.md" || entries[0].IsDir() {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("skill payload files = %v, want exactly SKILL.md", names)
	}
}

func TestOpenHealthSkillMarkdownLinksReferenceInstalledFiles(t *testing.T) {
	t.Parallel()

	skillDir := openHealthSkillDir(t)
	content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	linkPattern := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
		target := match[1]
		if shouldSkipLinkTarget(target) {
			continue
		}
		targetPath := filepath.Clean(filepath.Join(skillDir, target))
		if _, err := os.Stat(targetPath); err != nil {
			t.Fatalf("SKILL.md link target %q is not installed with the skill: %v", target, err)
		}
	}
}

func TestOpenHealthSkillUsesInstalledAgentOpsBinary(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(openHealthSkillDir(t), "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"agentops.RunWeightTask",
		"agentops.RunBloodPressureTask",
		"agentops.RunMedicationTask",
		"agentops.RunLabTask",
		"openhealth-agentops weight",
		"openhealth-agentops blood-pressure",
		"openhealth-agentops medications",
		"openhealth-agentops labs",
		"upsert_weights",
		"record_blood_pressure",
		"correct_blood_pressure",
		"record_medications",
		"correct_medication",
		"delete_medication",
		"record_labs",
		"correct_labs",
		"delete_labs",
		"Do not run repo-wide file discovery or broad searches",
		"reject directly without running code",
		"AgentOps `entries` are already newest-first",
		"2026/03/31",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"go run ./cmd/openhealth-agentops",
		"go run ./cmd/openhealth",
		"references/",
		"temporary Go module",
		"GOPROXY=off",
		"go run -mod=mod",
		"CLI fallback",
		"Generated Client Fallback",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("skill contains forbidden text %q", forbidden)
		}
	}
}

func openHealthSkillDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "skills", "openhealth"))
}

func shouldSkipLinkTarget(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		filepath.IsAbs(target)
}
