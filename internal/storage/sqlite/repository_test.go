package sqlite_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
	"github.com/yazanabuashour/openhealth/internal/testutil"
)

func TestRepositoryWeightLifecycle(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	ctx := context.Background()

	recordedAt := time.Date(2026, 3, 28, 13, 15, 0, 0, time.UTC)
	created, err := repo.CreateWeightEntry(ctx, health.CreateWeightEntryParams{
		RecordedAt:       recordedAt,
		Value:            150.2,
		Unit:             health.WeightUnitLb,
		Source:           "manual",
		SourceRecordHash: "weight-a",
		CreatedAt:        recordedAt,
		UpdatedAt:        recordedAt,
	})
	if err != nil {
		t.Fatalf("create weight entry: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected persisted id, got %d", created.ID)
	}

	items, err := repo.ListWeightEntries(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list weights: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 weight entry, got %d", len(items))
	}

	found, err := repo.FindManualWeightEntry(ctx, health.FindManualWeightEntryParams{
		RecordedAt: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		Unit:       health.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("find manual weight entry: %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Fatalf("found manual weight = %#v, want id %d", found, created.ID)
	}

	_, err = repo.CreateWeightEntry(ctx, health.CreateWeightEntryParams{
		RecordedAt:       time.Date(2026, 3, 28, 23, 59, 0, 0, time.UTC),
		Value:            149.9,
		Unit:             health.WeightUnitLb,
		Source:           "manual",
		SourceRecordHash: "weight-duplicate",
		CreatedAt:        recordedAt,
		UpdatedAt:        recordedAt,
	})
	var conflictErr *health.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("duplicate manual weight error = %v, want conflict", err)
	}

	updated, err := repo.UpdateWeightEntry(ctx, health.UpdateWeightEntryParams{
		ID:        created.ID,
		Value:     float64Pointer(149.8),
		UpdatedAt: recordedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update weight entry: %v", err)
	}
	if updated.Value != 149.8 {
		t.Fatalf("updated value = %v, want 149.8", updated.Value)
	}

	if err := repo.DeleteWeightEntry(ctx, health.DeleteWeightEntryParams{
		ID:        created.ID,
		DeletedAt: recordedAt.Add(2 * time.Hour),
		UpdatedAt: recordedAt.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete weight entry: %v", err)
	}

	items, err = repo.ListWeightEntries(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list weights after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected deleted weights to be hidden, got %d", len(items))
	}

	found, err = repo.FindManualWeightEntry(ctx, health.FindManualWeightEntryParams{
		RecordedAt: time.Date(2026, 3, 28, 23, 59, 0, 0, time.UTC),
		Unit:       health.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("find deleted manual weight entry: %v", err)
	}
	if found != nil {
		t.Fatalf("expected deleted manual weight to be hidden, got %#v", found)
	}

	_, err = repo.UpdateWeightEntry(ctx, health.UpdateWeightEntryParams{
		ID:        created.ID,
		Value:     float64Pointer(150),
		UpdatedAt: recordedAt.Add(3 * time.Hour),
	})
	var notFoundErr *health.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected not found after soft delete, got %v", err)
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}
