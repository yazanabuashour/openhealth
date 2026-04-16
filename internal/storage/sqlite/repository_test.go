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

func TestRepositoryNonWeightLifecycles(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)

	bp, err := repo.CreateBloodPressureEntry(ctx, health.CreateBloodPressureEntryParams{
		RecordedAt:       now,
		Systolic:         118,
		Diastolic:        76,
		Pulse:            intPointer(64),
		Source:           "manual",
		SourceRecordHash: "bp-a",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("create blood pressure: %v", err)
	}
	bp, err = repo.UpdateBloodPressureEntry(ctx, health.UpdateBloodPressureEntryParams{
		ID:         bp.ID,
		RecordedAt: now.Add(time.Hour),
		Systolic:   119,
		Diastolic:  77,
		UpdatedAt:  now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update blood pressure: %v", err)
	}
	if bp.Systolic != 119 || bp.Pulse != nil {
		t.Fatalf("updated blood pressure = %#v, want systolic 119 with nil pulse", bp)
	}
	if err := repo.DeleteBloodPressureEntry(ctx, health.DeleteBloodPressureEntryParams{
		ID:        bp.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete blood pressure: %v", err)
	}
	bps, err := repo.ListBloodPressureEntries(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list blood pressure after delete: %v", err)
	}
	if len(bps) != 0 {
		t.Fatalf("deleted blood pressures should be hidden, got %#v", bps)
	}

	med, err := repo.CreateMedicationCourse(ctx, health.CreateMedicationCourseParams{
		Name:       "Levothyroxine",
		DosageText: stringPointer("25 mcg"),
		StartDate:  "2026-01-01",
		EndDate:    stringPointer("2026-12-31"),
		Source:     "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("create medication: %v", err)
	}
	med, err = repo.UpdateMedicationCourse(ctx, health.UpdateMedicationCourseParams{
		ID:        med.ID,
		Name:      "Levothyroxine",
		StartDate: "2026-01-02",
		UpdatedAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update medication: %v", err)
	}
	if med.StartDate != "2026-01-02" || med.EndDate != nil {
		t.Fatalf("updated medication = %#v, want new start date and nil end date", med)
	}
	if err := repo.DeleteMedicationCourse(ctx, health.DeleteMedicationCourseParams{
		ID:        med.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete medication: %v", err)
	}
	meds, err := repo.ListMedicationCourses(ctx, health.MedicationStatusAll, now.Format(time.DateOnly))
	if err != nil {
		t.Fatalf("list medications after delete: %v", err)
	}
	if len(meds) != 0 {
		t.Fatalf("deleted medications should be hidden, got %#v", meds)
	}

	lab, err := repo.CreateLabCollection(ctx, health.CreateLabCollectionParams{
		CollectedAt: now,
		Source:      "manual",
		CreatedAt:   now,
		UpdatedAt:   now,
		Panels: []health.LabPanelWriteParams{
			{
				PanelName:    "Thyroid",
				DisplayOrder: 0,
				Results: []health.LabResultWriteParams{
					{
						TestName:      "TSH",
						CanonicalSlug: analytePointer(health.AnalyteSlugTSH),
						ValueText:     "3.1",
						ValueNumeric:  float64Pointer(3.1),
						Units:         stringPointer("uIU/mL"),
						DisplayOrder:  0,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create lab collection: %v", err)
	}
	if len(lab.Panels) != 1 || len(lab.Panels[0].Results) != 1 {
		t.Fatalf("created lab collection = %#v", lab)
	}
	lab, err = repo.UpdateLabCollection(ctx, health.UpdateLabCollectionParams{
		ID:          lab.ID,
		CollectedAt: now.Add(24 * time.Hour),
		UpdatedAt:   now.Add(time.Hour),
		Panels: []health.LabPanelWriteParams{
			{
				PanelName:    "Metabolic",
				DisplayOrder: 0,
				Results: []health.LabResultWriteParams{
					{
						TestName:      "Glucose",
						CanonicalSlug: analytePointer(health.AnalyteSlugGlucose),
						ValueText:     "89",
						ValueNumeric:  float64Pointer(89),
						Units:         stringPointer("mg/dL"),
						DisplayOrder:  0,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("update lab collection: %v", err)
	}
	if lab.Panels[0].PanelName != "Metabolic" || lab.Panels[0].Results[0].TestName != "Glucose" {
		t.Fatalf("updated lab collection = %#v", lab)
	}
	results, err := repo.ListLabResultsWithCollection(ctx)
	if err != nil {
		t.Fatalf("list lab results with collection: %v", err)
	}
	if len(results) != 1 || results[0].TestName != "Glucose" {
		t.Fatalf("lab results = %#v, want only replacement result", results)
	}
	if err := repo.DeleteLabCollection(ctx, health.DeleteLabCollectionParams{
		ID:        lab.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete lab collection: %v", err)
	}
	labs, err := repo.ListLabCollections(ctx)
	if err != nil {
		t.Fatalf("list lab collections after delete: %v", err)
	}
	if len(labs) != 0 {
		t.Fatalf("deleted lab collections should be hidden, got %#v", labs)
	}
	results, err = repo.ListLabResultsWithCollection(ctx)
	if err != nil {
		t.Fatalf("list lab results after delete: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("deleted lab collection results should be hidden, got %#v", results)
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func analytePointer(value health.AnalyteSlug) *health.AnalyteSlug {
	return &value
}
