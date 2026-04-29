package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

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
	noteRows, err := r.queries.ListLabResultNotes(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health lab result notes", err)
	}
	notesByResult := labResultNotesByResultID(noteRows)

	resultsByPanel := make(map[int][]health.LabResult)
	for _, result := range results {
		item := toLabResult(result, notesByResult[int(result.ID)])
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
			createdResult, err := queries.CreateLabResult(ctx, sqlc.CreateLabResultParams{
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
			if err := r.createLabResultNotes(ctx, queries, int(createdResult.ID), result.Notes); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repository) createLabResultNotes(ctx context.Context, queries *sqlc.Queries, resultID int, notes []string) error {
	for i, note := range notes {
		_, err := queries.CreateLabResultNote(ctx, sqlc.CreateLabResultNoteParams{
			LabResultID:  int64(resultID),
			NoteText:     note,
			DisplayOrder: int64(i),
		})
		if err != nil {
			return wrapDatabaseError("failed to create health lab result note", err)
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
	noteRows, err := queries.ListLabResultNotes(ctx)
	if err != nil {
		return health.LabCollection{}, wrapDatabaseError("failed to list health lab result notes", err)
	}
	notesByResult := labResultNotesByResultID(noteRows)
	resultsByPanel := make(map[int][]health.LabResult)
	for _, result := range results {
		item := toLabResult(result, notesByResult[int(result.ID)])
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
	noteRows, err := r.queries.ListLabResultNotes(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health lab result notes", err)
	}
	notesByResult := labResultNotesByResultID(noteRows)

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
				Notes:         notesByResult[int(row.ID)],
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

func toLabResult(row sqlc.HealthLabResult, notes []string) health.LabResult {
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
		Notes:         append([]string(nil), notes...),
		DisplayOrder:  int(row.DisplayOrder),
	}
}

func labResultNotesByResultID(rows []sqlc.HealthLabResultNote) map[int][]string {
	out := make(map[int][]string)
	for _, row := range rows {
		out[int(row.LabResultID)] = append(out[int(row.LabResultID)], row.NoteText)
	}
	return out
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
