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
		Note:       params.Note,
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

func isManualWeightUniqueConstraintError(err error) bool {
	var sqliteErr *moderncsqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code() != sqliteConstraintUnique {
		return false
	}
	return strings.Contains(sqliteErr.Error(), "idx_health_weight_entry_manual_date_unit_unique")
}
