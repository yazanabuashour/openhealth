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

func (r *Repository) ListSleepEntries(ctx context.Context, filter health.HistoryFilter) ([]health.SleepEntry, error) {
	rows, err := r.queries.ListSleepEntries(ctx, sqlc.ListSleepEntriesParams{
		FromRecordedAt: nullableInstant(filter.From),
		ToRecordedAt:   nullableInstant(filter.To),
		LimitCount:     nullableLimit(filter.Limit),
	})
	if err != nil {
		return nil, wrapDatabaseError("failed to list health sleep entries", err)
	}
	items := make([]health.SleepEntry, 0, len(rows))
	for _, row := range rows {
		item, err := toSleepEntry(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) FindManualSleepEntry(ctx context.Context, params health.FindManualSleepEntryParams) (*health.SleepEntry, error) {
	row, err := r.queries.FindManualSleepEntry(ctx, params.RecordedAt.UTC().Format(time.DateOnly))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, wrapDatabaseError("failed to find manual health sleep entry", err)
	}
	entry, err := toSleepEntry(row)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *Repository) CreateSleepEntry(ctx context.Context, params health.CreateSleepEntryParams) (health.SleepEntry, error) {
	row, err := r.queries.CreateSleepEntry(ctx, sqlc.CreateSleepEntryParams{
		RecordedAt:       serializeInstant(params.RecordedAt),
		QualityScore:     int64(params.QualityScore),
		WakeupCount:      nullableInt64(params.WakeupCount),
		Note:             params.Note,
		Source:           params.Source,
		SourceRecordHash: params.SourceRecordHash,
		CreatedAt:        serializeInstant(params.CreatedAt),
		UpdatedAt:        serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		if isManualSleepUniqueConstraintError(err) {
			return health.SleepEntry{}, &health.ConflictError{Message: "manual sleep entry already exists for recorded date"}
		}
		return health.SleepEntry{}, wrapDatabaseError("failed to create health sleep entry", err)
	}
	return toSleepEntry(row)
}

func (r *Repository) UpdateSleepEntry(ctx context.Context, params health.UpdateSleepEntryParams) (health.SleepEntry, error) {
	row, err := r.queries.UpdateSleepEntry(ctx, sqlc.UpdateSleepEntryParams{
		RecordedAt:   nullableInstantPointer(params.RecordedAt),
		QualityScore: nullableInt64(params.QualityScore),
		WakeupCount:  nullableInt64(params.WakeupCount),
		Note:         params.Note,
		UpdatedAt:    serializeInstant(params.UpdatedAt),
		ID:           int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return health.SleepEntry{}, &health.NotFoundError{
				Resource: "health_sleep_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		if isManualSleepUniqueConstraintError(err) {
			return health.SleepEntry{}, &health.ConflictError{Message: "manual sleep entry already exists for recorded date"}
		}
		return health.SleepEntry{}, wrapDatabaseError("failed to update health sleep entry", err)
	}
	return toSleepEntry(row)
}

func (r *Repository) DeleteSleepEntry(ctx context.Context, params health.DeleteSleepEntryParams) error {
	deletedAt := serializeInstant(params.DeletedAt)
	_, err := r.queries.DeleteSleepEntry(ctx, sqlc.DeleteSleepEntryParams{
		DeletedAt: &deletedAt,
		UpdatedAt: serializeInstant(params.UpdatedAt),
		ID:        int64(params.ID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &health.NotFoundError{
				Resource: "health_sleep_entry",
				ID:       fmt.Sprintf("%d", params.ID),
			}
		}
		return wrapDatabaseError("failed to delete health sleep entry", err)
	}
	return nil
}

func toSleepEntry(row sqlc.HealthSleepEntry) (health.SleepEntry, error) {
	recordedAt, err := parseInstant(row.RecordedAt)
	if err != nil {
		return health.SleepEntry{}, err
	}
	createdAt, err := parseInstant(row.CreatedAt)
	if err != nil {
		return health.SleepEntry{}, err
	}
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return health.SleepEntry{}, err
	}
	deletedAt, err := parseOptionalInstant(row.DeletedAt)
	if err != nil {
		return health.SleepEntry{}, err
	}
	return health.SleepEntry{
		ID:               int(row.ID),
		RecordedAt:       recordedAt,
		QualityScore:     int(row.QualityScore),
		WakeupCount:      nullableInt(row.WakeupCount),
		Note:             row.Note,
		Source:           row.Source,
		SourceRecordHash: row.SourceRecordHash,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
	}, nil
}

func isManualSleepUniqueConstraintError(err error) bool {
	var sqliteErr *moderncsqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code() != sqliteConstraintUnique {
		return false
	}
	return strings.Contains(sqliteErr.Error(), "idx_health_sleep_entry_manual_date_unique")
}
