package openhealthskill_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSkillMarkdownLinksReferenceInstalledFiles(t *testing.T) {
	t.Parallel()

	markdownFiles := []string{}
	if err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		markdownFiles = append(markdownFiles, path)
		return nil
	}); err != nil {
		t.Fatalf("walk markdown files: %v", err)
	}

	linkPattern := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	for _, path := range markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
			target := match[1]
			if shouldSkipLinkTarget(target) {
				continue
			}
			targetPath := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if _, err := os.Stat(targetPath); err != nil {
				t.Fatalf("%s link target %q is not installed with the skill: %v", path, target, err)
			}
		}
	}
}

func TestProductionSkillUsesAgentOpsForRoutineUserData(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("SKILL.md")
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"agentops.RunWeightTask",
		"agentops.RunBloodPressureTask",
		"github.com/yazanabuashour/openhealth/agentops",
		"GOPROXY=off GOSUMDB=off go run -mod=mod .",
		"references/weights.md",
		"references/blood-pressure.md",
		"AgentOps `entries` are already newest-first",
		"2026/03/31",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"go run ./cmd/openhealth",
		"CLI fallback",
		"Generated Client Fallback",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("skill contains forbidden text %q", forbidden)
		}
	}
}

func TestBloodPressureReferenceDocumentsCorrection(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("references", "blood-pressure.md"))
	if err != nil {
		t.Fatalf("read blood pressure reference: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"BloodPressureTaskActionCorrect",
		"updates exactly one existing reading",
		"multiple same-date readings",
		"rejection_reason",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("blood pressure reference missing %q", want)
		}
	}
}

func shouldSkipLinkTarget(target string) bool {
	return target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		filepath.IsAbs(target)
}
