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

	pending, err := PendingMigrations(ctx, db)
	if err != nil {
		t.Fatalf("pending migrations: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected legacy schema to baseline current migration, got %d pending", len(pending))
	}

	if err := EnsureCurrent(ctx, db); err != nil {
		t.Fatalf("ensure current: %v", err)
	}
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	applied, err := appliedMigrationNames(ctx, db)
	if err != nil {
		t.Fatalf("applied migration names: %v", err)
	}
	if _, ok := applied[migrations[0].Name]; !ok {
		t.Fatalf("expected %s to be recorded in schema_migrations", migrations[0].Name)
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
