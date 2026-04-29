package runner

import (
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

func normalizedLabCollectionFromClient(item health.LabCollection) normalizedLabCollectionInput {
	panels := make([]normalizedLabPanelInput, 0, len(item.Panels))
	for _, panel := range item.Panels {
		results := make([]normalizedLabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, normalizedLabResultInput{
				TestName:      result.TestName,
				CanonicalSlug: result.CanonicalSlug,
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				RangeText:     result.RangeText,
				Flag:          result.Flag,
				Notes:         append([]string(nil), result.Notes...),
			})
		}
		panels = append(panels, normalizedLabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return normalizedLabCollectionInput{
		CollectedAt: item.CollectedAt,
		Note:        item.Note,
		Panels:      panels,
	}
}

func toHealthLabCollectionInput(input normalizedLabCollectionInput) health.LabCollectionInput {
	panels := make([]health.LabPanelInput, 0, len(input.Panels))
	for _, panel := range input.Panels {
		results := make([]health.LabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, health.LabResultInput(result))
		}
		panels = append(panels, health.LabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return health.LabCollectionInput{
		CollectedAt: input.CollectedAt,
		Note:        input.Note,
		Panels:      panels,
	}
}

func labCollectionWrite(item health.LabCollection, status string) LabCollectionWrite {
	return LabCollectionWrite{
		ID:     item.ID,
		Date:   item.CollectedAt.Format(time.DateOnly),
		Status: status,
	}
}

func labCollectionEntry(item health.LabCollection, analyteSlug *health.AnalyteSlug) (LabCollectionEntry, bool) {
	entry := LabCollectionEntry{
		ID:   item.ID,
		Date: item.CollectedAt.Format(time.DateOnly),
		Note: item.Note,
	}
	for _, panel := range item.Panels {
		panelEntry := LabPanelEntry{PanelName: panel.PanelName}
		for _, result := range panel.Results {
			if analyteSlug != nil && (result.CanonicalSlug == nil || *result.CanonicalSlug != *analyteSlug) {
				continue
			}
			panelEntry.Results = append(panelEntry.Results, labResultEntry(result))
		}
		if len(panelEntry.Results) > 0 {
			entry.Panels = append(entry.Panels, panelEntry)
		}
	}
	if len(entry.Panels) == 0 {
		return LabCollectionEntry{}, false
	}
	return entry, true
}

func labResultEntry(item health.LabResult) LabResultEntry {
	var slug *string
	if item.CanonicalSlug != nil {
		value := string(*item.CanonicalSlug)
		slug = &value
	}
	return LabResultEntry{
		TestName:      item.TestName,
		CanonicalSlug: slug,
		ValueText:     item.ValueText,
		ValueNumeric:  item.ValueNumeric,
		Units:         item.Units,
		RangeText:     item.RangeText,
		Flag:          item.Flag,
		Notes:         append([]string(nil), item.Notes...),
	}
}
