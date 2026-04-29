package health

import (
	"context"
	"slices"
)

func (s *service) ListBloodPressure(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.ListBloodPressureEntries(ctx, filter)
}

func (s *service) RecordBloodPressure(ctx context.Context, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	normalized, err := normalizeBloodPressureRecordInput(input)
	if err != nil {
		return BloodPressureEntry{}, err
	}
	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return BloodPressureEntry{}, &DatabaseError{
			Message: "failed to generate blood pressure entry hash",
			Cause:   err,
		}
	}
	now := s.now().UTC()
	return s.repo.CreateBloodPressureEntry(ctx, CreateBloodPressureEntryParams{
		RecordedAt:       normalized.RecordedAt,
		Systolic:         normalized.Systolic,
		Diastolic:        normalized.Diastolic,
		Pulse:            normalized.Pulse,
		Note:             normalized.Note,
		Source:           manualSource,
		SourceRecordHash: sourceRecordHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func (s *service) ReplaceBloodPressure(ctx context.Context, id int, input BloodPressureRecordInput) (BloodPressureEntry, error) {
	if err := validateRecordID(id); err != nil {
		return BloodPressureEntry{}, err
	}
	normalized, err := normalizeBloodPressureRecordInput(input)
	if err != nil {
		return BloodPressureEntry{}, err
	}
	return s.repo.UpdateBloodPressureEntry(ctx, UpdateBloodPressureEntryParams{
		ID:         id,
		RecordedAt: normalized.RecordedAt,
		Systolic:   normalized.Systolic,
		Diastolic:  normalized.Diastolic,
		Pulse:      normalized.Pulse,
		Note:       normalized.Note,
		UpdatedAt:  s.now().UTC(),
	})
}

func (s *service) DeleteBloodPressure(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteBloodPressureEntry(ctx, DeleteBloodPressureEntryParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}

func (s *service) BloodPressureTrend(ctx context.Context, filter HistoryFilter) ([]BloodPressureEntry, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}
	entries, err := s.repo.ListBloodPressureEntries(ctx, filter)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(entries, compareBloodPressureEntryAscending)
	return entries, nil
}
