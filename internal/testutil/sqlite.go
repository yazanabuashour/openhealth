package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/yazanabuashour/openhealth/internal/storage/sqlite"
)

func NewSQLiteDB(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := sqlite.Open(tb.TempDir() + "/openhealth.db")
	if err != nil {
		tb.Fatalf("open sqlite db: %v", err)
	}

	if err := sqlite.ApplyMigrations(context.Background(), db); err != nil {
		_ = db.Close()
		tb.Fatalf("apply migrations: %v", err)
	}

	tb.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func MustExec(tb testing.TB, db *sql.DB, statement string, args ...any) {
	tb.Helper()

	if _, err := db.ExecContext(context.Background(), statement, args...); err != nil {
		tb.Fatalf("exec %q failed: %v", statement, err)
	}
}
