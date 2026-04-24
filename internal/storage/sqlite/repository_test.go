package sqlite_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
	"github.com/yazanabuashour/openhealth/internal/testutil"
)

func TestRepositoryConfigLifecycle(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	ctx := context.Background()
	updatedAt := time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC)

	missing, err := repo.GetConfigValue(ctx, "runner.default_limit")
	if err != nil {
		t.Fatalf("get missing config value: %v", err)
	}
	if missing != nil {
		t.Fatalf("missing config value = %#v, want nil", missing)
	}

	created, err := repo.UpsertConfigValue(ctx, sqlite.UpsertConfigValueParams{
		Key:       "runner.default_limit",
		ValueJSON: `{"value":10}`,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("upsert config value: %v", err)
	}
	if created.Key != "runner.default_limit" || created.ValueJSON != `{"value":10}` || !created.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("created config value = %#v", created)
	}

	reloaded, err := repo.GetConfigValue(ctx, "runner.default_limit")
	if err != nil {
		t.Fatalf("get config value: %v", err)
	}
	if reloaded == nil || reloaded.ValueJSON != `{"value":10}` {
		t.Fatalf("reloaded config value = %#v", reloaded)
	}

	updated, err := repo.UpsertConfigValue(ctx, sqlite.UpsertConfigValueParams{
		Key:       "runner.default_limit",
		ValueJSON: `{"value":25}`,
		UpdatedAt: updatedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update config value: %v", err)
	}
	if updated.ValueJSON != `{"value":25}` || !updated.UpdatedAt.Equal(updatedAt.Add(time.Hour)) {
		t.Fatalf("updated config value = %#v", updated)
	}

	values, err := repo.ListConfigValues(ctx)
	if err != nil {
		t.Fatalf("list config values: %v", err)
	}
	if len(values) != 1 || values[0].Key != "runner.default_limit" || values[0].ValueJSON != `{"value":25}` {
		t.Fatalf("config values = %#v", values)
	}

	deleted, err := repo.DeleteConfigValue(ctx, "runner.default_limit")
	if err != nil {
		t.Fatalf("delete config value: %v", err)
	}
	if !deleted {
		t.Fatal("delete config value returned false, want true")
	}
	deleted, err = repo.DeleteConfigValue(ctx, "runner.default_limit")
	if err != nil {
		t.Fatalf("delete missing config value: %v", err)
	}
	if deleted {
		t.Fatal("delete missing config value returned true, want false")
	}
}

func TestRepositoryConfigValidation(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	ctx := context.Background()

	_, err := repo.UpsertConfigValue(ctx, sqlite.UpsertConfigValueParams{
		Key:       "runner.default_limit",
		ValueJSON: `{"value":`,
		UpdatedAt: time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC),
	})
	if err == nil || !strings.Contains(err.Error(), "valid JSON") {
		t.Fatalf("invalid JSON error = %v, want valid JSON rejection", err)
	}

	_, err = repo.GetConfigValue(ctx, " ")
	if err == nil || !strings.Contains(err.Error(), "config key is required") {
		t.Fatalf("empty key error = %v, want key rejection", err)
	}
}

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
		Note:             stringPointer("morning scale"),
		CreatedAt:        recordedAt,
		UpdatedAt:        recordedAt,
	})
	if err != nil {
		t.Fatalf("create weight entry: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected persisted id, got %d", created.ID)
	}
	if created.Note == nil || *created.Note != "morning scale" {
		t.Fatalf("created weight note = %#v", created.Note)
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
		Note:      stringPointer("calibrated scale"),
		UpdatedAt: recordedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update weight entry: %v", err)
	}
	if updated.Value != 149.8 {
		t.Fatalf("updated value = %v, want 149.8", updated.Value)
	}
	if updated.Note == nil || *updated.Note != "calibrated scale" {
		t.Fatalf("updated weight note = %#v, want calibrated note", updated.Note)
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
		Note:             stringPointer("home cuff"),
		Source:           "manual",
		SourceRecordHash: "bp-a",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("create blood pressure: %v", err)
	}
	if bp.Note == nil || *bp.Note != "home cuff" {
		t.Fatalf("created blood pressure note = %#v", bp.Note)
	}
	bp, err = repo.UpdateBloodPressureEntry(ctx, health.UpdateBloodPressureEntryParams{
		ID:         bp.ID,
		RecordedAt: now.Add(time.Hour),
		Systolic:   119,
		Diastolic:  77,
		Note:       stringPointer("manual correction"),
		UpdatedAt:  now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update blood pressure: %v", err)
	}
	if bp.Systolic != 119 || bp.Pulse != nil {
		t.Fatalf("updated blood pressure = %#v, want systolic 119 with nil pulse", bp)
	}
	if bp.Note == nil || *bp.Note != "manual correction" {
		t.Fatalf("updated blood pressure note = %#v", bp.Note)
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
		Note:       stringPointer("started after annual exam"),
		Source:     "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("create medication: %v", err)
	}
	if med.Note == nil || *med.Note != "started after annual exam" {
		t.Fatalf("created medication note = %#v, want annual exam note", med.Note)
	}
	med, err = repo.UpdateMedicationCourse(ctx, health.UpdateMedicationCourseParams{
		ID:        med.ID,
		Name:      "Levothyroxine",
		StartDate: "2026-01-02",
		Note:      stringPointer("dose held before imaging"),
		UpdatedAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update medication: %v", err)
	}
	if med.StartDate != "2026-01-02" || med.EndDate != nil {
		t.Fatalf("updated medication = %#v, want new start date and nil end date", med)
	}
	if med.Note == nil || *med.Note != "dose held before imaging" {
		t.Fatalf("updated medication note = %#v, want held note", med.Note)
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
		Note:        stringPointer("labs look stable"),
		Source:      "manual",
		CreatedAt:   now,
		UpdatedAt:   now,
		Panels: []health.LabPanelWriteParams{
			{
				PanelName:    "Thyroid",
				DisplayOrder: 0,
				Results: []health.LabResultWriteParams{
					{
						TestName:      "Vitamin D",
						CanonicalSlug: analytePointer(health.AnalyteSlug("vitamin-d")),
						ValueText:     "32",
						ValueNumeric:  float64Pointer(32),
						Units:         stringPointer("ng/mL"),
						Notes:         []string{"HIV 4th gen narrative\nsecond line", "Hep C Ab reviewed"},
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
	if lab.Note == nil || *lab.Note != "labs look stable" {
		t.Fatalf("created lab note = %#v, want stable note", lab.Note)
	}
	if slug := lab.Panels[0].Results[0].CanonicalSlug; slug == nil || *slug != health.AnalyteSlug("vitamin-d") {
		t.Fatalf("created arbitrary lab slug = %#v, want vitamin-d", slug)
	}
	if !slices.Equal(lab.Panels[0].Results[0].Notes, []string{"HIV 4th gen narrative\nsecond line", "Hep C Ab reviewed"}) {
		t.Fatalf("created lab result notes = %#v", lab.Panels[0].Results[0].Notes)
	}
	lab, err = repo.UpdateLabCollection(ctx, health.UpdateLabCollectionParams{
		ID:          lab.ID,
		CollectedAt: now.Add(24 * time.Hour),
		Note:        stringPointer("glucose corrected"),
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
						Notes:         []string{"A1C context"},
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
	if lab.Note == nil || *lab.Note != "glucose corrected" {
		t.Fatalf("updated lab note = %#v, want corrected note", lab.Note)
	}
	results, err := repo.ListLabResultsWithCollection(ctx)
	if err != nil {
		t.Fatalf("list lab results with collection: %v", err)
	}
	if len(results) != 1 || results[0].TestName != "Glucose" {
		t.Fatalf("lab results = %#v, want only replacement result", results)
	}
	if !slices.Equal(results[0].Notes, []string{"A1C context"}) {
		t.Fatalf("lab result notes after replacement = %#v", results[0].Notes)
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

	bodyUnit := health.WeightUnitLb
	body, err := repo.CreateBodyCompositionEntry(ctx, health.CreateBodyCompositionEntryParams{
		RecordedAt:       now,
		BodyFatPercent:   float64Pointer(18.7),
		WeightValue:      float64Pointer(154.2),
		WeightUnit:       &bodyUnit,
		Method:           stringPointer("smart scale"),
		Note:             stringPointer("same row as weight import"),
		Source:           "manual",
		SourceRecordHash: "body-a",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("create body composition: %v", err)
	}
	if body.BodyFatPercent == nil || *body.BodyFatPercent != 18.7 || body.WeightUnit == nil || *body.WeightUnit != health.WeightUnitLb {
		t.Fatalf("created body composition = %#v", body)
	}
	body, err = repo.UpdateBodyCompositionEntry(ctx, health.UpdateBodyCompositionEntryParams{
		ID:             body.ID,
		RecordedAt:     now.Add(24 * time.Hour),
		BodyFatPercent: float64Pointer(18.1),
		Method:         stringPointer("dexa"),
		Note:           stringPointer("corrected method"),
		UpdatedAt:      now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update body composition: %v", err)
	}
	if body.BodyFatPercent == nil || *body.BodyFatPercent != 18.1 || body.WeightValue != nil || body.Method == nil || *body.Method != "dexa" {
		t.Fatalf("updated body composition = %#v", body)
	}
	if err := repo.DeleteBodyCompositionEntry(ctx, health.DeleteBodyCompositionEntryParams{
		ID:        body.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete body composition: %v", err)
	}
	bodies, err := repo.ListBodyCompositionEntries(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list body composition after delete: %v", err)
	}
	if len(bodies) != 0 {
		t.Fatalf("deleted body composition should be hidden, got %#v", bodies)
	}

	sleep, err := repo.CreateSleepEntry(ctx, health.CreateSleepEntryParams{
		RecordedAt:       now,
		QualityScore:     4,
		WakeupCount:      intPointer(2),
		Note:             stringPointer("woke up after storm"),
		Source:           "manual",
		SourceRecordHash: "sleep-a",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("create sleep: %v", err)
	}
	if sleep.QualityScore != 4 || sleep.WakeupCount == nil || *sleep.WakeupCount != 2 || sleep.Note == nil || *sleep.Note != "woke up after storm" {
		t.Fatalf("created sleep = %#v", sleep)
	}
	foundSleep, err := repo.FindManualSleepEntry(ctx, health.FindManualSleepEntryParams{RecordedAt: now})
	if err != nil {
		t.Fatalf("find manual sleep: %v", err)
	}
	if foundSleep == nil || foundSleep.ID != sleep.ID {
		t.Fatalf("found sleep = %#v, want id %d", foundSleep, sleep.ID)
	}
	_, err = repo.CreateSleepEntry(ctx, health.CreateSleepEntryParams{
		RecordedAt:       now,
		QualityScore:     3,
		Source:           "manual",
		SourceRecordHash: "sleep-duplicate",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	var conflictErr *health.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("duplicate sleep error = %v, want conflict", err)
	}
	sleep, err = repo.UpdateSleepEntry(ctx, health.UpdateSleepEntryParams{
		ID:           sleep.ID,
		QualityScore: intPointer(5),
		WakeupCount:  intPointer(0),
		UpdatedAt:    now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update sleep: %v", err)
	}
	if sleep.QualityScore != 5 || sleep.WakeupCount == nil || *sleep.WakeupCount != 0 || sleep.Note == nil || *sleep.Note != "woke up after storm" {
		t.Fatalf("updated sleep = %#v", sleep)
	}
	if err := repo.DeleteSleepEntry(ctx, health.DeleteSleepEntryParams{
		ID:        sleep.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete sleep: %v", err)
	}
	sleeps, err := repo.ListSleepEntries(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list sleep after delete: %v", err)
	}
	if len(sleeps) != 0 {
		t.Fatalf("deleted sleep should be hidden, got %#v", sleeps)
	}

	imaging, err := repo.CreateImagingRecord(ctx, health.CreateImagingRecordParams{
		PerformedAt:      now,
		Modality:         "X-ray",
		BodySite:         stringPointer("chest"),
		Title:            stringPointer("Chest X-ray"),
		Summary:          "No acute cardiopulmonary abnormality.",
		Impression:       stringPointer("Normal chest radiograph."),
		Note:             stringPointer("ordered for cough"),
		Notes:            []string{"XR TOE RIGHT narrative\nsecond line", "US Head/Neck findings"},
		Source:           "manual",
		SourceRecordHash: "imaging-a",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("create imaging: %v", err)
	}
	if imaging.BodySite == nil || *imaging.BodySite != "chest" || imaging.Note == nil || *imaging.Note != "ordered for cough" {
		t.Fatalf("created imaging = %#v", imaging)
	}
	if !slices.Equal(imaging.Notes, []string{"XR TOE RIGHT narrative\nsecond line", "US Head/Neck findings"}) {
		t.Fatalf("created imaging notes = %#v", imaging.Notes)
	}
	filtered, err := repo.ListImagingRecords(ctx, health.ImagingListParams{
		Modality: stringPointer("x-RAY"),
		BodySite: stringPointer("CHEST"),
	})
	if err != nil {
		t.Fatalf("list filtered imaging: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != imaging.ID {
		t.Fatalf("filtered imaging = %#v, want id %d", filtered, imaging.ID)
	}
	imaging, err = repo.UpdateImagingRecord(ctx, health.UpdateImagingRecordParams{
		ID:          imaging.ID,
		PerformedAt: now.Add(24 * time.Hour),
		Modality:    "CT",
		BodySite:    stringPointer("chest"),
		Summary:     "Stable small pulmonary nodule.",
		Note:        stringPointer("follow-up scan"),
		Notes:       []string{"US abdominal findings"},
		UpdatedAt:   now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("update imaging: %v", err)
	}
	if imaging.Modality != "CT" || imaging.Summary != "Stable small pulmonary nodule." || imaging.Title != nil {
		t.Fatalf("updated imaging = %#v", imaging)
	}
	if !slices.Equal(imaging.Notes, []string{"US abdominal findings"}) {
		t.Fatalf("updated imaging notes = %#v", imaging.Notes)
	}
	if err := repo.DeleteImagingRecord(ctx, health.DeleteImagingRecordParams{
		ID:        imaging.ID,
		DeletedAt: now.Add(2 * time.Hour),
		UpdatedAt: now.Add(2 * time.Hour),
	}); err != nil {
		t.Fatalf("delete imaging: %v", err)
	}
	images, err := repo.ListImagingRecords(ctx, health.ImagingListParams{})
	if err != nil {
		t.Fatalf("list imaging after delete: %v", err)
	}
	if len(images) != 0 {
		t.Fatalf("deleted imaging should be hidden, got %#v", images)
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
