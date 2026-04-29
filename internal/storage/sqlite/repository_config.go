package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

type ConfigValue struct {
	Key       string
	ValueJSON string
	UpdatedAt time.Time
}

type UpsertConfigValueParams struct {
	Key       string
	ValueJSON string
	UpdatedAt time.Time
}

func (r *Repository) GetConfigValue(ctx context.Context, key string) (*ConfigValue, error) {
	if err := validateConfigKey(key); err != nil {
		return nil, err
	}

	row, err := r.queries.GetConfigValue(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, wrapDatabaseError("failed to get OpenHealth config value", err)
	}

	value, err := toConfigValue(row)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *Repository) ListConfigValues(ctx context.Context) ([]ConfigValue, error) {
	rows, err := r.queries.ListConfigValues(ctx)
	if err != nil {
		return nil, wrapDatabaseError("failed to list OpenHealth config values", err)
	}

	values := make([]ConfigValue, 0, len(rows))
	for _, row := range rows {
		value, err := toConfigValue(row)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (r *Repository) UpsertConfigValue(ctx context.Context, params UpsertConfigValueParams) (ConfigValue, error) {
	if err := validateConfigKey(params.Key); err != nil {
		return ConfigValue{}, err
	}
	if !json.Valid([]byte(params.ValueJSON)) {
		return ConfigValue{}, fmt.Errorf("config value_json for %q must be valid JSON", params.Key)
	}

	row, err := r.queries.UpsertConfigValue(ctx, sqlc.UpsertConfigValueParams{
		Key:       params.Key,
		ValueJson: params.ValueJSON,
		UpdatedAt: serializeInstant(params.UpdatedAt),
	})
	if err != nil {
		return ConfigValue{}, wrapDatabaseError("failed to upsert OpenHealth config value", err)
	}
	return toConfigValue(row)
}

func (r *Repository) DeleteConfigValue(ctx context.Context, key string) (bool, error) {
	if err := validateConfigKey(key); err != nil {
		return false, err
	}

	deleted, err := r.queries.DeleteConfigValue(ctx, key)
	if err != nil {
		return false, wrapDatabaseError("failed to delete OpenHealth config value", err)
	}
	return deleted > 0, nil
}

func toConfigValue(row sqlc.OpenhealthConfig) (ConfigValue, error) {
	updatedAt, err := parseInstant(row.UpdatedAt)
	if err != nil {
		return ConfigValue{}, err
	}
	return ConfigValue{
		Key:       row.Key,
		ValueJSON: row.ValueJson,
		UpdatedAt: updatedAt,
	}, nil
}

func validateConfigKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("config key is required")
	}
	return nil
}
