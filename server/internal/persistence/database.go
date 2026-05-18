package persistence

import (
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDatabase(databaseURL string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		slog.Error("Failed to connect to database.", "error", err)
		os.Exit(1)
	}
	return db
}
