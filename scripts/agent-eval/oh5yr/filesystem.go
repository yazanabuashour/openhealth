package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyRepo(src string, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if shouldSkipCopy(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func shouldSkipCopy(rel string, entry fs.DirEntry) bool {
	name := entry.Name()
	if shouldSkipEvalPath(filepath.ToSlash(rel)) {
		return true
	}
	if entry.IsDir() {
		switch name {
		case ".git", ".beads", ".dolt", ".agents":
			return true
		}
	}
	if strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".db-shm") || strings.HasSuffix(name, ".db-wal") {
		return true
	}
	return false
}

func shouldSkipEvalPath(rel string) bool {
	switch rel {
	case "AGENTS.md", "docs/agent-evals.md", "docs/agent-eval-assets", "docs/agent-eval-results", "scripts/agent-eval":
		return true
	}
	return strings.HasPrefix(rel, "docs/agent-eval-assets/") ||
		strings.HasPrefix(rel, "docs/agent-eval-results/") ||
		strings.HasPrefix(rel, "scripts/agent-eval/")
}

func copyFile(src string, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func installVariant(repoRoot string, runRepo string, currentVariant variant) error {
	dest := filepath.Join(runRepo, ".agents", "skills", "openhealth")
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	var src string
	switch currentVariant.ID {
	case "production":
		src = filepath.Join(repoRoot, "skills", "openhealth")
	default:
		return fmt.Errorf("unknown variant %q", currentVariant.ID)
	}
	if err := copyDir(src, dest); err != nil {
		return err
	}
	return nil
}

func preflightEvalContext(repoRoot string, runRepo string, runDir string, cache cacheConfig) error {
	sourceSkill := filepath.Join(repoRoot, "skills", "openhealth", "SKILL.md")
	installedSkill := filepath.Join(runRepo, ".agents", "skills", "openhealth", "SKILL.md")
	sourceBytes, err := os.ReadFile(sourceSkill)
	if err != nil {
		return err
	}
	installedBytes, err := os.ReadFile(installedSkill)
	if err != nil {
		return err
	}
	if !bytes.Equal(sourceBytes, installedBytes) {
		return errors.New("installed production skill does not match shipped SKILL.md")
	}
	if _, err := os.Stat(filepath.Join(runRepo, "AGENTS.md")); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("production eval repo must not contain AGENTS.md")
		}
		return err
	}

	cmd := exec.Command("codex", "debug", "prompt-input", "Use OpenHealth to list latest weight.")
	cmd.Dir = runRepo
	cmd.Env = evalEnv(runDir, evalDatabasePath(runRepo), cache)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	rendered := string(output)
	if !strings.Contains(rendered, "- openhealth:") {
		return errors.New("rendered prompt is missing openhealth skill discovery")
	}
	if !strings.Contains(rendered, ".agents/skills/openhealth/SKILL.md") {
		return errors.New("rendered prompt does not point openhealth to the installed project skill")
	}
	if containsOpenHealthAgentsInstructions(rendered) {
		return errors.New("rendered prompt contains OpenHealth product instructions from AGENTS.md")
	}
	return nil
}

func containsOpenHealthAgentsInstructions(rendered string) bool {
	const marker = "# AGENTS.md instructions"
	index := strings.Index(rendered, marker)
	if index < 0 {
		return false
	}
	agentsText := rendered[index:]
	for _, forbidden := range []string{
		"openhealth",
		"upsert_weights",
		"record_blood_pressure",
		"record_medications",
		"record_labs",
		"ambiguous short date",
		"product data agent",
	} {
		if strings.Contains(agentsText, forbidden) {
			return true
		}
	}
	return false
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func writeJSON(path string, value any) error {
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return os.WriteFile(path, data.Bytes(), 0o644)
}
