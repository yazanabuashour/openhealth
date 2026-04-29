package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

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
		item, err := toBloodPressureEntryFields(row.ID, row.RecordedAt, row.Systolic, row.Diastolic, row.Pulse, row.Note, row.Source, row.SourceRecordHash, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
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
		Note:             params.Note,
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return health.BloodPressureEntry{}, wrapDatabaseError("failed to create health blood pressure entry", err)
	}
	return toBloodPressureEntryFields(row.ID, row.RecordedAt, row.Systolic, row.Diastolic, row.Pulse, row.Note, row.Source, row.SourceRecordHash, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
}

func (r *Repository) UpdateBloodPressureEntry(ctx context.Context, params health.UpdateBloodPressureEntryParams) (health.BloodPressureEntry, error) {
	row, err := r.queries.UpdateBloodPressureEntry(ctx, sqlc.UpdateBloodPressureEntryParams{
		RecordedAt: serializeInstant(params.RecordedAt),
		Systolic:   int64(params.Systolic),
		Diastolic:  int64(params.Diastolic),
		Pulse:      nullableInt64(params.Pulse),
		Note:       params.Note,
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
	return toBloodPressureEntryFields(row.ID, row.RecordedAt, row.Systolic, row.Diastolic, row.Pulse, row.Note, row.Source, row.SourceRecordHash, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
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

func toBloodPressureEntryFields(id int64, recordedAtValue string, systolic int64, diastolic int64, pulse *int64, note *string, source string, sourceRecordHash string, createdAtValue string, updatedAtValue string, deletedAtValue *string) (health.BloodPressureEntry, error) {
	recordedAt, err := parseInstant(recordedAtValue)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	createdAt, err := parseInstant(createdAtValue)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	updatedAt, err := parseInstant(updatedAtValue)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	deletedAt, err := parseOptionalInstant(deletedAtValue)
	if err != nil {
		return health.BloodPressureEntry{}, err
	}
	return health.BloodPressureEntry{
		ID:               int(id),
		RecordedAt:       recordedAt,
		Systolic:         int(systolic),
		Diastolic:        int(diastolic),
		Pulse:            nullableInt(pulse),
		Note:             note,
		Source:           source,
		SourceRecordHash: sourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}
