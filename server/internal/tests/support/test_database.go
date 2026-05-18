package support

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *sqlx.DB

func DatabaseSetup() func() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:18",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		slog.Error("Failed to start database test container.", "error", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		slog.Error("Failed to get database test container host.", "error", err)
		os.Exit(1)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		slog.Error("Failed to get database test container port.", "error", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	DB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error("Failed to connect to test database.", "error", err)
		os.Exit(1)
	}

	runMigrations()

	return func() {
		_ = DB.Close()
		_ = container.Terminate(ctx)
	}
}

func runMigrations() {
	driver, err := migratepostgres.WithInstance(DB.DB, &migratepostgres.Config{})
	if err != nil {
		slog.Error("Failed to create database migration driver.", "error", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash("../../migrations"),
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error("Failed to create database migration instance.", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Failed to run database migrations.", "error", err)
		os.Exit(1)
	}
}

func databaseReset(t *testing.T) {
	t.Helper()

	rows, err := DB.Query(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`)
	if err != nil {
		t.Fatalf("failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		tables = append(tables, `"`+name+`"`)
	}

	if len(tables) == 0 {
		return
	}

	query := fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))

	if _, err := DB.Exec(query); err != nil {
		t.Fatalf("database reset failed: %v", err)
	}
}
