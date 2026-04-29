package health

import (
	"context"
	"time"
)

func (s *service) Summary(ctx context.Context) (Summary, error) {
	weights, err := s.repo.ListWeightEntries(ctx, HistoryFilter{})
	if err != nil {
		return Summary{}, err
	}
	orderedWeights := sortWeightEntriesDescending(weights)
	latestWeight := firstWeightEntry(orderedWeights)

	average7d := calculateAverage7d(orderedWeights, latestWeight)
	delta30d := calculateDelta30d(orderedWeights, latestWeight)

	bloodPressureEntries, err := s.repo.ListBloodPressureEntries(ctx, HistoryFilter{
		Limit: intPointer(1),
	})
	if err != nil {
		return Summary{}, err
	}
	latestBloodPressure := firstBloodPressureEntry(bloodPressureEntries)

	sleepEntries, err := s.repo.ListSleepEntries(ctx, HistoryFilter{
		Limit: intPointer(1),
	})
	if err != nil {
		return Summary{}, err
	}
	latestSleep := firstSleepEntry(sleepEntries)

	today := s.now().UTC().Format(time.DateOnly)
	activeMedicationCount, err := s.repo.CountActiveMedicationCourses(ctx, today)
	if err != nil {
		return Summary{}, err
	}

	labCollections, err := s.repo.ListLabCollections(ctx)
	if err != nil {
		return Summary{}, err
	}

	latestLabHighlights := []LabResultWithCollection{}
	if len(labCollections) > 0 {
		latestLabHighlights = latestCollectionHighlights(labCollections[0])
	}

	return Summary{
		LatestWeight:          latestWeight,
		Average7d:             average7d,
		Delta30d:              delta30d,
		LatestBloodPressure:   latestBloodPressure,
		LatestSleep:           latestSleep,
		ActiveMedicationCount: activeMedicationCount,
		LatestLabHighlights:   latestLabHighlights,
	}, nil
}

func latestCollectionHighlights(collection LabCollection) []LabResultWithCollection {
	highlights := make([]LabResultWithCollection, 0)
	for _, panel := range collection.Panels {
		for _, result := range panel.Results {
			if !includeLabHighlight(result) {
				continue
			}
			highlight := LabResultWithCollection{
				LabResult:    result,
				CollectedAt:  collection.CollectedAt,
				CollectionID: collection.ID,
				PanelName:    panel.PanelName,
			}
			highlights = append(highlights, highlight)
			if len(highlights) == 6 {
				return highlights
			}
		}
	}
	return highlights
}

func includeLabHighlight(result LabResult) bool {
	if result.Flag != nil && *result.Flag != "" {
		return true
	}
	return result.CanonicalSlug != nil
}

func calculateAverage7d(entries []WeightEntry, latest *WeightEntry) *float64 {
	if latest == nil {
		return nil
	}
	windowStart := latest.RecordedAt.Add(-6 * 24 * time.Hour)
	total := 0.0
	count := 0
	for _, entry := range entries {
		if !entry.RecordedAt.Before(windowStart) && !entry.RecordedAt.After(latest.RecordedAt) {
			total += entry.Value
			count++
		}
	}
	if count == 0 {
		return nil
	}
	value := round2(total / float64(count))
	return &value
}

func calculateDelta30d(entries []WeightEntry, latest *WeightEntry) *float64 {
	if latest == nil {
		return nil
	}
	cutoff := latest.RecordedAt.AddDate(0, 0, -30)
	var baseline *WeightEntry
	for i := range entries {
		if !entries[i].RecordedAt.After(cutoff) {
			baseline = &entries[i]
			break
		}
	}
	if baseline == nil && len(entries) > 0 {
		baseline = &entries[len(entries)-1]
	}
	if baseline == nil {
		return nil
	}
	value := round2(latest.Value - baseline.Value)
	return &value
}

func firstBloodPressureEntry(entries []BloodPressureEntry) *BloodPressureEntry {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[0]
	return &entry
}

func firstSleepEntry(entries []SleepEntry) *SleepEntry {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[0]
	return &entry
}
