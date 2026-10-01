package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/rs/zerolog"
)

// migrationsTable tracks which migration files have already been applied,
// so Migrate is safe to run on every startup (idempotent, ordered).
const migrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version     VARCHAR(255) NOT NULL PRIMARY KEY,
	applied_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

// Migrate applies every "*.up.sql" file in migrationsFS that has not yet
// been recorded in schema_migrations, in lexical filename order (hence the
// numeric prefix convention: 0001_, 0002_, ...). Each migration runs in its
// own transaction; a failure stops the run and leaves already-applied
// migrations intact so the operator can fix and retry.
//
// This intentionally hand-rolls what a library like golang-migrate would
// otherwise provide, to keep the dependency footprint minimal.
func Migrate(ctx context.Context, db *sql.DB, migrationsFS fs.FS, log zerolog.Logger) error {
	if _, err := db.ExecContext(ctx, migrationsTable); err != nil {
		return fmt.Errorf("migrate: ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("migrate: load applied versions: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("migrate: read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}

		if err := applyMigration(ctx, db, version, string(content)); err != nil {
			return fmt.Errorf("migrate: apply %s: %w", name, err)
		}

		log.Info().Str("version", version).Msg("migrate: applied")
	}

	return nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, db *sql.DB, version, sqlScript string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // rollback is a no-op after a successful commit

	for _, stmt := range splitStatements(sqlScript) {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec statement: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	return tx.Commit()
}

// splitStatements performs a simple semicolon split. Migration files in
// this project are plain DDL/DML with no stored procedures, so this is
// sufficient without pulling in a full SQL parser.
func splitStatements(script string) []string {
	return strings.Split(script, ";")
}
