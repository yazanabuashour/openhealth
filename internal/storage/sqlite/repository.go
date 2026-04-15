package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) ListWeightEntries(ctx context.Context, filter health.HistoryFilter) ([]health.WeightEntry, error) {
	rows, err := r.queries.ListWeightEntries(ctx, sqlc.ListWeightEntriesParams{
		FromRecordedAt: nullableInstant(filter.From),
		ToRecordedAt:   nullableInstant(filter.To),
		LimitCount:     nullableLimit(filter.Limit),
	})
	if err != nil {
		return nil, wrapDatabaseError("failed to list health weight entries", err)
	}

	items := make([]health.WeightEntry, 0, len(rows))
	for _, row := range rows {
		item, err := toWeightEntry(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateWeightEntry(ctx context.Context, params health.CreateWeightEntryParams) (health.WeightEntry, error) {
	row, err := r.queries.CreateWeightEntry(ctx, sqlc.CreateWeightEntryParams{
		RecordedAt:       serializeInstant(params.RecordedAt),
		Value:            params.Value,
		Unit:             string(params.Unit),
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		Note:             params.Note,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.WeightEntry{}, wrapDatabaseError("failed to create health weight entry", err)
	}
	return toWeightEntry(row)
}

func (r *Repository) UpdateWeightEntry(ctx context.Context, params health.UpdateWeightEntryParams) (health.WeightEntry, error) {
	row, err := r.queries.UpdateWeightEntry(ctx, sqlc.UpdateWeightEntryParams{
		RecordedAt: nullableInstantPointer(params.RecordedAt),
		Value:      params.Value,
		Unit:       nullableWeightUnit(params.Unit),
		UpdatedAt:  serializeInstant(params.UpdatedAt),
		ID:         int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.WeightEntry{}, &health.NotFoundError{
				Resource: "health_weight_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.WeightEntry{}, wrapDatabaseError("failed to update health weight entry", err)
	}
	return toWeightEntry(row)
}

func (r *Repository) DeleteWeightEntry(ctx context.Context, params health.DeleteWeightEntryParams) error {
	_, err := r.queries.DeleteWeightEntry(ctx, sqlc.DeleteWeightEntryParams{
		DeletedAt: nullableInstantPointer(&params.DeletedAt),
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_weight_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health weight entry", err)
	}
	return nil
}

func (r *Repository) ListBloodPressureEntries(ctx context.Context, filter health.HistoryFilter) ([]health.BloodPressureEntry, error) {
	rows, err := r.queries.ListBloodPressureEntries(ctx, sqlc.ListBloodPressureEntriesParams{
		FromRecordedAt: nullableInstant(filter.From),
		ToRecordedAt:   nullableInstant(filter.To),
		LimitCount:     nullableLimit(filter.Limit),
	})
	if err != nil {
		return nil, wrapDatabaseError("failed to list health blood pressure entries", err)
	}

	items := make([]health.BloodPressureEntry, 0, len(rows))
	for _, row := range rows {
		item, err := toBloodPressureEntry(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListMedicationCourses(ctx context.Context, status health.MedicationStatus, today string) ([]health.MedicationCourse, error) {
	var (
		rows []sqlc.HealthMedicationCourse
		err  error
	)

	switch status {
	case health.MedicationStatusActive:
		rows, err = r.queries.ListActiveMedicationCourses(ctx, &today)
	case health.MedicationStatusAll:
		rows, err = r.queries.ListMedicationCourses(ctx)
	default:
		return nil, &health.ValidationError{Message: "status must be 'active' or 'all'"}
	}
	if err != nil {
		return nil, wrapDatabaseError("failed to list health medications", err)
	}

	items := make([]health.MedicationCourse, 0, len(rows))
	for _, row := range rows {
		item, err := toMedicationCourse(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CountActiveMedicationCourses(ctx context.Context, today string) (int, error) {
	count, err := r.queries.CountActiveMedicationCourses(ctx, &today)
	if err != nil {
		return 0, wrapDatabaseError("failed to count active medications", err)
	}
	return int(count), nil
}

func (r *Repository) ListLabCollections(ctx context.Context) ([]health.LabCollection, error) {
	collections, err := r.queries.ListLabCollections(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health lab collections", err)
	}
	panels, err := r.queries.ListLabPanels(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health lab panels", err)
	}
	results, err := r.queries.ListLabResults(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health lab results", err)
	}

	resultsByPanel := make(map[int][]health.LabResult)
	for _, result := range results {
		item := toLabResult(result)
		resultsByPanel[item.PanelID] = append(resultsByPanel[item.PanelID], item)
	}

	panelsByCollection := make(map[int][]health.LabPanel)
	for _, panel := range panels {
		item := health.LabPanel{
			ID:           int(panel.ID),
			CollectionID: int(panel.CollectionID),
			PanelName:    panel.PanelName,
			DisplayOrder: int(panel.DisplayOrder),
			Results:      resultsByPanel[int(panel.ID)],
		}
		panelsByCollection[item.CollectionID] = append(panelsByCollection[item.CollectionID], item)
	}

	items := make([]health.LabCollection, 0, len(collections))
	for _, collection := range collections {
		collectedAt, err := parseInstant(collection.CollectedAt)
		if err != nil {
			return nil, err
		}
		createdAt, err := parseInstant(collection.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, health.LabCollection{
			ID:          int(collection.ID),
			CollectedAt: collectedAt,
			Source:      collection.Source,
			CreatedAt:   createdAt,
			Panels:      panelsByCollection[int(collection.ID)],
		})
	}

	return items, nil
}

func (r *Repository) ListLabResultsWithCollection(ctx context.Context) ([]health.LabResultWithCollection, error) {
	rows, err := r.queries.ListLabResultsWithCollection(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list analyte results", err)
	}

	items := make([]health.LabResultWithCollection, 0, len(rows))
	for _, row := range rows {
		collectedAt, err := parseInstant(row.CollectedAt)
		if err != nil {
			return nil, err
		}
		item := health.LabResultWithCollection{
			LabResult: health.LabResult{
				ID:            int(row.ID),
				PanelID:       int(row.PanelID),
				TestName:      row.TestName,
				CanonicalSlug: normalizeStoredSlug(row.CanonicalSlug),
				ValueText:     row.ValueText,
				ValueNumeric:  row.ValueNumeric,
				Units:         row.Units,
				RangeText:     row.RangeText,
				Flag:          row.Flag,
				DisplayOrder:  int(row.DisplayOrder),
			},
			CollectedAt:  collectedAt,
			CollectionID: int(row.CollectionID),
			PanelName:    row.PanelName,
		}
		items = append(items, item)
	}

	return items, nil
}

func toWeightEntry(row sqlc.HealthWeightEntry) (health.WeightEntry, error) {
	recordedAt, err := parseInstant(row.RecordedAt)
	if err != nil {
		return health.WeightEntry{}, err
	}
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.WeightEntry{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.WeightEntry{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.WeightEntry{}, err
	}
	return health.WeightEntry{
		ID:               int(row.ID),
		RecordedAt:       recordedAt,
		Value:            row.Value,
		Unit:             health.WeightUnit(row.Unit),
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		Note:             row.Note,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

func toBloodPressureEntry(row sqlc.HealthBloodPressureEntry) (health.BloodPressureEntry, error) {
	recordedAt, err := parseInstant(row.RecordedAt)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	return health.BloodPressureEntry{
		ID:               int(row.ID),
		RecordedAt:       recordedAt,
		Systolic:         int(row.Systolic),
		Diastolic:        int(row.Diastolic),
		Pulse:            nullableInt(row.Pulse),
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

func toMedicationCourse(row sqlc.HealthMedicationCourse) (health.MedicationCourse, error) {
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	return health.MedicationCourse{
		ID:         int(row.ID),
		Name:       row.Name,
		DosageText: row.DosageText,
		StartDate:  row.StartDate,
		EndDate:    row.EndDate,
		Source:     row.Source,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  deletedAt,
	}, nil
}

func toLabResult(row sqlc.HealthLabResult) health.LabResult {
	return health.LabResult{
		ID:            int(row.ID),
		PanelID:       int(row.PanelID),
		TestName:      row.TestName,
		CanonicalSlug: normalizeStoredSlug(row.CanonicalSlug),
		ValueText:     row.ValueText,
		ValueNumeric:  row.ValueNumeric,
		Units:         row.Units,
		RangeText:     row.RangeText,
		Flag:          row.Flag,
		DisplayOrder:  int(row.DisplayOrder),
	}
}

func parseInstant(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, &health.DatabaseError{
			Message: "stored timestamp is invalid",
			Cause:   err,
		}
	}
	return parsed.UTC(), nil
}

func parseOptionalInstant(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseInstant(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func serializeInstant(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func nullableInstant(value *time.Time) interface{} {
	if value == nil {
		return nil
	}
	return serializeInstant(*value)
}

func nullableInstantPointer(value *time.Time) *string {
	if value == nil {
		return nil
	}
	serialized := serializeInstant(*value)
	return &serialized
}

func nullableLimit(value *int) interface{} {
	if value == nil {
		return nil
	}
	return int64(*value)
}

func nullableWeightUnit(value *health.WeightUnit) *string {
	if value == nil {
		return nil
	}
	unit := string(*value)
	return &unit
}

func normalizeStoredSlug(value *string) *health.AnalyteSlug {
	if value == nil {
		return nil
	}
	slug, ok := health.NormalizeAnalyteSlug(*value)
	if !ok {
		return nil
	}
	return &slug
}

func nullableInt(value *int64) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func wrapDatabaseError(message string, cause error) error {
	return &health.DatabaseError{
		Message: message,
		Cause:   cause,
	}
}
