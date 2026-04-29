package health

import (
	"context"
	"slices"
)

func (s *service) ListAnalytes(ctx context.Context) ([]AnalyteSummary, error) {
	rows, err := s.repo.ListLabResultsWithCollection(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[AnalyteSlug][]LabResultWithCollection)
	orderedSlugs := []AnalyteSlug{}
	for _, row := range rows {
		if row.CanonicalSlug == nil {
			continue
		}
		slug := *row.CanonicalSlug
		if _, ok := grouped[slug]; !ok {
			orderedSlugs = append(orderedSlugs, slug)
		}
		grouped[slug] = append(grouped[slug], row)
	}

	result := make([]AnalyteSummary, 0, len(orderedSlugs))
	for _, slug := range orderedSlugs {
		group := grouped[slug]
		slices.SortFunc(group, compareLabResultWithCollectionDescending)
		if len(group) == 0 {
			continue
		}
		summary := AnalyteSummary{
			Slug:   slug,
			Latest: group[0],
		}
		if len(group) > 1 {
			previous := group[1]
			summary.Previous = &previous
		}
		result = append(result, summary)
	}

	return result, nil
}

func (s *service) AnalyteTrend(ctx context.Context, slug AnalyteSlug) (AnalyteTrend, error) {
	validSlug, err := validateAnalyteSlug(slug)
	if err != nil {
		return AnalyteTrend{}, err
	}

	rows, err := s.repo.ListLabResultsWithCollection(ctx)
	if err != nil {
		return AnalyteTrend{}, err
	}

	points := make([]LabResultWithCollection, 0)
	for _, row := range rows {
		if row.CanonicalSlug != nil && *row.CanonicalSlug == validSlug {
			points = append(points, row)
		}
	}
	if len(points) == 0 {
		return AnalyteTrend{}, &NotFoundError{
			Resource: "health_analyte",
			ID:       string(validSlug),
		}
	}

	ascending := slices.Clone(points)
	slices.SortFunc(ascending, compareLabResultWithCollectionAscending)
	descending := slices.Clone(points)
	slices.SortFunc(descending, compareLabResultWithCollectionDescending)

	trend := AnalyteTrend{
		Slug:   validSlug,
		Points: ascending,
	}
	latest := descending[0]
	trend.Latest = &latest
	if len(descending) > 1 {
		previous := descending[1]
		trend.Previous = &previous
	}
	return trend, nil
}

func (s *service) ListLabCollections(ctx context.Context) ([]LabCollection, error) {
	return s.repo.ListLabCollections(ctx)
}

func (s *service) CreateLabCollection(ctx context.Context, input LabCollectionInput) (LabCollection, error) {
	normalized, err := normalizeLabCollectionInput(input)
	if err != nil {
		return LabCollection{}, err
	}
	now := s.now().UTC()
	return s.repo.CreateLabCollection(ctx, CreateLabCollectionParams{
		CollectedAt: normalized.CollectedAt,
		Note:        normalized.Note,
		Source:      manualSource,
		CreatedAt:   now,
		UpdatedAt:   now,
		Panels:      labPanelWriteParams(normalized.Panels),
	})
}

func (s *service) ReplaceLabCollection(ctx context.Context, id int, input LabCollectionInput) (LabCollection, error) {
	if err := validateRecordID(id); err != nil {
		return LabCollection{}, err
	}
	normalized, err := normalizeLabCollectionInput(input)
	if err != nil {
		return LabCollection{}, err
	}
	return s.repo.UpdateLabCollection(ctx, UpdateLabCollectionParams{
		ID:          id,
		CollectedAt: normalized.CollectedAt,
		Note:        normalized.Note,
		UpdatedAt:   s.now().UTC(),
		Panels:      labPanelWriteParams(normalized.Panels),
	})
}

func (s *service) DeleteLabCollection(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteLabCollection(ctx, DeleteLabCollectionParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}

func labPanelWriteParams(panels []LabPanelInput) []LabPanelWriteParams {
	out := make([]LabPanelWriteParams, 0, len(panels))
	for panelIndex, panel := range panels {
		results := make([]LabResultWriteParams, 0, len(panel.Results))
		for resultIndex, result := range panel.Results {
			results = append(results, LabResultWriteParams{
				TestName:      result.TestName,
				CanonicalSlug: result.CanonicalSlug,
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				RangeText:     result.RangeText,
				Flag:          result.Flag,
				Notes:         result.Notes,
				DisplayOrder:  resultIndex,
			})
		}
		out = append(out, LabPanelWriteParams{
			PanelName:    panel.PanelName,
			DisplayOrder: panelIndex,
			Results:      results,
		})
	}
	return out
}
