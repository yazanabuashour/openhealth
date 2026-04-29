package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

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
	noteRows, err := r.queries.ListImagingRecordNotes(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list health imaging record notes", err)
	}
	notesByRecord := imagingNotesByRecordID(noteRows)
	items := make([]health.ImagingRecord, 0, len(rows))
	for _, row := range rows {
		item, err := toImagingRecord(row, notesByRecord[int(row.ID)])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateImagingRecord(ctx context.Context, params health.CreateImagingRecordParams) (health.ImagingRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := r.queries.WithTx(tx)
	row, err := queries.CreateImagingRecord(ctx, sqlc.CreateImagingRecordParams{
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
	if err := r.createImagingRecordNotes(ctx, queries, int(row.ID), params.Notes); err != nil {
		return health.ImagingRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return health.ImagingRecord{}, err
	}
	return toImagingRecord(row, params.Notes)
}

func (r *Repository) UpdateImagingRecord(ctx context.Context, params health.UpdateImagingRecordParams) (health.ImagingRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return health.ImagingRecord{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := r.queries.WithTx(tx)
	row, err := queries.UpdateImagingRecord(ctx, sqlc.UpdateImagingRecordParams{
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
	if err := queries.DeleteImagingRecordNotesByRecord(ctx, int64(params.ID)); err != nil {
		return health.ImagingRecord{}, wrapDatabaseError("failed to replace health imaging record notes", err)
	}
	if err := r.createImagingRecordNotes(ctx, queries, int(row.ID), params.Notes); err != nil {
		return health.ImagingRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return health.ImagingRecord{}, err
	}
	return toImagingRecord(row, params.Notes)
}

func (r *Repository) createImagingRecordNotes(ctx context.Context, queries *sqlc.Queries, recordID int, notes []string) error {
	for i, note := range notes {
		_, err := queries.CreateImagingRecordNote(ctx, sqlc.CreateImagingRecordNoteParams{
			ImagingRecordID: int64(recordID),
			NoteText:        note,
			DisplayOrder:    int64(i),
		})
		if err != nil {
			return wrapDatabaseError("failed to create health imaging record note", err)
		}
	}
	return nil
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

func toImagingRecord(row sqlc.HealthImagingRecord, notes []string) (health.ImagingRecord, error) {
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
		Notes:            append([]string(nil), notes...),
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

func imagingNotesByRecordID(rows []sqlc.HealthImagingRecordNote) map[int][]string {
	out := make(map[int][]string)
	for _, row := range rows {
		out[int(row.ImagingRecordID)] = append(out[int(row.ImagingRecordID)], row.NoteText)
	}
	return out
}
