package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	Name string
	SQL  string
}

var legacyMigrationObjects = map[string]map[string][]string{
	"0001_health_schema.sql": {
		"table": {
			"health_weight_entry",
			"health_blood_pressure_entry",
			"health_medication_course",
			"health_lab_collection",
			"health_lab_panel",
			"health_lab_result",
		},
		"index": {
			"idx_health_weight_entry_recorded_at_desc",
			"idx_health_blood_pressure_entry_recorded_at_desc",
			"idx_health_medication_course_start_date_desc",
			"idx_health_lab_collection_collected_at_desc",
			"idx_health_lab_result_canonical_slug_panel_id",
			"idx_health_lab_result_panel_id_display_order",
		},
	},
}

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	available, err := loadMigrations()
	if err != nil {
		return err
	}

	exists, err := migrationTableExists(ctx, db)
	if err != nil {
		return err
	}
	if !exists {
		inferred, err := inferredLegacyMigrationNames(ctx, db, available)
		if err != nil {
			return err
		}
		if err := ensureMigrationTable(ctx, db); err != nil {
			return err
		}
		if err := recordAppliedMigrationNames(ctx, db, inferred); err != nil {
			return err
		}
	}

	pending, err := pendingMigrations(ctx, db, available)
	if err != nil {
		return err
	}

	for _, next := range pending {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, next.SQL); err != nil {
			_ = tx.Rollback()
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)`,
			next.Name,
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func PendingMigrations(ctx context.Context, db *sql.DB) ([]migration, error) {
	available, err := loadMigrations()
	if err != nil {
		return nil, err
	}
	return pendingMigrations(ctx, db, available)
}

func pendingMigrations(ctx context.Context, db *sql.DB, available []migration) ([]migration, error) {
	exists, err := migrationTableExists(ctx, db)
	if err != nil {
		return nil, err
	}
	if !exists {
		inferred, err := inferredLegacyMigrationNames(ctx, db, available)
		if err != nil {
			return nil, err
		}
		return filterPendingMigrations(available, inferred), nil
	}

	applied, err := appliedMigrationNames(ctx, db)
	if err != nil {
		return nil, err
	}
	return filterPendingMigrations(available, applied), nil
}

func filterPendingMigrations(available []migration, applied map[string]struct{}) []migration {
	pending := make([]migration, 0, len(available))
	for _, next := range available {
		if _, ok := applied[next.Name]; !ok {
			pending = append(pending, next)
		}
	}
	return pending
}

func EnsureCurrent(ctx context.Context, db *sql.DB) error {
	pending, err := PendingMigrations(ctx, db)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	return fmt.Errorf("database has pending migrations; run `openhealth migrate` before serving")
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  name TEXT PRIMARY KEY,
  applied_at TEXT NOT NULL
)`)
	return err
}

func migrationTableExists(ctx context.Context, db *sql.DB) (bool, error) {
	row := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func appliedMigrationNames(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	names := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

func inferredLegacyMigrationNames(ctx context.Context, db *sql.DB, available []migration) (map[string]struct{}, error) {
	inferred := make(map[string]struct{})
	for _, next := range available {
		objectSets, ok := legacyMigrationObjects[next.Name]
		if !ok {
			continue
		}

		present, partial, err := migrationObjectsPresent(ctx, db, objectSets)
		if err != nil {
			return nil, err
		}
		if partial {
			return nil, fmt.Errorf(
				"database contains a partial pre-migration schema for %s; reconcile the existing objects before running migrations",
				next.Name,
			)
		}
		if present {
			inferred[next.Name] = struct{}{}
		}
	}
	return inferred, nil
}

func migrationObjectsPresent(ctx context.Context, db *sql.DB, objectSets map[string][]string) (bool, bool, error) {
	allPresent := true
	anyPresent := false
	for objectType, names := range objectSets {
		presentCount, err := countSchemaObjects(ctx, db, objectType, names)
		if err != nil {
			return false, false, err
		}
		switch {
		case presentCount == 0:
			allPresent = false
		case presentCount == len(names):
			anyPresent = true
		default:
			return false, true, nil
		}
	}
	if allPresent && anyPresent {
		return true, false, nil
	}
	if anyPresent {
		return false, true, nil
	}
	return false, false, nil
}

func countSchemaObjects(ctx context.Context, db *sql.DB, objectType string, names []string) (int, error) {
	count := 0
	for _, name := range names {
		row := db.QueryRowContext(
			ctx,
			`SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?`,
			objectType,
			name,
		)
		var present int
		if err := row.Scan(&present); err != nil {
			return 0, err
		}
		if present > 0 {
			count++
		}
	}
	return count, nil
}

func recordAppliedMigrationNames(ctx context.Context, db *sql.DB, names map[string]struct{}) error {
	if len(names) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	appliedAt := time.Now().UTC().Format(time.RFC3339Nano)
	for name := range names {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT OR IGNORE INTO schema_migrations (name, applied_at) VALUES (?, ?)`,
			name,
			appliedAt,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		body, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration{
			Name: entry.Name(),
			SQL:  string(body),
		})
	}

	slices.SortFunc(migrations, func(left, right migration) int {
		switch {
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		default:
			return 0
		}
	})

	return migrations, nil
}
