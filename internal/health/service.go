package health

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

const (
	manualSource       = "manual"
	manualWeightSource = manualSource
)

type HashGenerator func() (string, error)

type Option func(*service)

type service struct {
	repo          Repository
	now           func() time.Time
	hashGenerator HashGenerator
}

func WithClock(now func() time.Time) Option {
	return func(s *service) {
		s.now = now
	}
}

func WithHashGenerator(fn HashGenerator) Option {
	return func(s *service) {
		s.hashGenerator = fn
	}
}

func NewService(repo Repository, opts ...Option) Service {
	s := &service{
		repo: repo,
		now: func() time.Time {
			return time.Now().UTC()
		},
		hashGenerator: defaultHashGenerator,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

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
		ActiveMedicationCount: activeMedicationCount,
		LatestLabHighlights:   latestLabHighlights,
	}, nil
}

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
	if equalWeightValue(existing.Value, input.Value) {
		return WeightWriteResult{
			Entry:  existing,
			Status: WeightWriteStatusAlreadyExists,
		}, nil
	}

	value := input.Value
	entry, err := s.UpdateWeight(ctx, existing.ID, WeightUpdateInput{
		Value: &value,
	})
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

func (s *service) ListMedications(ctx context.Context, params MedicationListParams) ([]MedicationCourse, error) {
	status, err := normalizeMedicationStatus(params.Status)
	if err != nil {
		return nil, err
	}
	return s.repo.ListMedicationCourses(ctx, status, s.now().UTC().Format(time.DateOnly))
}

func (s *service) CreateMedicationCourse(ctx context.Context, input MedicationCourseInput) (MedicationCourse, error) {
	normalized, err := normalizeMedicationCourseInput(input)
	if err != nil {
		return MedicationCourse{}, err
	}
	now := s.now().UTC()
	return s.repo.CreateMedicationCourse(ctx, CreateMedicationCourseParams{
		Name:       normalized.Name,
		DosageText: normalized.DosageText,
		StartDate:  normalized.StartDate,
		EndDate:    normalized.EndDate,
		Source:     manualSource,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

func (s *service) ReplaceMedicationCourse(ctx context.Context, id int, input MedicationCourseInput) (MedicationCourse, error) {
	if err := validateRecordID(id); err != nil {
		return MedicationCourse{}, err
	}
	normalized, err := normalizeMedicationCourseInput(input)
	if err != nil {
		return MedicationCourse{}, err
	}
	return s.repo.UpdateMedicationCourse(ctx, UpdateMedicationCourseParams{
		ID:         id,
		Name:       normalized.Name,
		DosageText: normalized.DosageText,
		StartDate:  normalized.StartDate,
		EndDate:    normalized.EndDate,
		UpdatedAt:  s.now().UTC(),
	})
}

func (s *service) DeleteMedicationCourse(ctx context.Context, id int) error {
	if err := validateRecordID(id); err != nil {
		return err
	}
	now := s.now().UTC()
	return s.repo.DeleteMedicationCourse(ctx, DeleteMedicationCourseParams{
		ID:        id,
		DeletedAt: now,
		UpdatedAt: now,
	})
}

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

func defaultHashGenerator() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func validateHistoryFilter(filter HistoryFilter) error {
	if filter.Limit != nil && (*filter.Limit <= 0 || *filter.Limit > 3650) {
		return &ValidationError{Message: "limit must be between 1 and 3650"}
	}
	return nil
}

func normalizeWeightRecordInput(input WeightRecordInput) (WeightRecordInput, error) {
	if err := validateWeightRecordInput(input); err != nil {
		return WeightRecordInput{}, err
	}
	input.RecordedAt = input.RecordedAt.UTC()
	return input, nil
}

func validateWeightRecordInput(input WeightRecordInput) error {
	if input.RecordedAt.IsZero() {
		return &ValidationError{Message: "recorded_at is required"}
	}
	if input.Value <= 0 {
		return &ValidationError{Message: "value must be greater than 0"}
	}
	if input.Unit != WeightUnitLb {
		return &ValidationError{Message: "unit must be 'lb'"}
	}
	return nil
}

func normalizeBloodPressureRecordInput(input BloodPressureRecordInput) (BloodPressureRecordInput, error) {
	if input.RecordedAt.IsZero() {
		return BloodPressureRecordInput{}, &ValidationError{Message: "recorded_at is required"}
	}
	if input.Systolic <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "systolic must be greater than 0"}
	}
	if input.Diastolic <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "diastolic must be greater than 0"}
	}
	if input.Pulse != nil && *input.Pulse <= 0 {
		return BloodPressureRecordInput{}, &ValidationError{Message: "pulse must be greater than 0"}
	}
	input.RecordedAt = input.RecordedAt.UTC()
	return input, nil
}

func normalizeMedicationCourseInput(input MedicationCourseInput) (MedicationCourseInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return MedicationCourseInput{}, &ValidationError{Message: "name is required"}
	}
	if input.StartDate == "" {
		return MedicationCourseInput{}, &ValidationError{Message: "start_date is required"}
	}
	startDate, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return MedicationCourseInput{}, &ValidationError{Message: "start_date must be YYYY-MM-DD"}
	}
	if input.EndDate != nil {
		if *input.EndDate == "" {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be YYYY-MM-DD"}
		}
		endDate, err := time.Parse(time.DateOnly, *input.EndDate)
		if err != nil {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be YYYY-MM-DD"}
		}
		if endDate.Before(startDate) {
			return MedicationCourseInput{}, &ValidationError{Message: "end_date must be on or after start_date"}
		}
	}
	return input, nil
}

func normalizeLabCollectionInput(input LabCollectionInput) (LabCollectionInput, error) {
	if input.CollectedAt.IsZero() {
		return LabCollectionInput{}, &ValidationError{Message: "collected_at is required"}
	}
	if len(input.Panels) == 0 {
		return LabCollectionInput{}, &ValidationError{Message: "at least one lab panel is required"}
	}
	input.CollectedAt = input.CollectedAt.UTC()
	for panelIndex := range input.Panels {
		panel := &input.Panels[panelIndex]
		panel.PanelName = strings.TrimSpace(panel.PanelName)
		if panel.PanelName == "" {
			return LabCollectionInput{}, &ValidationError{Message: "panel_name is required"}
		}
		if len(panel.Results) == 0 {
			return LabCollectionInput{}, &ValidationError{Message: "at least one lab result is required"}
		}
		for resultIndex := range panel.Results {
			result := &panel.Results[resultIndex]
			result.TestName = strings.TrimSpace(result.TestName)
			if result.TestName == "" {
				return LabCollectionInput{}, &ValidationError{Message: "test_name is required"}
			}
			result.ValueText = strings.TrimSpace(result.ValueText)
			if result.ValueText == "" {
				return LabCollectionInput{}, &ValidationError{Message: "value_text is required"}
			}
			if result.CanonicalSlug != nil {
				validSlug, err := validateAnalyteSlug(*result.CanonicalSlug)
				if err != nil {
					return LabCollectionInput{}, err
				}
				result.CanonicalSlug = &validSlug
			}
		}
	}
	return input, nil
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

func equalWeightValue(left float64, right float64) bool {
	return math.Abs(left-right) < 0.000000001
}

func validateWeightUpdateInput(input WeightUpdateInput) error {
	if input.RecordedAt == nil && input.Value == nil && input.Unit == nil {
		return &ValidationError{Message: "At least one field must be provided"}
	}
	if input.Value != nil && *input.Value <= 0 {
		return &ValidationError{Message: "value must be greater than 0"}
	}
	if input.Unit != nil && *input.Unit != WeightUnitLb {
		return &ValidationError{Message: "unit must be 'lb'"}
	}
	return nil
}

func validateWeightID(id int) error {
	return validateRecordID(id)
}

func validateRecordID(id int) error {
	if id <= 0 {
		return &ValidationError{Message: "id must be greater than 0"}
	}
	return nil
}

func normalizeWeightRange(value WeightRange) (WeightRange, error) {
	if value == "" {
		return WeightRange90d, nil
	}
	switch value {
	case WeightRange30d, WeightRange90d, WeightRange1y, WeightRangeAll:
		return value, nil
	default:
		return "", &ValidationError{Message: "range must be one of 30d, 90d, 1y, all"}
	}
}

func normalizeMedicationStatus(value MedicationStatus) (MedicationStatus, error) {
	if value == "" {
		return MedicationStatusActive, nil
	}
	switch value {
	case MedicationStatusActive, MedicationStatusAll:
		return value, nil
	default:
		return "", &ValidationError{Message: "status must be 'active' or 'all'"}
	}
}

func validateAnalyteSlug(value AnalyteSlug) (AnalyteSlug, error) {
	slug, ok := NormalizeAnalyteSlug(string(value))
	if !ok {
		return "", &ValidationError{Message: "slug must be a valid analyte slug"}
	}
	return slug, nil
}

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

func firstBloodPressureEntry(entries []BloodPressureEntry) *BloodPressureEntry {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[0]
	return &entry
}

func intPointer(value int) *int {
	return &value
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
