package sqlite

import (
	"database/sql"

	"github.com/yazanabuashour/openhealth/internal/storage/sqlite/sqlc"
)

const sqliteConstraintUnique = 2067

type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:      db,
		queries: sqlc.New(db),
	}
}
