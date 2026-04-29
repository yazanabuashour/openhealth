package sqlite

import (
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
)

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
