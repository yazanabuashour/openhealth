package health_test

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/yazanabuashour/openhealth/internal/health"
	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
	"github.com/yazanabuashour/openhealth/internal/testutil"
)

func TestServiceSummaryAndAnalyteTrend(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	}))

	testutil.MustExec(t, db, `
INSERT INTO health_weight_entry (recorded_at, value, unit, source, source_record_hash, note, created_at, updated_at)
VALUES
  (?, ?, 'lb', 'test', 'weight-a', NULL, ?, ?),
  (?, ?, 'lb', 'test', 'weight-b', NULL, ?, ?),
  (?, ?, 'lb', 'test', 'weight-c', NULL, ?, ?)
`,
		ts("2026-02-20T14:00:00Z"), 155.0, ts("2026-02-20T14:00:00Z"), ts("2026-02-20T14:00:00Z"),
		ts("2026-03-25T13:00:00Z"), 149.0, ts("2026-03-25T13:00:00Z"), ts("2026-03-25T13:00:00Z"),
		ts("2026-03-28T13:00:00Z"), 150.0, ts("2026-03-28T13:00:00Z"), ts("2026-03-28T13:00:00Z"),
	)
	testutil.MustExec(t, db, `
INSERT INTO health_blood_pressure_entry (recorded_at, systolic, diastolic, pulse, source, source_record_hash, created_at, updated_at)
VALUES (?, 110, 74, 65, 'test', 'bp-a', ?, ?)
`,
		ts("2026-03-28T14:00:00Z"), ts("2026-03-28T14:00:00Z"), ts("2026-03-28T14:00:00Z"),
	)
	testutil.MustExec(t, db, `
INSERT INTO health_sleep_entry (recorded_at, quality_score, wakeup_count, note, source, source_record_hash, created_at, updated_at)
VALUES (?, 4, 1, 'slept well', 'test', 'sleep-a', ?, ?)
`,
		ts("2026-03-29T00:00:00Z"), ts("2026-03-29T00:00:00Z"), ts("2026-03-29T00:00:00Z"),
	)
	testutil.MustExec(t, db, `
INSERT INTO health_medication_course (name, dosage_text, start_date, end_date, source, created_at, updated_at)
VALUES
  ('Levothyroxine', '25 mcg', '2025-05-01', NULL, 'test', ?, ?),
  ('Expired Med', NULL, '2025-01-01', '2026-01-15', 'test', ?, ?)
`,
		ts("2026-03-01T10:00:00Z"), ts("2026-03-01T10:00:00Z"),
		ts("2026-03-01T10:00:00Z"), ts("2026-03-01T10:00:00Z"),
	)

	collectionID := insertReturningID(t, db, `
INSERT INTO health_lab_collection (collected_at, source, created_at)
VALUES (?, 'test', ?)
RETURNING id
`, ts("2026-03-20T13:00:00Z"), ts("2026-03-20T13:00:00Z"))
	panelID := insertReturningID(t, db, `
INSERT INTO health_lab_panel (collection_id, panel_name, display_order)
VALUES (?, 'TSH', 0)
RETURNING id
`, collectionID)
	testutil.MustExec(t, db, `
INSERT INTO health_lab_result (panel_id, test_name, canonical_slug, value_text, value_numeric, units, range_text, flag, display_order)
VALUES
  (?, 'TSH', 'tsh', '4.90', 4.90, 'uIU/mL', '0.57-3.74', 'High', 0),
  (?, 'Glucose', 'glucose', '89', 89.00, 'mg/dL', '70-99', NULL, 1),
  (?, 'Vitamin D', 'vitamin-d', '32', 32.00, 'ng/mL', '30-100', NULL, 2),
  (?, 'Legacy Unsupported', NULL, '1', 1.00, 'mg/dL', '0-1', NULL, 3),
  (?, 'Legacy Flagged', NULL, '2', 2.00, 'mg/dL', '0-1', 'High', 4)
`,
		panelID, panelID, panelID, panelID, panelID,
	)

	summary, err := service.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.LatestWeight == nil || summary.LatestWeight.Value != 150 {
		t.Fatalf("latest weight = %#v", summary.LatestWeight)
	}
	if summary.Average7d == nil || *summary.Average7d != 149.5 {
		t.Fatalf("average7d = %#v, want 149.5", summary.Average7d)
	}
	if summary.Delta30d == nil || *summary.Delta30d != -5 {
		t.Fatalf("delta30d = %#v, want -5", summary.Delta30d)
	}
	if summary.LatestBloodPressure == nil || summary.LatestBloodPressure.Systolic != 110 {
		t.Fatalf("latest blood pressure = %#v", summary.LatestBloodPressure)
	}
	if summary.LatestSleep == nil || summary.LatestSleep.QualityScore != 4 || summary.LatestSleep.WakeupCount == nil || *summary.LatestSleep.WakeupCount != 1 {
		t.Fatalf("latest sleep = %#v", summary.LatestSleep)
	}
	if summary.ActiveMedicationCount != 1 {
		t.Fatalf("activeMedicationCount = %d, want 1", summary.ActiveMedicationCount)
	}
	if len(summary.LatestLabHighlights) != 4 {
		t.Fatalf("unexpected highlights: %#v", summary.LatestLabHighlights)
	}
	names := []string{
		summary.LatestLabHighlights[0].TestName,
		summary.LatestLabHighlights[1].TestName,
		summary.LatestLabHighlights[2].TestName,
		summary.LatestLabHighlights[3].TestName,
	}
	if names[0] != "TSH" || names[1] != "Glucose" || names[2] != "Vitamin D" || names[3] != "Legacy Flagged" {
		t.Fatalf("unexpected highlights: %#v", summary.LatestLabHighlights)
	}

	trend, err := service.AnalyteTrend(context.Background(), health.AnalyteSlug("Vitamin D"))
	if err != nil {
		t.Fatalf("analyte trend: %v", err)
	}
	if trend.Slug != health.AnalyteSlug("vitamin-d") || len(trend.Points) != 1 || trend.Latest == nil || trend.Latest.TestName != "Vitamin D" {
		t.Fatalf("unexpected analyte trend: %#v", trend)
	}
}

func TestServiceWeightWriteIdempotency(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	}))
	ctx := context.Background()
	recordedAt := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	sameDate := time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC)

	created, err := service.UpsertWeight(ctx, health.WeightRecordInput{
		RecordedAt: recordedAt,
		Value:      152.2,
		Unit:       health.WeightUnitLb,
		Note:       stringPointer(" morning scale "),
	})
	if err != nil {
		t.Fatalf("create weight through upsert: %v", err)
	}
	if created.Status != health.WeightWriteStatusCreated {
		t.Fatalf("created status = %q, want %q", created.Status, health.WeightWriteStatusCreated)
	}
	if created.Entry.Note == nil || *created.Entry.Note != "morning scale" {
		t.Fatalf("created weight note = %#v, want trimmed note", created.Entry.Note)
	}

	again, err := service.UpsertWeight(ctx, health.WeightRecordInput{
		RecordedAt: sameDate,
		Value:      152.2,
		Unit:       health.WeightUnitLb,
	})
	if err != nil {
		t.Fatalf("repeat weight through upsert: %v", err)
	}
	if again.Status != health.WeightWriteStatusAlreadyExists || again.Entry.ID != created.Entry.ID {
		t.Fatalf("repeat result = %#v, want already_exists for id %d", again, created.Entry.ID)
	}

	updated, err := service.UpsertWeight(ctx, health.WeightRecordInput{
		RecordedAt: sameDate,
		Value:      151.6,
		Unit:       health.WeightUnitLb,
		Note:       stringPointer("calibrated scale"),
	})
	if err != nil {
		t.Fatalf("update weight through upsert: %v", err)
	}
	if updated.Status != health.WeightWriteStatusUpdated || updated.Entry.ID != created.Entry.ID || updated.Entry.Value != 151.6 {
		t.Fatalf("updated result = %#v, want updated id %d value 151.6", updated, created.Entry.ID)
	}
	if updated.Entry.Note == nil || *updated.Entry.Note != "calibrated scale" {
		t.Fatalf("updated weight note = %#v, want calibrated note", updated.Entry.Note)
	}

	weights, err := service.ListWeight(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list weights: %v", err)
	}
	if len(weights) != 1 {
		t.Fatalf("weight count = %d, want 1", len(weights))
	}

	_, err = service.RecordWeight(ctx, health.WeightRecordInput{
		RecordedAt: sameDate,
		Value:      151.6,
		Unit:       health.WeightUnitLb,
	})
	var conflictErr *health.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("record duplicate error = %v, want conflict", err)
	}
}

func TestServiceNonWeightValidationAndNotFound(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	}))
	ctx := context.Background()

	_, err := service.RecordBloodPressure(ctx, health.BloodPressureRecordInput{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Systolic:   120,
		Diastolic:  0,
	})
	var validationErr *health.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("blood pressure validation error = %v, want validation", err)
	}

	_, err = service.RecordBloodPressure(ctx, health.BloodPressureRecordInput{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Systolic:   80,
		Diastolic:  80,
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("blood pressure relation validation error = %v, want validation", err)
	}

	validBP, err := service.RecordBloodPressure(ctx, health.BloodPressureRecordInput{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Systolic:   120,
		Diastolic:  80,
		Note:       stringPointer(" seated home cuff "),
	})
	if err != nil {
		t.Fatalf("record valid blood pressure: %v", err)
	}
	if validBP.Note == nil || *validBP.Note != "seated home cuff" {
		t.Fatalf("blood pressure note = %#v, want trimmed note", validBP.Note)
	}
	_, err = service.ReplaceBloodPressure(ctx, validBP.ID, health.BloodPressureRecordInput{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Systolic:   75,
		Diastolic:  80,
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("blood pressure replace relation validation error = %v, want validation", err)
	}

	_, err = service.CreateMedicationCourse(ctx, health.MedicationCourseInput{
		Name:      "Expired",
		StartDate: "2026-04-15",
		EndDate:   stringPointer("2026-04-14"),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("medication validation error = %v, want validation", err)
	}
	_, err = service.CreateMedicationCourse(ctx, health.MedicationCourseInput{
		Name:      "Blank Note",
		StartDate: "2026-04-15",
		Note:      stringPointer("  "),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("medication empty note validation error = %v, want validation", err)
	}

	_, err = service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Panels: []health.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results: []health.LabResultInput{
					{
						TestName:      "Invalid",
						CanonicalSlug: analytePointer(health.AnalyteSlug("bad/slug")),
						ValueText:     "1",
					},
				},
			},
		},
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("lab validation error = %v, want validation", err)
	}
	_, err = service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Note:        stringPointer(" "),
		Panels: []health.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results:   []health.LabResultInput{{TestName: "Glucose", ValueText: "89"}},
			},
		},
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("lab empty note validation error = %v, want validation", err)
	}
	_, err = service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Panels: []health.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results:   []health.LabResultInput{{TestName: "Glucose", ValueText: "89", Notes: []string{"valid", " "}}},
			},
		},
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("lab empty result note validation error = %v, want validation", err)
	}

	_, err = service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("body composition missing measurement validation error = %v, want validation", err)
	}
	_, err = service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:     time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		BodyFatPercent: float64Pointer(101),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("body composition body fat validation error = %v, want validation", err)
	}
	_, err = service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:  time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		WeightValue: float64Pointer(154),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("body composition weight pair validation error = %v, want validation", err)
	}
	kilograms := health.WeightUnit("kg")
	_, err = service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:  time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		WeightValue: float64Pointer(70),
		WeightUnit:  &kilograms,
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("body composition weight unit validation error = %v, want validation", err)
	}
	_, err = service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:     time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		BodyFatPercent: float64Pointer(18.7),
		Note:           stringPointer(" "),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("body composition empty note validation error = %v, want validation", err)
	}

	_, err = service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		QualityScore: 0,
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("sleep quality validation error = %v, want validation", err)
	}
	_, err = service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		QualityScore: 4,
		WakeupCount:  intPointer(-1),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("sleep wakeup validation error = %v, want validation", err)
	}
	_, err = service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		QualityScore: 4,
		Note:         stringPointer(" "),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("sleep empty note validation error = %v, want validation", err)
	}

	_, err = service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Modality:    " ",
		Summary:     "No acute abnormality.",
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("imaging modality validation error = %v, want validation", err)
	}
	_, err = service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Modality:    "X-ray",
		Summary:     " ",
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("imaging summary validation error = %v, want validation", err)
	}
	_, err = service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Modality:    "X-ray",
		Summary:     "No acute abnormality.",
		Note:        stringPointer(" "),
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("imaging empty note validation error = %v, want validation", err)
	}
	_, err = service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC),
		Modality:    "X-ray",
		Summary:     "No acute abnormality.",
		Notes:       []string{"finding", " "},
	})
	if !errors.As(err, &validationErr) {
		t.Fatalf("imaging empty notes validation error = %v, want validation", err)
	}

	err = service.DeleteLabCollection(ctx, 999)
	var notFoundErr *health.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("delete missing lab collection error = %v, want not found", err)
	}
}

func TestServiceImportContextLifecycles(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	}))
	ctx := context.Background()

	med, err := service.CreateMedicationCourse(ctx, health.MedicationCourseInput{
		Name:       "Semaglutide",
		DosageText: stringPointer("0.25 mg subcutaneous injection weekly"),
		StartDate:  "2026-04-01",
		Note:       stringPointer(" Started after insurance approval "),
	})
	if err != nil {
		t.Fatalf("create medication with note: %v", err)
	}
	if med.Note == nil || *med.Note != "Started after insurance approval" {
		t.Fatalf("medication note = %#v, want trimmed note", med.Note)
	}

	lab, err := service.CreateLabCollection(ctx, health.LabCollectionInput{
		CollectedAt: time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC),
		Note:        stringPointer(" labs look stable, keep moving "),
		Panels: []health.LabPanelInput{
			{
				PanelName: "Metabolic",
				Results:   []health.LabResultInput{{TestName: "Glucose", ValueText: "92", Notes: []string{" HIV 4th gen narrative\nsecond line ", "Hep C Ab reviewed"}}},
			},
		},
	})
	if err != nil {
		t.Fatalf("create lab with note: %v", err)
	}
	if lab.Note == nil || *lab.Note != "labs look stable, keep moving" {
		t.Fatalf("lab note = %#v, want trimmed note", lab.Note)
	}
	if !slices.Equal(lab.Panels[0].Results[0].Notes, []string{"HIV 4th gen narrative\nsecond line", "Hep C Ab reviewed"}) {
		t.Fatalf("lab result notes = %#v", lab.Panels[0].Results[0].Notes)
	}

	unit := health.WeightUnitLb
	body, err := service.CreateBodyComposition(ctx, health.BodyCompositionInput{
		RecordedAt:     time.Date(2026, 4, 12, 7, 0, 0, 0, time.UTC),
		BodyFatPercent: float64Pointer(18.7),
		WeightValue:    float64Pointer(154.2),
		WeightUnit:     &unit,
		Method:         stringPointer(" smart scale "),
		Note:           stringPointer(" same import row as scale weight "),
	})
	if err != nil {
		t.Fatalf("create body composition: %v", err)
	}
	if body.BodyFatPercent == nil || *body.BodyFatPercent != 18.7 || body.Method == nil || *body.Method != "smart scale" {
		t.Fatalf("created body composition = %#v", body)
	}
	body, err = service.ReplaceBodyComposition(ctx, body.ID, health.BodyCompositionInput{
		RecordedAt:     time.Date(2026, 4, 13, 7, 0, 0, 0, time.UTC),
		BodyFatPercent: float64Pointer(18.2),
		Note:           stringPointer("corrected scan"),
	})
	if err != nil {
		t.Fatalf("replace body composition: %v", err)
	}
	if body.BodyFatPercent == nil || *body.BodyFatPercent != 18.2 || body.WeightValue != nil || body.Note == nil || *body.Note != "corrected scan" {
		t.Fatalf("replaced body composition = %#v", body)
	}
	if err := service.DeleteBodyComposition(ctx, body.ID); err != nil {
		t.Fatalf("delete body composition: %v", err)
	}
	bodies, err := service.ListBodyComposition(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list body composition after delete: %v", err)
	}
	if len(bodies) != 0 {
		t.Fatalf("deleted body composition should be hidden, got %#v", bodies)
	}

	sleep, err := service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 12, 7, 0, 0, 0, time.UTC),
		QualityScore: 4,
		WakeupCount:  intPointer(2),
		Note:         stringPointer(" woke up after storm "),
	})
	if err != nil {
		t.Fatalf("create sleep: %v", err)
	}
	if sleep.Status != health.SleepWriteStatusCreated || sleep.Entry.Note == nil || *sleep.Entry.Note != "woke up after storm" {
		t.Fatalf("created sleep = %#v", sleep)
	}
	again, err := service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
		QualityScore: 4,
	})
	if err != nil {
		t.Fatalf("repeat sleep: %v", err)
	}
	if again.Status != health.SleepWriteStatusAlreadyExists || again.Entry.ID != sleep.Entry.ID {
		t.Fatalf("repeat sleep = %#v", again)
	}
	updated, err := service.UpsertSleep(ctx, health.SleepInput{
		RecordedAt:   time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC),
		QualityScore: 5,
		WakeupCount:  intPointer(0),
	})
	if err != nil {
		t.Fatalf("update sleep: %v", err)
	}
	if updated.Status != health.SleepWriteStatusUpdated || updated.Entry.QualityScore != 5 || updated.Entry.WakeupCount == nil || *updated.Entry.WakeupCount != 0 {
		t.Fatalf("updated sleep = %#v", updated)
	}
	if err := service.DeleteSleep(ctx, updated.Entry.ID); err != nil {
		t.Fatalf("delete sleep: %v", err)
	}
	sleeps, err := service.ListSleep(ctx, health.HistoryFilter{})
	if err != nil {
		t.Fatalf("list sleep after delete: %v", err)
	}
	if len(sleeps) != 0 {
		t.Fatalf("deleted sleep should be hidden, got %#v", sleeps)
	}

	imaging, err := service.CreateImaging(ctx, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC),
		Modality:    " X-ray ",
		BodySite:    stringPointer(" chest "),
		Title:       stringPointer(" Chest X-ray "),
		Summary:     " No acute cardiopulmonary abnormality. ",
		Impression:  stringPointer(" Normal chest radiograph. "),
		Note:        stringPointer(" ordered for cough "),
		Notes:       []string{" XR TOE RIGHT narrative\nsecond line ", "US Head/Neck findings"},
	})
	if err != nil {
		t.Fatalf("create imaging: %v", err)
	}
	if imaging.Modality != "X-ray" || imaging.Summary != "No acute cardiopulmonary abnormality." || imaging.Note == nil || *imaging.Note != "ordered for cough" {
		t.Fatalf("created imaging = %#v", imaging)
	}
	if !slices.Equal(imaging.Notes, []string{"XR TOE RIGHT narrative\nsecond line", "US Head/Neck findings"}) {
		t.Fatalf("created imaging notes = %#v", imaging.Notes)
	}
	filtered, err := service.ListImaging(ctx, health.ImagingListParams{
		Modality: stringPointer("x-RAY"),
		BodySite: stringPointer("CHEST"),
	})
	if err != nil {
		t.Fatalf("list filtered imaging: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != imaging.ID {
		t.Fatalf("filtered imaging = %#v, want id %d", filtered, imaging.ID)
	}
	imaging, err = service.ReplaceImaging(ctx, imaging.ID, health.ImagingRecordInput{
		PerformedAt: time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC),
		Modality:    "CT",
		BodySite:    stringPointer("chest"),
		Summary:     "Stable small pulmonary nodule.",
		Note:        stringPointer("follow-up scan"),
		Notes:       []string{"US abdominal findings"},
	})
	if err != nil {
		t.Fatalf("replace imaging: %v", err)
	}
	if imaging.Modality != "CT" || imaging.Title != nil || imaging.Note == nil || *imaging.Note != "follow-up scan" {
		t.Fatalf("replaced imaging = %#v", imaging)
	}
	if !slices.Equal(imaging.Notes, []string{"US abdominal findings"}) {
		t.Fatalf("replaced imaging notes = %#v", imaging.Notes)
	}
	if err := service.DeleteImaging(ctx, imaging.ID); err != nil {
		t.Fatalf("delete imaging: %v", err)
	}
	images, err := service.ListImaging(ctx, health.ImagingListParams{})
	if err != nil {
		t.Fatalf("list imaging after delete: %v", err)
	}
	if len(images) != 0 {
		t.Fatalf("deleted imaging should be hidden, got %#v", images)
	}
}

func TestServiceWeightTrendDefaultsAndMedicationFilter(t *testing.T) {
	t.Parallel()

	db := testutil.NewSQLiteDB(t)
	repo := sqlite.NewRepository(db)
	service := health.NewService(repo, health.WithClock(func() time.Time {
		return time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)
	}))

	testutil.MustExec(t, db, `
INSERT INTO health_weight_entry (recorded_at, value, unit, source, source_record_hash, note, created_at, updated_at)
VALUES
  (?, ?, 'lb', 'test', 'weight-old', NULL, ?, ?),
  (?, ?, 'lb', 'test', 'weight-may', NULL, ?, ?),
  (?, ?, 'lb', 'test', 'weight-june-a', NULL, ?, ?),
  (?, ?, 'lb', 'test', 'weight-june-b', NULL, ?, ?)
`,
		ts("2026-02-01T13:00:00Z"), 160.0, ts("2026-02-01T13:00:00Z"), ts("2026-02-01T13:00:00Z"),
		ts("2026-05-01T13:00:00Z"), 152.0, ts("2026-05-01T13:00:00Z"), ts("2026-05-01T13:00:00Z"),
		ts("2026-06-01T13:00:00Z"), 151.0, ts("2026-06-01T13:00:00Z"), ts("2026-06-01T13:00:00Z"),
		ts("2026-06-30T13:00:00Z"), 150.0, ts("2026-06-30T13:00:00Z"), ts("2026-06-30T13:00:00Z"),
	)
	testutil.MustExec(t, db, `
INSERT INTO health_medication_course (name, dosage_text, start_date, end_date, source, created_at, updated_at)
VALUES
  ('Active Med', NULL, '2026-01-01', NULL, 'test', ?, ?),
  ('Completed Med', NULL, '2026-01-01', '2026-02-01', 'test', ?, ?)
`,
		ts("2026-01-01T10:00:00Z"), ts("2026-01-01T10:00:00Z"),
		ts("2026-01-01T10:00:00Z"), ts("2026-01-01T10:00:00Z"),
	)

	trend, err := service.WeightTrend(context.Background(), health.WeightTrendParams{})
	if err != nil {
		t.Fatalf("weight trend: %v", err)
	}
	if trend.Range != health.WeightRange90d {
		t.Fatalf("default range = %q, want %q", trend.Range, health.WeightRange90d)
	}
	if len(trend.RawPoints) != 3 {
		t.Fatalf("default trend raw point count = %d, want 3", len(trend.RawPoints))
	}
	if trend.RawPoints[0].RecordedAt.After(trend.RawPoints[1].RecordedAt) {
		t.Fatalf("trend should be chronological: %#v", trend.RawPoints)
	}
	if len(trend.MonthlyAverageBuckets) != 2 || trend.MonthlyAverageBuckets[0].Month != "2026-05" || trend.MonthlyAverageBuckets[1].Month != "2026-06" {
		t.Fatalf("monthly buckets = %#v", trend.MonthlyAverageBuckets)
	}

	active, err := service.ListMedications(context.Background(), health.MedicationListParams{})
	if err != nil {
		t.Fatalf("list active medications: %v", err)
	}
	if len(active) != 1 || active[0].Name != "Active Med" {
		t.Fatalf("active medications = %#v", active)
	}

	all, err := service.ListMedications(context.Background(), health.MedicationListParams{
		Status: health.MedicationStatusAll,
	})
	if err != nil {
		t.Fatalf("list all medications: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("all medications count = %d, want 2", len(all))
	}
}

func insertReturningID(t *testing.T, db *sql.DB, statement string, args ...any) int {
	t.Helper()

	var id int
	if err := db.QueryRowContext(context.Background(), statement, args...).Scan(&id); err != nil {
		t.Fatalf("insert returning id failed: %v", err)
	}
	return id
}

func ts(value string) string {
	return value
}

func stringPointer(value string) *string {
	return &value
}

func float64Pointer(value float64) *float64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func analytePointer(value health.AnalyteSlug) *health.AnalyteSlug {
	return &value
}
