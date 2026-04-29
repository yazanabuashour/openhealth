package health

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *service) ListWeight(ctx context.Context, filter HistoryFilter) ([]WeightEntry, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.ListWeightEntries(ctx, filter)
}

func (s *service) RecordWeight(ctx context.Context, input WeightRecordInput) (WeightEntry, error) {
	normalized, err := normalizeWeightRecordInput(input)
	if err != nil {
		return WeightEntry{}, err
	}

	existing, err := s.repo.FindManualWeightEntry(ctx, FindManualWeightEntryParams{
		RecordedAt: normalized.RecordedAt,
		Unit:       normalized.Unit,
	})
	if err != nil {
		return WeightEntry{}, err
	}
	if existing != nil {
		return WeightEntry{}, &ConflictError{
			Message: fmt.Sprintf("manual weight already exists on %s %s", normalized.RecordedAt.Format(time.DateOnly), normalized.Unit),
		}
	}

	return s.createManualWeight(ctx, normalized)
}

func (s *service) UpsertWeight(ctx context.Context, input WeightRecordInput) (WeightWriteResult, error) {
	normalized, err := normalizeWeightRecordInput(input)
	if err != nil {
		return WeightWriteResult{}, err
	}

	existing, err := s.repo.FindManualWeightEntry(ctx, FindManualWeightEntryParams{
		RecordedAt: normalized.RecordedAt,
		Unit:       normalized.Unit,
	})
	if err != nil {
		return WeightWriteResult{}, err
	}
	if existing == nil {
		entry, err := s.createManualWeight(ctx, normalized)
		if err != nil {
			var conflictErr *ConflictError
			if errors.As(err, &conflictErr) {
				return s.upsertExistingWeight(ctx, normalized)
			}
			return WeightWriteResult{}, err
		}
		return WeightWriteResult{
			Entry:  entry,
			Status: WeightWriteStatusCreated,
		}, nil
	}
	return upsertWeightEntryValue(s, ctx, *existing, normalized)
}

func (s *service) upsertExistingWeight(ctx context.Context, input WeightRecordInput) (WeightWriteResult, error) {
	existing, err := s.repo.FindManualWeightEntry(ctx, FindManualWeightEntryParams{
		RecordedAt: input.RecordedAt,
		Unit:       input.Unit,
	})
	if err != nil {
		return WeightWriteResult{}, err
	}
	if existing == nil {
		return WeightWriteResult{}, &ConflictError{
			Message: fmt.Sprintf("manual weight already exists on %s %s", input.RecordedAt.Format(time.DateOnly), input.Unit),
		}
	}
	return upsertWeightEntryValue(s, ctx, *existing, input)
}

func upsertWeightEntryValue(s *service, ctx context.Context, existing WeightEntry, input WeightRecordInput) (WeightWriteResult, error) {
	noteChanged := input.Note != nil && !equalStringPointer(existing.Note, input.Note)
	if equalWeightValue(existing.Value, input.Value) && !noteChanged {
		return WeightWriteResult{
			Entry:  existing,
			Status: WeightWriteStatusAlreadyExists,
		}, nil
	}

	update := WeightUpdateInput{}
	if !equalWeightValue(existing.Value, input.Value) {
		value := input.Value
		update.Value = &value
	}
	if noteChanged {
		update.Note = input.Note
	}
	entry, err := s.UpdateWeight(ctx, existing.ID, update)
	if err != nil {
		return WeightWriteResult{}, err
	}
	return WeightWriteResult{
		Entry:  entry,
		Status: WeightWriteStatusUpdated,
	}, nil
}

func (s *service) createManualWeight(ctx context.Context, input WeightRecordInput) (WeightEntry, error) {
	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return WeightEntry{}, &DatabaseError{
			Message: "failed to generate weight entry hash",
			Cause:   err,
		}
	}

	now := s.now().UTC()
	return s.repo.CreateWeightEntry(ctx, CreateWeightEntryParams{
		RecordedAt:       input.RecordedAt,
		Value:            input.Value,
		Unit:             input.Unit,
		Source:           manualWeightSource,
		SourceRecordHash: sourceRecordHash,
		Note:             input.Note,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func (s *service) UpdateWeight(ctx context.Context, id int, input WeightUpdateInput) (WeightEntry, error) {
	if err := validateWeightID(id); err != nil {
		return WeightEntry{}, err
	}
	if err := validateWeightUpdateInput(input); err != nil {
		return WeightEntry{}, err
	}

	params := UpdateWeightEntryParams{
		ID:        id,
		UpdatedAt: s.now().UTC(),
	}
	if input.RecordedAt != nil {
		recordedAt := input.RecordedAt.UTC()
		params.RecordedAt = &recordedAt
	}
	if input.Value != nil {
		value := *input.Value
		params.Value = &value
	}
	if input.Unit != nil {
		unit := *input.Unit
		params.Unit = &unit
	}
	if input.Note != nil {
		note, err := normalizeOptionalText(input.Note, "note")
		if err != nil {
			return WeightEntry{}, err
		}
		params.Note = note
	}

	return s.repo.UpdateWeightEntry(ctx, params)
}

func (s *service) DeleteWeight(ctx context.Context, id int) error {
	if err := validateWeightID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteWeightEntry(ctx, DeleteWeightEntryParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}

func (s *service) WeightTrend(ctx context.Context, params WeightTrendParams) (WeightTrend, error) {
	rangeValue, err := normalizeWeightRange(params.Range)
	if err != nil {
		return WeightTrend{}, err
	}

	weights, err := s.repo.ListWeightEntries(ctx, HistoryFilter{})
	if err != nil {
		return WeightTrend{}, err
	}

	ascending := sortWeightEntriesAscending(weights)
	latest := firstWeightEntryFromEnd(ascending)
	start := rangeStart(rangeValue, latest)
	filtered := ascending
	if start != nil {
		filtered = filteredWeightEntries(ascending, *start)
	}

	return WeightTrend{
		Range:                 rangeValue,
		RawPoints:             filtered,
		MovingAveragePoints:   movingAveragePoints(filtered),
		MonthlyAverageBuckets: monthlyAverageBuckets(filtered),
	}, nil
}
