package health

import (
	"math"
	"slices"
	"time"
)

func sortWeightEntriesDescending(entries []WeightEntry) []WeightEntry {
	sorted := slices.Clone(entries)
	slices.SortFunc(sorted, compareWeightEntryDescending)
	return sorted
}

func sortWeightEntriesAscending(entries []WeightEntry) []WeightEntry {
	sorted := slices.Clone(entries)
	slices.SortFunc(sorted, compareWeightEntryAscending)
	return sorted
}

func filteredWeightEntries(entries []WeightEntry, start time.Time) []WeightEntry {
	filtered := make([]WeightEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.RecordedAt.Before(start) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func rangeStart(rangeValue WeightRange, latest *WeightEntry) *time.Time {
	if latest == nil || rangeValue == WeightRangeAll {
		return nil
	}

	start := latest.RecordedAt.UTC()
	switch rangeValue {
	case WeightRange30d:
		start = start.AddDate(0, 0, -30)
	case WeightRange90d:
		start = start.AddDate(0, 0, -90)
	case WeightRange1y:
		start = start.AddDate(-1, 0, 0)
	}
	return &start
}

func movingAveragePoints(entries []WeightEntry) []MovingAveragePoint {
	points := make([]MovingAveragePoint, 0, len(entries))
	for _, entry := range entries {
		windowStart := entry.RecordedAt.Add(-6 * 24 * time.Hour)
		total := 0.0
		count := 0
		for _, candidate := range entries {
			if !candidate.RecordedAt.Before(windowStart) && !candidate.RecordedAt.After(entry.RecordedAt) {
				total += candidate.Value
				count++
			}
		}
		if count == 0 {
			continue
		}
		points = append(points, MovingAveragePoint{
			RecordedAt: entry.RecordedAt,
			Value:      round2(total / float64(count)),
		})
	}
	return points
}

func monthlyAverageBuckets(entries []WeightEntry) []MonthlyAverageBucket {
	groups := map[string][]float64{}
	orderedMonths := []string{}
	for _, entry := range entries {
		month := entry.RecordedAt.UTC().Format("2006-01")
		if _, ok := groups[month]; !ok {
			orderedMonths = append(orderedMonths, month)
		}
		groups[month] = append(groups[month], entry.Value)
	}

	buckets := make([]MonthlyAverageBucket, 0, len(orderedMonths))
	for _, month := range orderedMonths {
		values := groups[month]
		total := 0.0
		for _, value := range values {
			total += value
		}
		buckets = append(buckets, MonthlyAverageBucket{
			Month: month,
			Value: round2(total / float64(len(values))),
		})
	}
	return buckets
}

func compareWeightEntryDescending(left, right WeightEntry) int {
	if cmp := compareTimeDescending(left.RecordedAt, right.RecordedAt); cmp != 0 {
		return cmp
	}
	return compareIntDescending(left.ID, right.ID)
}

func compareWeightEntryAscending(left, right WeightEntry) int {
	if cmp := compareTimeAscending(left.RecordedAt, right.RecordedAt); cmp != 0 {
		return cmp
	}
	return compareIntAscending(left.ID, right.ID)
}

func compareBloodPressureEntryAscending(left, right BloodPressureEntry) int {
	if cmp := compareTimeAscending(left.RecordedAt, right.RecordedAt); cmp != 0 {
		return cmp
	}
	return compareIntAscending(left.ID, right.ID)
}

func compareLabResultWithCollectionDescending(left, right LabResultWithCollection) int {
	if cmp := compareTimeDescending(left.CollectedAt, right.CollectedAt); cmp != 0 {
		return cmp
	}
	return compareIntDescending(left.ID, right.ID)
}

func compareLabResultWithCollectionAscending(left, right LabResultWithCollection) int {
	if cmp := compareTimeAscending(left.CollectedAt, right.CollectedAt); cmp != 0 {
		return cmp
	}
	return compareIntAscending(left.ID, right.ID)
}

func compareTimeDescending(left, right time.Time) int {
	if left.After(right) {
		return -1
	}
	if left.Before(right) {
		return 1
	}
	return 0
}

func compareTimeAscending(left, right time.Time) int {
	if left.Before(right) {
		return -1
	}
	if left.After(right) {
		return 1
	}
	return 0
}

func compareIntDescending(left, right int) int {
	switch {
	case left > right:
		return -1
	case left < right:
		return 1
	default:
		return 0
	}
}

func compareIntAscending(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func firstWeightEntry(entries []WeightEntry) *WeightEntry {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[0]
	return &entry
}

func firstWeightEntryFromEnd(entries []WeightEntry) *WeightEntry {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[len(entries)-1]
	return &entry
}

func intPointer(value int) *int {
	return &value
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
