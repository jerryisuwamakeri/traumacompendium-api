package mysql

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/platform/database"
	"github.com/traumacompendium/backend/migrations"
)

// openTestDB connects to a real MySQL instance for repository integration
// tests. It is skipped unless TEST_DATABASE_DSN is set (wired up in CI
// against a MySQL service container) so `go test ./...` still passes for
// anyone running it locally without a database.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping MySQL repository integration test")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	var migFS fs.FS = migrations.FS
	if err := database.Migrate(context.Background(), db, migFS, zerolog.Nop()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() { truncateAll(t, db) })

	return db
}

func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE refresh_tokens",
		"TRUNCATE TABLE otp_codes",
		"TRUNCATE TABLE users",
		"SET FOREIGN_KEY_CHECKS = 1",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("truncate: %s: %v", stmt, err)
		}
	}
}
