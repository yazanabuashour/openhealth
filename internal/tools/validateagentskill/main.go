package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	skillNamePattern           = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	markdownLinkPattern        = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	retiredHumanCLIPattern     = regexp.MustCompile(`\bopenhealth\s+(weight|blood-pressure)\s+(add|list|correct)\b`)
	retiredAgentOpsNamePattern = regexp.MustCompile(`\bopenhealth-agentops\b`)
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: scripts/validate-agent-skill.sh <skill-directory>")
	}
	skillDir := strings.TrimRight(args[0], string(os.PathSeparator))
	if skillDir == "" {
		skillDir = "."
	}
	if err := validateSkillDir(skillDir); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "validated %s\n", skillDir)
	return err
}

func validateSkillDir(skillDir string) error {
	info, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill directory not found: %s", skillDir)
		}
		return fmt.Errorf("stat skill directory %s: %w", skillDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill directory not found: %s", skillDir)
	}

	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return fmt.Errorf("read skill directory %s: %w", skillDir, err)
	}
	if len(entries) != 1 || entries[0].Name() != "SKILL.md" || entries[0].IsDir() {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		return fmt.Errorf("%s must contain only SKILL.md; found %s", skillDir, strings.Join(names, ", "))
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", skillFile, err)
	}
	frontmatter, err := extractFrontmatter(skillFile, string(content))
	if err != nil {
		return err
	}
	metadata, err := parseFrontmatter(skillFile, frontmatter)
	if err != nil {
		return err
	}
	if err := validateMetadata(skillDir, skillFile, metadata); err != nil {
		return err
	}
	if err := validateMarkdownLinks(skillDir, skillFile, string(content)); err != nil {
		return err
	}
	if err := validateRetiredGuidance(skillFile, string(content)); err != nil {
		return err
	}
	return nil
}

func extractFrontmatter(skillFile string, content string) (string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSuffix(lines[0], "\r") != "---" {
		return "", fmt.Errorf("%s must start with YAML frontmatter delimited by ---", skillFile)
	}

	closingLine := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSuffix(lines[i], "\r") == "---" {
			closingLine = i
			break
		}
	}
	if closingLine == -1 {
		return "", fmt.Errorf("%s must include a closing --- line for YAML frontmatter", skillFile)
	}
	if closingLine == 1 {
		return "", fmt.Errorf("%s frontmatter must contain at least the required fields", skillFile)
	}
	return strings.Join(lines[1:closingLine], "\n"), nil
}

func parseFrontmatter(skillFile string, frontmatter string) (map[string]string, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(frontmatter))
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("%s frontmatter must be valid YAML: %w", skillFile, err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("%s frontmatter must contain a single YAML document", skillFile)
		}
		return nil, fmt.Errorf("%s frontmatter must be valid YAML: %w", skillFile, err)
	}

	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s frontmatter must be a YAML mapping", skillFile)
	}

	out := map[string]string{}
	mapping := doc.Content[0]
	for i := 0; i < len(mapping.Content); i += 2 {
		keyNode := mapping.Content[i]
		valueNode := mapping.Content[i+1]
		if keyNode.Kind != yaml.ScalarNode || keyNode.Tag != "!!str" {
			return nil, fmt.Errorf("%s frontmatter keys must be strings", skillFile)
		}
		key := keyNode.Value
		if _, exists := out[key]; exists {
			return nil, fmt.Errorf("%s frontmatter field %q must not be duplicated", skillFile, key)
		}
		switch key {
		case "name", "description", "compatibility":
			if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!str" {
				return nil, fmt.Errorf("%s frontmatter field %q must be a string", skillFile, key)
			}
			out[key] = strings.TrimSpace(valueNode.Value)
		default:
			continue
		}
	}
	return out, nil
}

func validateMetadata(skillDir string, skillFile string, metadata map[string]string) error {
	name := metadata["name"]
	if name == "" {
		return fmt.Errorf("%s frontmatter must define a non-empty name", skillFile)
	}
	parentDir := filepath.Base(skillDir)
	if name != parentDir {
		return fmt.Errorf("%s name must match the parent directory (%q)", skillFile, parentDir)
	}
	if len([]rune(name)) > 64 {
		return fmt.Errorf("%s name must be 64 characters or fewer", skillFile)
	}
	if !skillNamePattern.MatchString(name) {
		return fmt.Errorf("%s name must use lowercase letters, numbers, and single hyphens only", skillFile)
	}

	description := metadata["description"]
	if description == "" {
		return fmt.Errorf("%s frontmatter must define a non-empty description", skillFile)
	}
	if len([]rune(description)) > 1024 {
		return fmt.Errorf("%s description must be 1024 characters or fewer", skillFile)
	}

	if compatibility, ok := metadata["compatibility"]; ok {
		if compatibility == "" {
			return fmt.Errorf("%s compatibility must be non-empty when provided", skillFile)
		}
		if len([]rune(compatibility)) > 500 {
			return fmt.Errorf("%s compatibility must be 500 characters or fewer", skillFile)
		}
	}
	return nil
}

func validateMarkdownLinks(skillDir string, skillFile string, content string) error {
	for _, match := range markdownLinkPattern.FindAllStringSubmatch(content, -1) {
		target := match[1]
		if shouldSkipLinkTarget(target) {
			continue
		}
		targetPath := filepath.Clean(filepath.Join(skillDir, target))
		if _, err := os.Stat(targetPath); err != nil {
			return fmt.Errorf("%s link target %q is not installed with the skill: %w", skillFile, target, err)
		}
	}
	return nil
}

func shouldSkipLinkTarget(target string) bool {
	if target == "" || strings.HasPrefix(target, "#") || filepath.IsAbs(target) {
		return true
	}
	if parsed, err := url.Parse(target); err == nil && parsed.Scheme != "" {
		return true
	}
	return false
}

func validateRetiredGuidance(skillFile string, content string) error {
	forbiddenSubstrings := []string{
		"go run ./cmd/openhealth",
		"cmd/openhealth-agentops",
		"CLI fallback",
		"Generated Client Fallback",
	}
	for _, forbidden := range forbiddenSubstrings {
		if strings.Contains(content, forbidden) {
			return fmt.Errorf("%s contains retired product guidance %q", skillFile, forbidden)
		}
	}
	if retiredAgentOpsNamePattern.MatchString(content) {
		return fmt.Errorf("%s contains retired product binary name openhealth-agentops", skillFile)
	}
	if retiredHumanCLIPattern.MatchString(content) {
		return fmt.Errorf("%s contains retired human CLI command guidance", skillFile)
	}
	return nil
}
