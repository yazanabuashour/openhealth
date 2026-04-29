package main

import (
	"path/filepath"
	"strings"
	"unicode"
)

func isFileInspectionCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "sed", "cat", "find", "ls", "awk", "head", "tail", "nl")
}

func inspectsGeneratedFileCommand(command string, output string) bool {
	if !isFileInspectionCommand(command) || isBroadRepoSearchCommand(command) {
		return false
	}
	if onlyAgentSkillTargets(commandPathTargets(commandFields(command))) {
		return false
	}
	if mentionsGeneratedPath(command) {
		return true
	}
	return isContentSearchCommand(command) && outputHasGeneratedResultPath(output)
}

func isBroadRepoSearchCommand(command string) bool {
	fields := commandFields(command)
	normalized := " " + strings.Join(fields, " ") + " "
	switch {
	case commandHasExecutable(command, "rg"):
		return rgBroadRepoSearch(fields, normalized)
	case commandHasExecutable(command, "grep"):
		return grepBroadRepoSearch(fields)
	case commandHasExecutable(command, "find"):
		return hasRepoRootTarget(fields)
	default:
		return false
	}
}

func rgBroadRepoSearch(fields []string, normalized string) bool {
	args := commandArgsAfterExecutable(fields, "rg")
	filesMode := strings.Contains(normalized, " --files ")
	targets := rgPathTargets(args, filesMode)
	return len(targets) == 0 || hasRepoRootTarget(targets) || onlyRepoRootTargets(targets)
}

func rgPathTargets(args []string, filesMode bool) []string {
	targets := []string{}
	patternSeen := filesMode
	for i := 0; i < len(args); i++ {
		field := args[i]
		if field == "" {
			continue
		}
		if field == "--" {
			for _, rest := range args[i+1:] {
				if filesMode || patternSeen {
					targets = append(targets, rest)
					continue
				}
				patternSeen = true
			}
			break
		}
		if strings.HasPrefix(field, "--") {
			if longOptionTakesValue(field) && !strings.Contains(field, "=") {
				i++
			}
			continue
		}
		if strings.HasPrefix(field, "-") && field != "-" {
			if shortOptionTakesValue(field) && len(field) == 2 {
				i++
			}
			continue
		}
		if filesMode || patternSeen {
			targets = append(targets, field)
			continue
		}
		patternSeen = true
	}
	return targets
}

func longOptionTakesValue(field string) bool {
	switch field {
	case "--glob", "--iglob", "--type", "--type-not", "--type-add", "--regexp", "--file", "--max-count", "--after-context", "--before-context", "--context", "--colors", "--engine", "--sort", "--sortr", "--path-separator":
		return true
	default:
		return false
	}
}

func shortOptionTakesValue(field string) bool {
	switch field {
	case "-g", "-e", "-f", "-t", "-T", "-m", "-A", "-B", "-C", "-E", "-M":
		return true
	default:
		return false
	}
}

func commandArgsAfterExecutable(fields []string, names ...string) []string {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for i, field := range fields {
		if _, ok := nameSet[field]; ok {
			return fields[i+1:]
		}
		if _, ok := nameSet[filepath.Base(field)]; ok {
			return fields[i+1:]
		}
	}
	return fields
}

func grepBroadRepoSearch(fields []string) bool {
	if !hasRepoRootTarget(fields) {
		return false
	}
	for _, field := range fields {
		if field == "-r" || field == "-R" || strings.Contains(field, "r") && strings.HasPrefix(field, "-") {
			return true
		}
	}
	return false
}

func commandPathTargets(fields []string) []string {
	targets := []string{}
	for _, field := range fields {
		if field == "" || strings.HasPrefix(field, "-") {
			continue
		}
		switch filepath.Base(field) {
		case "rg", "grep", "find", "awk", "zsh", "bash", "sh":
			continue
		}
		switch field {
		case "lc":
			continue
		}
		if field == "." || field == "./" || strings.HasPrefix(field, "./") || strings.HasPrefix(field, "../") || strings.Contains(field, "/") {
			targets = append(targets, field)
		}
	}
	return targets
}

func onlyRepoRootTargets(targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		if target != "." && target != "./" && target != "./." {
			return false
		}
	}
	return true
}

func onlyAgentSkillTargets(targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	for _, target := range targets {
		cleaned := filepath.ToSlash(strings.Trim(target, `"'`))
		if !strings.HasPrefix(cleaned, ".agents/skills/") &&
			!strings.HasPrefix(cleaned, "skills/openhealth/") {
			return false
		}
	}
	return true
}

func hasRepoRootTarget(fields []string) bool {
	for _, field := range fields {
		if field == "." || field == "./" || field == "./." {
			return true
		}
	}
	return false
}

func isContentSearchCommand(command string) bool {
	return commandHasExecutable(command, "rg", "grep", "awk")
}

func mentionsGeneratedPath(text string) bool {
	return strings.Contains(text, "internal/storage/sqlite/sqlc")
}

func outputHasGeneratedResultPath(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "./")
		if strings.HasPrefix(trimmed, "internal/storage/sqlite/sqlc/") {
			return true
		}
	}
	return false
}

func inspectsModuleCache(command string) bool {
	lower := strings.ToLower(command)
	return strings.Contains(lower, "gomodcache") ||
		strings.Contains(lower, "pkg/mod")
}

func usesOpenHealthCLI(command string) bool {
	for _, segment := range shellCommandSegments(shellCommandPayload(command)) {
		if segmentUsesOpenHealthCLI(segment) {
			return true
		}
	}
	return false
}

func segmentUsesOpenHealthCLI(segment string) bool {
	fields := commandFields(segment)
	if primaryExecutableIs(fields, "rg", "grep", "sed", "awk", "find", "cat", "head", "tail", "nl", "echo", "printf") {
		return false
	}
	for i, field := range fields {
		if filepath.Base(field) == "go" && i+2 < len(fields) && fields[i+1] == "run" {
			for _, candidate := range fields[i+2:] {
				trimmed := strings.Trim(candidate, `"'`)
				if trimmed == "./cmd/openhealth" || trimmed == "cmd/openhealth" || strings.HasSuffix(trimmed, "/cmd/openhealth") {
					return true
				}
			}
		}
		if (field == "openhealth" || strings.HasSuffix(field, "/openhealth")) && i+1 < len(fields) {
			if fields[i+1] == "migrate" || fields[i+1] == "serve" {
				return true
			}
			if i+2 < len(fields) {
				switch fields[i+1] {
				case "weight":
					if fields[i+2] == "add" || fields[i+2] == "list" {
						return true
					}
				case "blood-pressure":
					if fields[i+2] == "add" || fields[i+2] == "list" || fields[i+2] == "correct" {
						return true
					}
				}
			}
		}
	}
	return false
}

func shellCommandPayload(command string) string {
	lower := strings.ToLower(command)
	for _, marker := range []string{" -lc ", " -c "} {
		index := strings.Index(lower, marker)
		if index < 0 {
			continue
		}
		return trimShellArgument(strings.TrimSpace(command[index+len(marker):]))
	}
	return command
}

func trimShellArgument(value string) string {
	if len(value) < 2 {
		return value
	}
	quote := value[0]
	if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
		return value[1 : len(value)-1]
	}
	return value
}

func shellCommandSegments(command string) []string {
	segments := []string{}
	start := 0
	var quote rune
	for i, r := range command {
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ';' || r == '|' || r == '&' || r == '\n' {
			if segment := strings.TrimSpace(command[start:i]); segment != "" {
				segments = append(segments, segment)
			}
			start = i + 1
		}
	}
	if segment := strings.TrimSpace(command[start:]); segment != "" {
		segments = append(segments, segment)
	}
	return segments
}

func primaryExecutableIs(fields []string, names ...string) bool {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for _, field := range fields {
		base := filepath.Base(field)
		if base == "zsh" || base == "bash" || base == "sh" || field == "lc" || field == "cd" || strings.Contains(field, "=") {
			continue
		}
		_, ok := nameSet[base]
		return ok
	}
	return false
}

func usesDirectSQLite(command string) bool {
	lower := strings.ToLower(command)
	return commandHasExecutable(command, "sqlite3") ||
		strings.Contains(lower, "import sqlite3") ||
		strings.Contains(lower, "modernc.org/sqlite") ||
		(strings.Contains(lower, "database/sql") && strings.Contains(lower, "sqlite"))
}

func addMetricEvidence(evidence *[]string, command string) {
	const maxEvidence = 5
	sanitized := sanitizeMetricEvidence(command)
	for _, existing := range *evidence {
		if existing == sanitized {
			return
		}
	}
	if len(*evidence) >= maxEvidence {
		return
	}
	*evidence = append(*evidence, sanitized)
}

func sanitizeMetricEvidence(command string) string {
	fields := strings.Fields(command)
	for i, field := range fields {
		if strings.Contains(field, "openhealth-oh") {
			fields[i] = "<run-root>"
		}
	}
	return strings.Join(fields, " ")
}

func commandHasExecutable(command string, names ...string) bool {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for _, field := range commandFields(command) {
		if _, ok := nameSet[field]; ok {
			return true
		}
		if _, ok := nameSet[filepath.Base(field)]; ok {
			return true
		}
	}
	return false
}

func commandFields(command string) []string {
	return strings.FieldsFunc(strings.ToLower(command), func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("'\"`;&|()", r)
	})
}
