package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func TestPendingAndApplyMigrationsBaselineLegacySchema(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one migration")
	}
	if _, err := db.ExecContext(ctx, migrations[0].SQL); err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}
	collectionID := insertTestRowID(t, db, `
INSERT INTO health_lab_collection (collected_at, source, created_at)
VALUES ('2026-04-01T00:00:00Z', 'test', '2026-04-01T00:00:00Z')
RETURNING id
`)
	panelID := insertTestRowID(t, db, `
INSERT INTO health_lab_panel (collection_id, panel_name, display_order)
VALUES (?, 'Metabolic', 0)
RETURNING id
`, collectionID)
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_lab_result (panel_id, test_name, canonical_slug, value_text, display_order)
VALUES (?, 'Glucose', 'glucose', '89', 0)
`, panelID); err != nil {
		t.Fatalf("seed legacy lab result: %v", err)
	}

	pending, err := PendingMigrations(ctx, db)
	if err != nil {
		t.Fatalf("pending migrations: %v", err)
	}
	if len(pending) != len(migrations)-1 {
		t.Fatalf("expected legacy schema to baseline first migration, got %d pending of %d migrations", len(pending), len(migrations))
	}

	if err := EnsureCurrent(ctx, db); err == nil {
		t.Fatal("expected legacy schema with unapplied follow-up migrations to be out of date")
	}
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	applied, err := appliedMigrationNames(ctx, db)
	if err != nil {
		t.Fatalf("applied migration names: %v", err)
	}
	for _, migration := range migrations {
		if _, ok := applied[migration.Name]; !ok {
			t.Fatalf("expected %s to be recorded in schema_migrations", migration.Name)
		}
	}

	var preservedSlug string
	if err := db.QueryRowContext(ctx, `
SELECT canonical_slug
FROM health_lab_result
WHERE test_name = 'Glucose'
`).Scan(&preservedSlug); err != nil {
		t.Fatalf("query preserved lab result: %v", err)
	}
	if preservedSlug != "glucose" {
		t.Fatalf("preserved slug = %q, want glucose", preservedSlug)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_lab_result (panel_id, test_name, canonical_slug, value_text, display_order)
VALUES (?, 'Vitamin D', 'vitamin-d', '32', 1)
`, panelID); err != nil {
		t.Fatalf("insert arbitrary lab slug after migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
UPDATE health_lab_collection
SET note = 'labs look stable'
WHERE id = ?
`, collectionID); err != nil {
		t.Fatalf("update migrated lab collection note: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_medication_course (name, dosage_text, start_date, note, source, created_at, updated_at)
VALUES ('Semaglutide', '0.25 mg subcutaneous injection weekly', '2026-04-01', 'coverage approved', 'test', '2026-04-01T00:00:00Z', '2026-04-01T00:00:00Z')
`); err != nil {
		t.Fatalf("insert migrated medication note: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_body_composition_entry (
  recorded_at, body_fat_percent, weight_value, weight_unit, method, note, source, source_record_hash, created_at, updated_at
)
VALUES (
  '2026-04-01T00:00:00Z', 18.7, 154.2, 'lb', 'smart scale', 'same import row as weight', 'test', 'body-a', '2026-04-01T00:00:00Z', '2026-04-01T00:00:00Z'
)
`); err != nil {
		t.Fatalf("insert migrated body composition: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_sleep_entry (
  recorded_at, quality_score, wakeup_count, note, source, source_record_hash, created_at, updated_at
)
VALUES (
  '2026-04-01T00:00:00Z', 4, 2, 'woke up after storm', 'test', 'sleep-a', '2026-04-01T00:00:00Z', '2026-04-01T00:00:00Z'
)
`); err != nil {
		t.Fatalf("insert migrated sleep: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO health_imaging_record (
  performed_at, modality, body_site, title, summary, impression, note, source, source_record_hash, created_at, updated_at
)
VALUES (
  '2026-04-01T00:00:00Z', 'X-ray', 'chest', 'Chest X-ray', 'No acute cardiopulmonary abnormality.', 'Normal chest radiograph.', 'ordered for cough', 'test', 'imaging-a', '2026-04-01T00:00:00Z', '2026-04-01T00:00:00Z'
)
`); err != nil {
		t.Fatalf("insert migrated imaging: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO openhealth_config (key, value_json, updated_at)
VALUES ('runner.default_limit', '{"value":10}', '2026-04-01T00:00:00Z')
`); err != nil {
		t.Fatalf("insert migrated config value: %v", err)
	}
}

func TestPendingMigrationsRejectsPartialLegacySchema(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `
CREATE TABLE health_weight_entry (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recorded_at TEXT NOT NULL,
  value REAL NOT NULL,
  unit TEXT NOT NULL,
  source TEXT NOT NULL,
  source_record_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
)`); err != nil {
		t.Fatalf("seed partial legacy schema: %v", err)
	}

	_, err := PendingMigrations(ctx, db)
	if err == nil {
		t.Fatal("expected partial legacy schema to fail inference")
	}
	if !strings.Contains(err.Error(), "partial pre-migration schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(t.TempDir() + "/openhealth.db")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func insertTestRowID(t *testing.T, db *sql.DB, statement string, args ...any) int {
	t.Helper()

	var id int
	if err := db.QueryRowContext(context.Background(), statement, args...).Scan(&id); err != nil {
		t.Fatalf("insert test row: %v", err)
	}
	return id
}
