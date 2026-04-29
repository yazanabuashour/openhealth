package health

import (
	"context"
	"strings"
	"time"
)

func (s *service) ResolveBodyCompositionTarget(ctx context.Context, target BodyCompositionTarget) (BodyCompositionEntry, string, error) {
	items, err := s.ListBodyComposition(ctx, HistoryFilter{})
	if err != nil {
		return BodyCompositionEntry{}, "", err
	}
	matches := make([]BodyCompositionEntry, 0, 1)
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if sameDate(item.RecordedAt, target.RecordedAt) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return BodyCompositionEntry{}, "no matching body composition entry", nil
	case 1:
		return matches[0], "", nil
	default:
		return BodyCompositionEntry{}, "multiple matching body composition entries; target is ambiguous", nil
	}
}

func (s *service) ResolveSleepTarget(ctx context.Context, target SleepTarget) (SleepEntry, string, error) {
	items, err := s.ListSleep(ctx, HistoryFilter{})
	if err != nil {
		return SleepEntry{}, "", err
	}
	matches := make([]SleepEntry, 0, 1)
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if sameDate(item.RecordedAt, target.RecordedAt) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return SleepEntry{}, "no matching sleep entry", nil
	case 1:
		return matches[0], "", nil
	default:
		return SleepEntry{}, "multiple matching sleep entries; target is ambiguous", nil
	}
}

func (s *service) ResolveImagingTarget(ctx context.Context, target ImagingTarget) (ImagingRecord, string, error) {
	items, err := s.ListImaging(ctx, ImagingListParams{})
	if err != nil {
		return ImagingRecord{}, "", err
	}
	matches := make([]ImagingRecord, 0, 1)
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if sameDate(item.PerformedAt, target.PerformedAt) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return ImagingRecord{}, "no matching imaging record", nil
	case 1:
		return matches[0], "", nil
	default:
		return ImagingRecord{}, "multiple matching imaging records; target is ambiguous", nil
	}
}

func (s *service) ResolveMedicationTarget(ctx context.Context, target MedicationTarget) (MedicationCourse, string, error) {
	items, err := s.ListMedications(ctx, MedicationListParams{Status: MedicationStatusAll})
	if err != nil {
		return MedicationCourse{}, "", err
	}
	matches := make([]MedicationCourse, 0, 1)
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if strings.EqualFold(item.Name, target.Name) && item.StartDate == target.StartDate {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return MedicationCourse{}, "no matching medication", nil
	case 1:
		return matches[0], "", nil
	default:
		return MedicationCourse{}, "multiple matching medications; target is ambiguous", nil
	}
}

func (s *service) ResolveLabCollectionTarget(ctx context.Context, target LabCollectionTarget) (LabCollection, string, error) {
	items, err := s.ListLabCollections(ctx)
	if err != nil {
		return LabCollection{}, "", err
	}
	matches := make([]LabCollection, 0, 1)
	for _, item := range items {
		if target.ID > 0 {
			if item.ID == target.ID {
				matches = append(matches, item)
			}
			continue
		}
		if sameDate(item.CollectedAt, target.CollectedAt) {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return LabCollection{}, "no matching lab collection", nil
	case 1:
		return matches[0], "", nil
	default:
		return LabCollection{}, "multiple matching lab collections; target is ambiguous", nil
	}
}

func sameDate(left time.Time, right *time.Time) bool {
	return right != nil && left.Format(time.DateOnly) == right.Format(time.DateOnly)
}
