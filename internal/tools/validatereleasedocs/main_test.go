package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsValidReleaseDocs(t *testing.T) {
	root := writeReleaseRepo(t, "v0.3.0", validReleaseNotes("v0.3.0"), changelogFor("v0.3.0"))
	var stdout bytes.Buffer
	withWorkingDir(t, root, func() {
		if err := run([]string{"v0.3.0"}, &stdout); err != nil {
			t.Fatalf("run validator: %v", err)
		}
	})
	if !strings.Contains(stdout.String(), "validated release docs for v0.3.0") {
		t.Fatalf("stdout = %q, want validated message", stdout.String())
	}
}

func TestValidateReleaseDocsRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tag       string
		notes     *string
		changelog string
		wantErr   string
	}{
		{
			name:      "invalid tag",
			tag:       "0.3.0",
			notes:     strPtr(validReleaseNotes("v0.3.0")),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "tag must match",
		},
		{
			name:      "missing release notes",
			tag:       "v0.3.0",
			notes:     nil,
			changelog: changelogFor("v0.3.0"),
			wantErr:   "docs/release-notes/v0.3.0.md not found",
		},
		{
			name:      "wrong title",
			tag:       "v0.3.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.3.0"), "# OpenHealth v0.3.0", "# OpenHealth 0.3.0", 1)),
			changelog: changelogFor("v0.3.0"),
			wantErr:   `must start with "# OpenHealth v0.3.0"`,
		},
		{
			name:      "missing changed section",
			tag:       "v0.3.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.3.0"), "## Changed", "## Updates", 1)),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "must include ## Changed",
		},
		{
			name:      "missing verification section",
			tag:       "v0.3.0",
			notes:     strPtr(strings.Replace(validReleaseNotes("v0.3.0"), "## Verification", "## Tests", 1)),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "must include ## Verification",
		},
		{
			name:      "missing changelog link",
			tag:       "v0.3.0",
			notes:     strPtr(validReleaseNotes("v0.3.0")),
			changelog: "# Changelog\n\nNo matching release link.\n",
			wantErr:   "CHANGELOG.md must link",
		},
		{
			name: "hard wrapped prose",
			tag:  "v0.3.0",
			notes: strPtr(`# OpenHealth v0.3.0

This paragraph was manually wrapped before the end of the prose sentence
and should be rejected by the release-doc validator.

## Changed

- Added a thing.

## Verification

- Checked a thing.
`),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "appears to hard-wrap prose",
		},
		{
			name: "hard wrapped indented bullet",
			tag:  "v0.3.0",
			notes: strPtr(`# OpenHealth v0.3.0

## Changed

- Added release notes validation that should reject manually wrapped
  list item continuation text.

## Verification

- Checked a thing.
`),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "appears to hard-wrap list item",
		},
		{
			name: "hard wrapped flush left bullet",
			tag:  "v0.3.0",
			notes: strPtr(`# OpenHealth v0.3.0

## Changed

- Added release notes validation that should reject manually wrapped
list item continuation text.

## Verification

- Checked a thing.
`),
			changelog: changelogFor("v0.3.0"),
			wantErr:   "appears to hard-wrap list item",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := writeReleaseRepo(t, tt.tag, derefString(tt.notes), tt.changelog)
			if tt.notes == nil {
				notesPath := filepath.Join(root, "docs", "release-notes", tt.tag+".md")
				if err := os.Remove(notesPath); err != nil && !os.IsNotExist(err) {
					t.Fatalf("remove notes: %v", err)
				}
			}
			err := validateReleaseDocs(root, tt.tag)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateReleaseDocs error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateReleaseNotesAllowsMarkdownStructure(t *testing.T) {
	t.Parallel()

	notes := `# OpenHealth v0.3.0

Short standalone prose is okay.

## Changed

- Bullets may wrap naturally in rendered Markdown without counting as prose.
- Separate bullet details are allowed when they stay on one source line.
1. Ordered list items are allowed.

` + "```" + `
Code fences can have consecutive plain-looking lines.
They are not release prose.
` + "```" + `

## Verification

- Verification bullets are okay.
`
	if err := validateReleaseNotes("docs/release-notes/v0.3.0.md", notes, "v0.3.0"); err != nil {
		t.Fatalf("validateReleaseNotes: %v", err)
	}
}

func validReleaseNotes(tag string) string {
	return "# OpenHealth " + tag + `

This release adds notes coverage and keeps release prose on one source line so GitHub Releases can wrap it naturally.

## Changed

- Added ordered result notes for labs and imaging.
- Added top-level notes for weight and blood pressure.

## Verification

- Production eval passed all 50 scenarios with stop-loss false.
`
}

func changelogFor(tag string) string {
	return "# Changelog\n\n- [" + tag + "](https://github.com/yazanabuashour/openhealth/releases/tag/" + tag + ") adds release docs validation.\n"
}

func writeReleaseRepo(t *testing.T, tag string, notes string, changelog string) string {
	t.Helper()

	root := t.TempDir()
	if notes != "" {
		notesPath := filepath.Join(root, "docs", "release-notes", tag+".md")
		if err := os.MkdirAll(filepath.Dir(notesPath), 0o755); err != nil {
			t.Fatalf("mkdir release notes dir: %v", err)
		}
		if err := os.WriteFile(notesPath, []byte(notes), 0o644); err != nil {
			t.Fatalf("write release notes: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "CHANGELOG.md"), []byte(changelog), 0o644); err != nil {
		t.Fatalf("write changelog: %v", err)
	}
	return root
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore dir %s: %v", oldDir, err)
		}
	}()
	fn()
}

func strPtr(value string) *string {
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
