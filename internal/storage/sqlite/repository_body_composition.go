package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

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
