package runner

import (
	"time"

	client "github.com/yazanabuashour/openhealth/internal/runclient"
)

func normalizedLabCollectionFromClient(item client.LabCollection) normalizedLabCollectionInput {
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

func toClientLabCollectionInput(input normalizedLabCollectionInput) client.LabCollectionInput {
	panels := make([]client.LabPanelInput, 0, len(input.Panels))
	for _, panel := range input.Panels {
		results := make([]client.LabResultInput, 0, len(panel.Results))
		for _, result := range panel.Results {
			results = append(results, client.LabResultInput(result))
		}
		panels = append(panels, client.LabPanelInput{
			PanelName: panel.PanelName,
			Results:   results,
		})
	}
	return client.LabCollectionInput{
		CollectedAt: input.CollectedAt,
		Note:        input.Note,
		Panels:      panels,
	}
}

func labCollectionWrite(item client.LabCollection, status string) LabCollectionWrite {
	return LabCollectionWrite{
		ID:     item.ID,
		Date:   item.CollectedAt.Format(time.DateOnly),
		Status: status,
	}
}

func labCollectionEntry(item client.LabCollection, analyteSlug *client.AnalyteSlug) (LabCollectionEntry, bool) {
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

func labResultEntry(item client.LabResult) LabResultEntry {
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
