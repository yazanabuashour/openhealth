package runner

import "strings"

func normalizeNoteList(values []string, field string) ([]string, string) {
	if len(values) == 0 {
		return nil, ""
	}
	notes := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, field + " must not contain empty values"
		}
		notes = append(notes, trimmed)
	}
	return notes, ""
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
