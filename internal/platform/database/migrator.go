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

// splitStatements splits a migration file into individual statements on
// the ';' that terminates each one. It tracks two kinds of "don't split
// here" regions so SQL syntax inside plain content doesn't get mistaken
// for statement structure:
//
//   - single-quoted string literals, including the standard SQL escape
//     for a literal quote (two single quotes in a row) — needed because
//     seeded lesson content contains real sentences, e.g. "...a flaw in
//     you; it is a learned pattern...", with semicolons of their own.
//   - "--" line comments — needed because a comment can itself contain
//     an apostrophe (e.g. "-- the brief's learning framework"), which
//     would otherwise be misread as opening a string literal.
//
// Migration files in this project are plain DDL/DML with no stored
// procedures, block comments, or double-quoted/backtick identifiers
// containing semicolons, so this is sufficient without pulling in a
// full SQL parser.
func splitStatements(script string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	inComment := false

	runes := []rune(script)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		current.WriteRune(r)

		switch {
		case inComment:
			if r == '\n' {
				inComment = false
			}
		case inString && r == '\'' && i+1 < len(runes) && runes[i+1] == '\'':
			// Escaped quote (two single quotes) inside a string literal:
			// consume both runes as literal content, stay inString.
			current.WriteRune(runes[i+1])
			i++
		case r == '\'':
			inString = !inString
		case !inString && r == '-' && i+1 < len(runes) && runes[i+1] == '-':
			inComment = true
			current.WriteRune(runes[i+1])
			i++
		case r == ';' && !inString:
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		statements = append(statements, current.String())
	}

	return statements
}
