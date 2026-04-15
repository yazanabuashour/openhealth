package health

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math"
	"slices"
	"time"
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
	if err := validateWeightRecordInput(input); err != nil {
		return WeightEntry{}, err
	}

	sourceRecordHash, err := s.hashGenerator()
	if err != nil {
		return WeightEntry{}, &DatabaseError{
			Message: "failed to generate weight entry hash",
			Cause:   err,
		}
	}

	now := s.now().UTC()
	return s.repo.CreateWeightEntry(ctx, CreateWeightEntryParams{
		RecordedAt:       input.RecordedAt.UTC(),
		Value:            input.Value,
		Unit:             input.Unit,
		Source:           "manual",
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

func validateWeightRecordInput(input WeightRecordInput) error {
	if input.Value <= 0 {
		return &ValidationError{Message: "value must be greater than 0"}
	}
	if input.Unit != WeightUnitLb {
		return &ValidationError{Message: "unit must be 'lb'"}
	}
	return nil
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
	if _, ok := validAnalyteSlugs[value]; !ok {
		return "", &ValidationError{Message: "slug must be a supported analyte"}
	}
	return value, nil
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
	if result.CanonicalSlug == nil {
		return false
	}
	_, ok := validAnalyteSlugs[*result.CanonicalSlug]
	return ok
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
