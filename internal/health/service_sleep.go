package health

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *service) ListSleep(ctx context.Context, filter HistoryFilter) ([]SleepEntry, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.ListSleepEntries(ctx, filter)
}

func (s *service) UpsertSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error) {
	normalized, err := normalizeSleepInput(input)
	if err != nil {
		return SleepWriteResult{}, err
	}

	existing, err := s.repo.FindManualSleepEntry(ctx, FindManualSleepEntryParams{
		RecordedAt: normalized.RecordedAt,
	})
	if err != nil {
		return SleepWriteResult{}, err
	}
	if existing == nil {
		entry, err := s.createManualSleep(ctx, normalized)
		if err != nil {
			var conflictErr *ConflictError
			if errors.As(err, &conflictErr) {
				return s.upsertExistingSleep(ctx, normalized)
			}
			return SleepWriteResult{}, err
		}
		return SleepWriteResult{
			Entry:  entry,
			Status: SleepWriteStatusCreated,
		}, nil
	}
	return upsertSleepEntryValue(s, ctx, *existing, normalized)
}

func (s *service) upsertExistingSleep(ctx context.Context, input SleepInput) (SleepWriteResult, error) {
	existing, err := s.repo.FindManualSleepEntry(ctx, FindManualSleepEntryParams{
		RecordedAt: input.RecordedAt,
	})
	if err != nil {
		return SleepWriteResult{}, err
	}
	if existing == nil {
		return SleepWriteResult{}, &ConflictError{
			Message: fmt.Sprintf("manual sleep entry already exists on %s", input.RecordedAt.Format(time.DateOnly)),
		}
	}
	return upsertSleepEntryValue(s, ctx, *existing, input)
}

func upsertSleepEntryValue(s *service, ctx context.Context, existing SleepEntry, input SleepInput) (SleepWriteResult, error) {
	qualityChanged := existing.QualityScore != input.QualityScore
	wakeupChanged := input.WakeupCount != nil && !equalIntPointer(existing.WakeupCount, input.WakeupCount)
	noteChanged := input.Note != nil && !equalStringPointer(existing.Note, input.Note)
	if !qualityChanged && !wakeupChanged && !noteChanged {
		return SleepWriteResult{
			Entry:  existing,
			Status: SleepWriteStatusAlreadyExists,
		}, nil
	}

	update := UpdateSleepEntryParams{
		ID:        existing.ID,
		UpdatedAt: s.now().UTC(),
	}
	if qualityChanged {
		quality := input.QualityScore
		update.QualityScore = &quality
	}
	if wakeupChanged {
		update.WakeupCount = input.WakeupCount
	}
	if noteChanged {
		update.Note = input.Note
	}
	entry, err := s.repo.UpdateSleepEntry(ctx, update)
	if err != nil {
		return SleepWriteResult{}, err
	}
	return SleepWriteResult{
		Entry:  entry,
		Status: SleepWriteStatusUpdated,
	}, nil
}

func (s *service) createManualSleep(ctx context.Context, input SleepInput) (SleepEntry, error) {
	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return SleepEntry{}, &DatabaseError{
			Message: "failed to generate sleep entry hash",
			Cause:   err,
		}
	}
	now := s.now().UTC()
	return s.repo.CreateSleepEntry(ctx, CreateSleepEntryParams{
		RecordedAt:       input.RecordedAt,
		QualityScore:     input.QualityScore,
		WakeupCount:      input.WakeupCount,
		Note:             input.Note,
		Source:           manualSource,
		SourceRecordHash: sourceRecordHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func (s *service) DeleteSleep(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteSleepEntry(ctx, DeleteSleepEntryParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}
