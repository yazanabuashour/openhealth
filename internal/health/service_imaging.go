package health

import "context"

func (s *service) ListImaging(ctx context.Context, params ImagingListParams) ([]ImagingRecord, error) {
	if err := validateHistoryFilter(params.HistoryFilter); err != nil {
		return nil, err
	}
	modality, err := normalizeOptionalText(params.Modality, "modality")
	if err != nil {
		return nil, err
	}
	bodySite, err := normalizeOptionalText(params.BodySite, "body_site")
	if err != nil {
		return nil, err
	}
	params.Modality = modality
	params.BodySite = bodySite
	return s.repo.ListImagingRecords(ctx, params)
}

func (s *service) CreateImaging(ctx context.Context, input ImagingRecordInput) (ImagingRecord, error) {
	normalized, err := normalizeImagingRecordInput(input)
	if err != nil {
		return ImagingRecord{}, err
	}
	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return ImagingRecord{}, &DatabaseError{
			Message: "failed to generate imaging record hash",
			Cause:   err,
		}
	}
	now := s.now().UTC()
	return s.repo.CreateImagingRecord(ctx, CreateImagingRecordParams{
		PerformedAt:      normalized.PerformedAt,
		Modality:         normalized.Modality,
		BodySite:         normalized.BodySite,
		Title:            normalized.Title,
		Summary:          normalized.Summary,
		Impression:       normalized.Impression,
		Note:             normalized.Note,
		Notes:            normalized.Notes,
		Source:           manualSource,
		SourceRecordHash: sourceRecordHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
}

func (s *service) ReplaceImaging(ctx context.Context, id int, input ImagingRecordInput) (ImagingRecord, error) {
	if err := validateRecordID(id); err != nil {
		return ImagingRecord{}, err
	}
	normalized, err := normalizeImagingRecordInput(input)
	if err != nil {
		return ImagingRecord{}, err
	}
	return s.repo.UpdateImagingRecord(ctx, UpdateImagingRecordParams{
		ID:          id,
		PerformedAt: normalized.PerformedAt,
		Modality:    normalized.Modality,
		BodySite:    normalized.BodySite,
		Title:       normalized.Title,
		Summary:     normalized.Summary,
		Impression:  normalized.Impression,
		Note:        normalized.Note,
		Notes:       normalized.Notes,
		UpdatedAt:   s.now().UTC(),
	})
}

func (s *service) DeleteImaging(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteImagingRecord(ctx, DeleteImagingRecordParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}
