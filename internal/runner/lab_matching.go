package runner

import (
	"math"
	"strings"
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

func applyLabResultUpdate(collection *normalizedLabCollectionInput, update normalizedLabResultUpdateInput) string {
	panelIndex := -1
	for i, panel := range collection.Panels {
		if strings.EqualFold(strings.TrimSpace(panel.PanelName), update.PanelName) {
			if panelIndex != -1 {
				return "multiple matching lab panels; patch is ambiguous"
			}
			panelIndex = i
		}
	}
	if panelIndex == -1 {
		return "no matching lab panel"
	}

	resultIndex := -1
	for i, result := range collection.Panels[panelIndex].Results {
		if !labResultMatchesPatch(result, update) {
			continue
		}
		if resultIndex != -1 {
			return "multiple matching lab results; patch is ambiguous"
		}
		resultIndex = i
	}
	if resultIndex == -1 {
		return "no matching lab result"
	}

	collection.Panels[panelIndex].Results[resultIndex] = update.Result
	return ""
}

func labResultMatchesPatch(result normalizedLabResultInput, update normalizedLabResultUpdateInput) bool {
	if update.MatchCanonicalSlug != nil {
		return result.CanonicalSlug != nil && *result.CanonicalSlug == *update.MatchCanonicalSlug
	}
	return strings.EqualFold(strings.TrimSpace(result.TestName), update.MatchTestName)
}

func matchingLabCollection(items []client.LabCollection, input normalizedLabCollectionInput) (client.LabCollection, bool) {
	for _, item := range items {
		if labCollectionMatches(item, input) {
			return item, true
		}
	}
	return client.LabCollection{}, false
}

func labCollectionMatches(item client.LabCollection, input normalizedLabCollectionInput) bool {
	if item.CollectedAt.Format(time.DateOnly) != input.CollectedAt.Format(time.DateOnly) ||
		!equalStringPointer(item.Note, input.Note) ||
		len(item.Panels) != len(input.Panels) {
		return false
	}
	for i, panel := range item.Panels {
		if panel.PanelName != input.Panels[i].PanelName || len(panel.Results) != len(input.Panels[i].Results) {
			return false
		}
		for j, result := range panel.Results {
			if !labResultMatches(result, input.Panels[i].Results[j]) {
				return false
			}
		}
	}
	return true
}

func labResultMatches(item client.LabResult, input normalizedLabResultInput) bool {
	if item.TestName != input.TestName ||
		!equalAnalyteSlugPointer(item.CanonicalSlug, input.CanonicalSlug) ||
		item.ValueText != input.ValueText ||
		!equalFloatPointer(item.ValueNumeric, input.ValueNumeric) ||
		!equalStringPointer(item.Units, input.Units) ||
		!equalStringPointer(item.RangeText, input.RangeText) ||
		!equalStringPointer(item.Flag, input.Flag) ||
		!equalStringSlices(item.Notes, input.Notes) {
		return false
	}
	return true
}

func equalAnalyteSlugPointer(left *client.AnalyteSlug, right *client.AnalyteSlug) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func equalFloatPointer(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return math.Abs(*left-*right) < 0.000000001
}
