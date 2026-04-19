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

func TestOpenHealthSkillUsesInstalledRunner(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(openHealthSkillDir(t), "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"openhealth weight",
		"openhealth body-composition",
		"openhealth blood-pressure",
		"openhealth medications",
		"openhealth labs",
		"openhealth imaging",
		"upsert_weights",
		"record_body_composition",
		"correct_body_composition",
		"delete_body_composition",
		"record_blood_pressure",
		"correct_blood_pressure",
		"record_medications",
		"correct_medication",
		"delete_medication",
		"record_labs",
		"correct_labs",
		"patch_labs",
		"delete_labs",
		"record_imaging",
		"correct_imaging",
		"delete_imaging",
		"Do not run repo-wide file discovery or broad searches",
		"reject directly without running code",
		"Runner `entries` are already newest-first",
		"do not call `--help`",
		"2026/03/31",
		"vitamin-d",
		"hemoglobin-a1c",
		"ferritin",
		"urine-albumin-creatinine-ratio",
		"subcutaneous injection",
		"topical cream",
		"1 patch every 24 hours",
		"systolic not greater than diastolic",
		"result_updates",
		"date is ambiguous",
		"body_fat_percent",
		"Keep scale weight in `openhealth weight`",
		"call `openhealth weight` for the weight and",
		"empty optional text field or note string",
		"Optional text fields cannot be cleared",
		"optional `note`",
		"narrowest matching note field",
		"`results[].notes`",
		"Note arrays preserve order and multiline text",
		"weight or\nblood-pressure `note`",
		"XR TOE RIGHT narrative",
		"HIV 4th gen narrative",
		"summary",
		"impression",
		"modality",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("skill missing %q", want)
		}
	}
	for _, forbidden := range []string{
		"go run ./cmd/openhealth",
		"agent" + "ops",
		"Agent" + "Ops",
		"openhealth-" + "agent" + "ops",
		"openhealth weight add",
		"openhealth weight list",
		"openhealth blood-pressure add",
		"openhealth blood-pressure list",
		"openhealth blood-pressure correct",
		"references/",
		"temporary Go module",
		"GOPROXY=off",
		"go run -mod=mod",
		"CLI fallback",
		"Generated Client Fallback",
		"generated files",
		"unsupported lab " + "analyte " + "slug",
		"Supported `canonical_" + "slug` and `analyte_" + "slug` values",
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
