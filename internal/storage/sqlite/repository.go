package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	moderncsqlite "modernc.org/sqlite"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

const sqliteConstraintUnique = 2067

type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:      db,
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

func (r *Repository) FindManualWeightEntry(ctx context.Context, params health.FindManualWeightEntryParams) (*health.WeightEntry, error) {
	row, err := r.queries.FindManualWeightEntry(ctx, sqlc.FindManualWeightEntryParams{
		RecordedDate: params.RecordedAt.UTC().Format(time.DateOnly),
		Unit:         string(params.Unit),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, wrapDatabaseError("failed to find manual health weight entry", err)
	}
	entry, err := toWeightEntry(row)
	if err != nil {
		return nil, err
	}
	return &entry, nil
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
		if isManualWeightUniqueConstraintError(err) {
			return health.WeightEntry{}, &health.ConflictError{Message: "manual weight already exists for recorded date and unit"}
		}
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

func (r *Repository) CreateBloodPressureEntry(ctx context.Context, params health.CreateBloodPressureEntryParams) (health.BloodPressureEntry, error) {
	row, err := r.queries.CreateBloodPressureEntry(ctx, sqlc.CreateBloodPressureEntryParams{
		RecordedAt:       serializeInstant(params.RecordedAt),
		Systolic:         int64(params.Systolic),
		Diastolic:        int64(params.Diastolic),
		Pulse:            nullableInt64(params.Pulse),
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.BloodPressureEntry{}, wrapDatabaseError("failed to create health blood pressure entry", err)
	}
	return toBloodPressureEntry(row)
}

func (r *Repository) UpdateBloodPressureEntry(ctx context.Context, params health.UpdateBloodPressureEntryParams) (health.BloodPressureEntry, error) {
	row, err := r.queries.UpdateBloodPressureEntry(ctx, sqlc.UpdateBloodPressureEntryParams{
		RecordedAt: serializeInstant(params.RecordedAt),
		Systolic:   int64(params.Systolic),
		Diastolic:  int64(params.Diastolic),
		Pulse:      nullableInt64(params.Pulse),
		UpdatedAt:  serializeInstant(params.UpdatedAt),
		ID:         int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.BloodPressureEntry{}, &health.NotFoundError{
				Resource: "health_blood_pressure_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.BloodPressureEntry{}, wrapDatabaseError("failed to update health blood pressure entry", err)
	}
	return toBloodPressureEntry(row)
}

func (r *Repository) DeleteBloodPressureEntry(ctx context.Context, params health.DeleteBloodPressureEntryParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteBloodPressureEntry(ctx, sqlc.DeleteBloodPressureEntryParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_blood_pressure_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health blood pressure entry", err)
	}
	return nil
}

func (r *Repository) ListMedicationCourses(ctx context.Context, status health.MedicationStatus, today string) ([]health.MedicationCourse, error) {
	items := []health.MedicationCourse{}
	switch status {
	case health.MedicationStatusActive:
		rows, err := r.queries.ListActiveMedicationCourses(ctx, &today)
		if err != nil {
			return nil, wrapDatabaseError("failed to list health medications", err)
		}
		for _, row := range rows {
			item, err := toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	case health.MedicationStatusAll:
		rows, err := r.queries.ListMedicationCourses(ctx)
		if err != nil {
			return nil, wrapDatabaseError("failed to list health medications", err)
		}
		for _, row := range rows {
			item, err := toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
	default:
		return nil, &health.ValidationError{Message: "status must be 'active' or 'all'"}
	}
	return items, nil
}

func (r *Repository) CreateMedicationCourse(ctx context.Context, params health.CreateMedicationCourseParams) (health.MedicationCourse, error) {
	row, err := r.queries.CreateMedicationCourse(ctx, sqlc.CreateMedicationCourseParams{
		Name:       params.Name,
		DosageText: params.DosageText,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Note:       params.Note,
		Source:     params.Source,
		CreatedAt:  serializeInstant(params.CreatedAt),
		UpdatedAt:  serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.MedicationCourse{}, wrapDatabaseError("failed to create health medication course", err)
	}
	return toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
}

func (r *Repository) UpdateMedicationCourse(ctx context.Context, params health.UpdateMedicationCourseParams) (health.MedicationCourse, error) {
	row, err := r.queries.UpdateMedicationCourse(ctx, sqlc.UpdateMedicationCourseParams{
		Name:       params.Name,
		DosageText: params.DosageText,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Note:       params.Note,
		UpdatedAt:  serializeInstant(params.UpdatedAt),
		ID:         int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.MedicationCourse{}, &health.NotFoundError{
				Resource: "health_medication_course",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.MedicationCourse{}, wrapDatabaseError("failed to update health medication course", err)
	}
	return toMedicationCourseFields(row.ID, row.Name, row.DosageText, row.StartDate, row.EndDate, row.Note, row.Source, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
}

func (r *Repository) DeleteMedicationCourse(ctx context.Context, params health.DeleteMedicationCourseParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteMedicationCourse(ctx, sqlc.DeleteMedicationCourseParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_medication_course",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health medication course", err)
	}
	return nil
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
		item, err := toLabCollectionFields(
			collection.ID,
			collection.CollectedAt,
			collection.Note,
			collection.Source,
			collection.CreatedAt,
			collection.UpdatedAt,
			collection.DeletedAt,
			panelsByCollection[int(collection.ID)],
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) CreateLabCollection(ctx context.Context, params health.CreateLabCollectionParams) (health.LabCollection, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return health.LabCollection{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := r.queries.WithTx(tx)
	updatedAt := serializeInstant(params.UpdatedAt)
	collection, err := queries.CreateLabCollection(ctx, sqlc.CreateLabCollectionParams{
		CollectedAt: serializeInstant(params.CollectedAt),
		Note:        params.Note,
		Source:      params.Source,
		CreatedAt:   serializeInstant(params.CreatedAt),
		UpdatedAt:   &updatedAt,
	})
	if err != nil {
		return health.LabCollection{}, wrapDatabaseError("failed to create health lab collection", err)
	}
	if err := r.replaceLabPanels(ctx, queries, int(collection.ID), params.Panels); err != nil {
		return health.LabCollection{}, err
	}
	result, err := r.getLabCollectionWithQueries(ctx, queries, int(collection.ID))
	if err != nil {
		return health.LabCollection{}, err
	}
	if err := tx.Commit(); err != nil {
		return health.LabCollection{}, err
	}
	return result, nil
}

func (r *Repository) UpdateLabCollection(ctx context.Context, params health.UpdateLabCollectionParams) (health.LabCollection, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return health.LabCollection{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := r.queries.WithTx(tx)
	updatedAt := serializeInstant(params.UpdatedAt)
	collection, err := queries.UpdateLabCollection(ctx, sqlc.UpdateLabCollectionParams{
		CollectedAt: serializeInstant(params.CollectedAt),
		Note:        params.Note,
		UpdatedAt:   &updatedAt,
		ID:          int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.LabCollection{}, &health.NotFoundError{
				Resource: "health_lab_collection",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.LabCollection{}, wrapDatabaseError("failed to update health lab collection", err)
	}
	if err := queries.DeleteLabPanelsByCollection(ctx, collection.ID); err != nil {
		return health.LabCollection{}, wrapDatabaseError("failed to replace health lab panels", err)
	}
	if err := r.replaceLabPanels(ctx, queries, int(collection.ID), params.Panels); err != nil {
		return health.LabCollection{}, err
	}
	result, err := r.getLabCollectionWithQueries(ctx, queries, int(collection.ID))
	if err != nil {
		return health.LabCollection{}, err
	}
	if err := tx.Commit(); err != nil {
		return health.LabCollection{}, err
	}
	return result, nil
}

func (r *Repository) DeleteLabCollection(ctx context.Context, params health.DeleteLabCollectionParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	updatedAt := serializeInstant(params.UpdatedAt)
	_, err := r.queries.DeleteLabCollection(ctx, sqlc.DeleteLabCollectionParams{
		DeletedAt: &deletedAt,
		UpdatedAt: &updatedAt,
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_lab_collection",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health lab collection", err)
	}
	return nil
}

func (r *Repository) replaceLabPanels(ctx context.Context, queries *sqlc.Queries, collectionID int, panels []health.LabPanelWriteParams) error {
	for _, panel := range panels {
		createdPanel, err := queries.CreateLabPanel(ctx, sqlc.CreateLabPanelParams{
			CollectionID: int64(collectionID),
			PanelName:    panel.PanelName,
			DisplayOrder: int64(panel.DisplayOrder),
		})
		if err != nil {
			return wrapDatabaseError("failed to create health lab panel", err)
		}
		for _, result := range panel.Results {
			_, err := queries.CreateLabResult(ctx, sqlc.CreateLabResultParams{
				PanelID:       createdPanel.ID,
				TestName:      result.TestName,
				CanonicalSlug: nullableAnalyteSlug(result.CanonicalSlug),
				ValueText:     result.ValueText,
				ValueNumeric:  result.ValueNumeric,
				Units:         result.Units,
				RangeText:     result.RangeText,
				Flag:          result.Flag,
				DisplayOrder:  int64(result.DisplayOrder),
			})
			if err != nil {
				return wrapDatabaseError("failed to create health lab result", err)
			}
		}
	}
	return nil
}

func (r *Repository) getLabCollectionWithQueries(ctx context.Context, queries *sqlc.Queries, id int) (health.LabCollection, error) {
	collection, err := queries.GetLabCollection(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.LabCollection{}, &health.NotFoundError{
				Resource: "health_lab_collection",
				ID:       fmt.Sprintf("%d", id),
			}
		}
		return health.LabCollection{}, wrapDatabaseError("failed to get health lab collection", err)
	}
	panels, err := queries.ListLabPanels(ctx)
	if err != nil {
		return health.LabCollection{}, wrapDatabaseError("failed to list health lab panels", err)
	}
	results, err := queries.ListLabResults(ctx)
	if err != nil {
		return health.LabCollection{}, wrapDatabaseError("failed to list health lab results", err)
	}
	resultsByPanel := make(map[int][]health.LabResult)
	for _, result := range results {
		item := toLabResult(result)
		resultsByPanel[item.PanelID] = append(resultsByPanel[item.PanelID], item)
	}
	outPanels := make([]health.LabPanel, 0)
	for _, panel := range panels {
		if int(panel.CollectionID) != id {
			continue
		}
		outPanels = append(outPanels, health.LabPanel{
			ID:           int(panel.ID),
			CollectionID: int(panel.CollectionID),
			PanelName:    panel.PanelName,
			DisplayOrder: int(panel.DisplayOrder),
			Results:      resultsByPanel[int(panel.ID)],
		})
	}
	return toLabCollectionFields(
		collection.ID,
		collection.CollectedAt,
		collection.Note,
		collection.Source,
		collection.CreatedAt,
		collection.UpdatedAt,
		collection.DeletedAt,
		outPanels,
	)
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

func (r *Repository) ListBodyCompositionEntries(ctx context.Context, filter health.HistoryFilter) ([]health.BodyCompositionEntry, error) {
	rows, err := r.queries.ListBodyCompositionEntries(ctx, sqlc.ListBodyCompositionEntriesParams{
		FromRecordedAt: nullableInstant(filter.From),
		ToRecordedAt:   nullableInstant(filter.To),
		LimitCount:     nullableLimit(filter.Limit),
	})
	if err != nil {
		return nil, wrapDatabaseError("failed to list health body composition entries", err)
	}
	items := make([]health.BodyCompositionEntry, 0, len(rows))
	for _, row := range rows {
		item, err := toBodyCompositionEntry(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateBodyCompositionEntry(ctx context.Context, params health.CreateBodyCompositionEntryParams) (health.BodyCompositionEntry, error) {
	row, err := r.queries.CreateBodyCompositionEntry(ctx, sqlc.CreateBodyCompositionEntryParams{
		RecordedAt:       serializeInstant(params.RecordedAt),
		BodyFatPercent:   params.BodyFatPercent,
		WeightValue:      params.WeightValue,
		WeightUnit:       nullableWeightUnit(params.WeightUnit),
		Method:           params.Method,
		Note:             params.Note,
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.BodyCompositionEntry{}, wrapDatabaseError("failed to create health body composition entry", err)
	}
	return toBodyCompositionEntry(row)
}

func (r *Repository) UpdateBodyCompositionEntry(ctx context.Context, params health.UpdateBodyCompositionEntryParams) (health.BodyCompositionEntry, error) {
	row, err := r.queries.UpdateBodyCompositionEntry(ctx, sqlc.UpdateBodyCompositionEntryParams{
		RecordedAt:     serializeInstant(params.RecordedAt),
		BodyFatPercent: params.BodyFatPercent,
		WeightValue:    params.WeightValue,
		WeightUnit:     nullableWeightUnit(params.WeightUnit),
		Method:         params.Method,
		Note:           params.Note,
		UpdatedAt:      serializeInstant(params.UpdatedAt),
		ID:             int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.BodyCompositionEntry{}, &health.NotFoundError{
				Resource: "health_body_composition_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.BodyCompositionEntry{}, wrapDatabaseError("failed to update health body composition entry", err)
	}
	return toBodyCompositionEntry(row)
}

func (r *Repository) DeleteBodyCompositionEntry(ctx context.Context, params health.DeleteBodyCompositionEntryParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteBodyCompositionEntry(ctx, sqlc.DeleteBodyCompositionEntryParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_body_composition_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health body composition entry", err)
	}
	return nil
}

func (r *Repository) ListImagingRecords(ctx context.Context, params health.ImagingListParams) ([]health.ImagingRecord, error) {
	rows, err := r.queries.ListImagingRecords(ctx, sqlc.ListImagingRecordsParams{
		FromPerformedAt: nullableInstant(params.From),
		ToPerformedAt:   nullableInstant(params.To),
		Modality:        nullableString(params.Modality),
		BodySite:        nullableString(params.BodySite),
		LimitCount:      nullableLimit(params.Limit),
	})
	if err != nil {
		return nil, wrapDatabaseError("failed to list health imaging records", err)
	}
	items := make([]health.ImagingRecord, 0, len(rows))
	for _, row := range rows {
		item, err := toImagingRecord(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateImagingRecord(ctx context.Context, params health.CreateImagingRecordParams) (health.ImagingRecord, error) {
	row, err := r.queries.CreateImagingRecord(ctx, sqlc.CreateImagingRecordParams{
		PerformedAt:      serializeInstant(params.PerformedAt),
		Modality:         params.Modality,
		BodySite:         params.BodySite,
		Title:            params.Title,
		Summary:          params.Summary,
		Impression:       params.Impression,
		Note:             params.Note,
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.ImagingRecord{}, wrapDatabaseError("failed to create health imaging record", err)
	}
	return toImagingRecord(row)
}

func (r *Repository) UpdateImagingRecord(ctx context.Context, params health.UpdateImagingRecordParams) (health.ImagingRecord, error) {
	row, err := r.queries.UpdateImagingRecord(ctx, sqlc.UpdateImagingRecordParams{
		PerformedAt: serializeInstant(params.PerformedAt),
		Modality:    params.Modality,
		BodySite:    params.BodySite,
		Title:       params.Title,
		Summary:     params.Summary,
		Impression:  params.Impression,
		Note:        params.Note,
		UpdatedAt:   serializeInstant(params.UpdatedAt),
		ID:          int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.ImagingRecord{}, &health.NotFoundError{
				Resource: "health_imaging_record",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return health.ImagingRecord{}, wrapDatabaseError("failed to update health imaging record", err)
	}
	return toImagingRecord(row)
}

func (r *Repository) DeleteImagingRecord(ctx context.Context, params health.DeleteImagingRecordParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteImagingRecord(ctx, sqlc.DeleteImagingRecordParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_imaging_record",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health imaging record", err)
	}
	return nil
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

func toMedicationCourseFields(id int64, name string, dosageText *string, startDate string, endDate *string, note *string, source string, createdAtValue string, updatedAtValue string, deletedAtValue *string) (health.MedicationCourse, error) {
	createdAt, err := parseInstant(createdAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	updatedAt, err := parseInstant(updatedAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	deletedAt, err := parseOptionalInstant(deletedAtValue)
	if err != nil {
		return health.MedicationCourse{}, err
	}
	return health.MedicationCourse{
		ID:         int(id),
		Name:       name,
		DosageText: dosageText,
		StartDate:  startDate,
		EndDate:    endDate,
		Note:       note,
		Source:     source,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  deletedAt,
	}, nil
}

func toLabCollectionFields(id int64, collectedAtValue string, note *string, source string, createdAtValue string, updatedAtValue *string, deletedAtValue *string, panels []health.LabPanel) (health.LabCollection, error) {
	collectedAt, err := parseInstant(collectedAtValue)
	if err != nil {
		return health.LabCollection{}, err
	}
	createdAt, err := parseInstant(createdAtValue)
	if err != nil {
		return health.LabCollection{}, err
	}
	updatedAt := createdAt
	if updatedAtValue != nil {
		updatedAt, err = parseInstant(*updatedAtValue)
		if err != nil {
			return health.LabCollection{}, err
		}
	}
	deletedAt, err := parseOptionalInstant(deletedAtValue)
	if err != nil {
		return health.LabCollection{}, err
	}
	return health.LabCollection{
		ID:          int(id),
		CollectedAt: collectedAt,
		Note:        note,
		Source:      source,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
		Panels:      panels,
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

func toBodyCompositionEntry(row sqlc.HealthBodyCompositionEntry) (health.BodyCompositionEntry, error) {
	recordedAt, err := parseInstant(row.RecordedAt)
	if err != nil {
		return health.BodyCompositionEntry{}, err
	}
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.BodyCompositionEntry{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.BodyCompositionEntry{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.BodyCompositionEntry{}, err
	}
	return health.BodyCompositionEntry{
		ID:               int(row.ID),
		RecordedAt:       recordedAt,
		BodyFatPercent:   row.BodyFatPercent,
		WeightValue:      row.WeightValue,
		WeightUnit:       optionalWeightUnit(row.WeightUnit),
		Method:           row.Method,
		Note:             row.Note,
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

func toImagingRecord(row sqlc.HealthImagingRecord) (health.ImagingRecord, error) {
	performedAt, err := parseInstant(row.PerformedAt)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	return health.ImagingRecord{
		ID:               int(row.ID),
		PerformedAt:      performedAt,
		Modality:         row.Modality,
		BodySite:         row.BodySite,
		Title:            row.Title,
		Summary:          row.Summary,
		Impression:       row.Impression,
		Note:             row.Note,
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
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

func optionalWeightUnit(value *string) *health.WeightUnit {
	if value == nil {
		return nil
	}
	unit := health.WeightUnit(*value)
	return &unit
}

func nullableString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt64(value *int) *int64 {
	if value == nil {
		return nil
	}
	converted := int64(*value)
	return &converted
}

func nullableAnalyteSlug(value *health.AnalyteSlug) *string {
	if value == nil {
		return nil
	}
	converted := string(*value)
	return &converted
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

func isManualWeightUniqueConstraintError(err error) bool {
	var sqliteErr *moderncsqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code() != sqliteConstraintUnique {
		return false
	}
	return strings.Contains(sqliteErr.Error(), "idx_health_weight_entry_manual_date_unit_unique")
}
