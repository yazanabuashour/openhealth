package health

import "context"

func (s *service) ListBodyComposition(ctx context.Context, filter HistoryFilter) ([]BodyCompositionEntry, error) {
	if err := validateHistoryFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.ListBodyCompositionEntries(ctx, filter)
}

func (s *service) CreateBodyComposition(ctx context.Context, input BodyCompositionInput) (BodyCompositionEntry, error) {
	normalized, err := normalizeBodyCompositionInput(input)
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return BodyCompositionEntry{}, &DatabaseError{
			Message: "failed to generate body composition entry hash",
			Cause:   err,
		}
	}
	now := s.now().UTC()
	return s.repo.CreateBodyCompositionEntry(ctx, CreateBodyCompositionEntryParams{
		RecordedAt:       normalized.RecordedAt,
		BodyFatPercent:   normalized.BodyFatPercent,
		WeightValue:      normalized.WeightValue,
		WeightUnit:       normalized.WeightUnit,
		Method:           normalized.Method,
		Note:             normalized.Note,
		Source:           manualSource,
		SourceRecordHash: sourceRecordHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func (s *service) ReplaceBodyComposition(ctx context.Context, id int, input BodyCompositionInput) (BodyCompositionEntry, error) {
	if err := validateRecordID(id); err != nil {
		return BodyCompositionEntry{}, err
	}
	normalized, err := normalizeBodyCompositionInput(input)
	if err != nil {
		return BodyCompositionEntry{}, err
	}
	return s.repo.UpdateBodyCompositionEntry(ctx, UpdateBodyCompositionEntryParams{
		ID:             id,
		RecordedAt:     normalized.RecordedAt,
		BodyFatPercent: normalized.BodyFatPercent,
		WeightValue:    normalized.WeightValue,
		WeightUnit:     normalized.WeightUnit,
		Method:         normalized.Method,
		Note:           normalized.Note,
		UpdatedAt:      s.now().UTC(),
	})
}

func (s *service) DeleteBodyComposition(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteBodyCompositionEntry(ctx, DeleteBodyCompositionEntryParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}
