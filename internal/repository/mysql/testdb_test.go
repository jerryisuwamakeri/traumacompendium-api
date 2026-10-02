package mysql

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/user"
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

// truncateAll clears per-user state between tests. It deliberately does
// NOT truncate the learning content tables (levels, lessons, quizzes,
// quiz_questions, quiz_options, badges): that data is reference content
// seeded once by migration 0014, analogous to a lookup table, and
// learning repository tests rely on it being present and stable across
// the whole test run rather than re-seeding it themselves.
func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE refresh_tokens",
		"TRUNCATE TABLE otp_codes",
		"TRUNCATE TABLE journal_entries",
		"TRUNCATE TABLE mood_entries",
		"TRUNCATE TABLE user_lesson_progress",
		"TRUNCATE TABLE user_badges",
		"TRUNCATE TABLE safereach_location_updates",
		"TRUNCATE TABLE safereach_alerts",
		"TRUNCATE TABLE trusted_contacts",
		"TRUNCATE TABLE device_tokens",
		"TRUNCATE TABLE community_reports",
		"TRUNCATE TABLE community_comments",
		"TRUNCATE TABLE community_posts",
		"TRUNCATE TABLE users",
		"SET FOREIGN_KEY_CHECKS = 1",
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("truncate: %s: %v", stmt, err)
		}
	}
}

// seedTestUser inserts a minimal, valid user row (via UserRepository, so
// it stays in sync with that table's constraints) and returns its ID —
// used by other repositories' tests whose tables carry a foreign key to
// users but have no reason to exercise UserRepository themselves.
func seedTestUser(t *testing.T, db *sql.DB) string {
	t.Helper()

	id := uuid.NewString()
	suffix := id[:8]
	now := time.Now().UTC().Truncate(time.Second)

	u := &user.User{
		ID:           id,
		FullName:     "Test User",
		Email:        "test+" + suffix + "@example.com",
		Phone:        "+23480" + suffix,
		PasswordHash: "bcrypt-hash-placeholder",
		Role:         user.RoleUser,
		Status:       user.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := NewUserRepository(db).Create(context.Background(), u); err != nil {
		t.Fatalf("seedTestUser: create: %v", err)
	}

	return id
}
