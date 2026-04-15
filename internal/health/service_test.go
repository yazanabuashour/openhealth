package health_test

import (
	"context"
	"database/sql"
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
  (?, 'Legacy Unsupported', NULL, '1', 1.00, 'mg/dL', '0-1', NULL, 2),
  (?, 'Legacy Flagged', NULL, '2', 2.00, 'mg/dL', '0-1', 'High', 3)
`,
		panelID, panelID, panelID, panelID,
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
	if summary.ActiveMedicationCount != 1 {
		t.Fatalf("activeMedicationCount = %d, want 1", summary.ActiveMedicationCount)
	}
	if names := []string{
		summary.LatestLabHighlights[0].TestName,
		summary.LatestLabHighlights[1].TestName,
		summary.LatestLabHighlights[2].TestName,
	}; len(summary.LatestLabHighlights) != 3 || names[0] != "TSH" || names[1] != "Glucose" || names[2] != "Legacy Flagged" {
		t.Fatalf("unexpected highlights: %#v", summary.LatestLabHighlights)
	}

	trend, err := service.AnalyteTrend(context.Background(), health.AnalyteSlugTSH)
	if err != nil {
		t.Fatalf("analyte trend: %v", err)
	}
	if len(trend.Points) != 1 || trend.Latest == nil || trend.Latest.TestName != "TSH" {
		t.Fatalf("unexpected analyte trend: %#v", trend)
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
