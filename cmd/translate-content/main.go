// Command translate-content regenerates draft translations for one
// locale by calling the configured Translator (internal/notification/translate)
// against every level, lesson, quiz question, quiz option, and badge's
// canonical English text, then upserting the result into that entity's
// *_translations table.
//
// This is a content-pipeline tool, run by hand or in CI, not part of the
// running API — translating static lesson/quiz/badge content happens
// ahead of time, not on every request (see internal/notification/translate's
// package doc for why).
//
// Every row this tool writes still gets needs_review = true, regardless
// of TRANSLATE_MODE. Calling a real translation API instead of using an
// LLM-drafted seed does not make a row reviewed — it only changes which
// machine produced the draft. See README.md's Localization section.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/notification/translate"
	"github.com/traumacompendium/backend/internal/platform/database"
	"github.com/traumacompendium/backend/internal/platform/logger"
)

func main() {
	target := flag.String("target", "", "target locale to translate into (one of: pcm, yo, ha, ig, fr, sw, ar)")
	force := flag.Bool("force", false, "re-translate every row for the target locale, even if it already has one (default: only rows missing a translation)")
	flag.Parse()

	if err := run(*target, *force); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run(target string, force bool) error {
	if target == "" || target == learning.DefaultLocale || !learning.IsSupportedLocale(target) {
		return fmt.Errorf("translate-content: --target must be one of %v (got %q)", learning.SupportedLocales[1:], target)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db, err := database.Open(ctx, cfg.Database, log)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close() //nolint:errcheck // best-effort on process exit

	tr := translate.New(cfg.Translate, log)

	log.Info().
		Str("target_locale", target).
		Bool("force", force).
		Str("translate_mode", string(cfg.Translate.Mode)).
		Msg("translate-content: starting")

	levels, err := translateLevels(ctx, db, tr, target, force)
	if err != nil {
		return fmt.Errorf("translate levels: %w", err)
	}
	lessons, err := translateLessons(ctx, db, tr, target, force)
	if err != nil {
		return fmt.Errorf("translate lessons: %w", err)
	}
	questions, err := translateQuizQuestions(ctx, db, tr, target, force)
	if err != nil {
		return fmt.Errorf("translate quiz questions: %w", err)
	}
	options, err := translateQuizOptions(ctx, db, tr, target, force)
	if err != nil {
		return fmt.Errorf("translate quiz options: %w", err)
	}
	badges, err := translateBadges(ctx, db, tr, target, force)
	if err != nil {
		return fmt.Errorf("translate badges: %w", err)
	}

	log.Info().
		Str("target_locale", target).
		Int("levels", levels).
		Int("lessons", lessons).
		Int("quiz_questions", questions).
		Int("quiz_options", options).
		Int("badges", badges).
		Msg("translate-content: done — every written row is needs_review=true; this is a draft, not a substitute for human review")

	return nil
}

// selectCandidates runs allQuery (force) or missingQuery bound to target
// (only rows without an existing translation for it) and returns the
// resulting rows.
func selectCandidates(ctx context.Context, db *sql.DB, force bool, target, allQuery, missingQuery string) (*sql.Rows, error) {
	if force {
		return db.QueryContext(ctx, allQuery)
	}
	return db.QueryContext(ctx, missingQuery, target)
}

// translateText is a thin wrapper so every call site reads the same way:
// translate from canonical English into target.
func translateText(ctx context.Context, tr translate.Translator, text, target string) (string, error) {
	result, err := tr.Translate(ctx, translate.Request{
		Text:         text,
		SourceLocale: learning.DefaultLocale,
		TargetLocale: target,
	})
	if err != nil {
		return "", err
	}
	return result.TranslatedText, nil
}

func translateLevels(ctx context.Context, db *sql.DB, tr translate.Translator, target string, force bool) (int, error) {
	rows, err := selectCandidates(ctx, db, force, target,
		"SELECT id, name, description FROM levels",
		"SELECT id, name, description FROM levels WHERE id NOT IN (SELECT level_id FROM level_translations WHERE locale = ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("query levels: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query, nothing to clean up on error

	type levelRow struct{ id, name, description string }
	var candidates []levelRow
	for rows.Next() {
		var r levelRow
		if err := rows.Scan(&r.id, &r.name, &r.description); err != nil {
			return 0, fmt.Errorf("scan level row: %w", err)
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate level rows: %w", err)
	}

	for _, r := range candidates {
		name, err := translateText(ctx, tr, r.name, target)
		if err != nil {
			return 0, fmt.Errorf("translate level %s name: %w", r.id, err)
		}
		description, err := translateText(ctx, tr, r.description, target)
		if err != nil {
			return 0, fmt.Errorf("translate level %s description: %w", r.id, err)
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO level_translations (level_id, locale, name, description, needs_review)
			VALUES (?, ?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description), needs_review = TRUE`,
			r.id, target, name, description,
		); err != nil {
			return 0, fmt.Errorf("upsert level %s translation: %w", r.id, err)
		}
	}

	return len(candidates), nil
}

func translateLessons(ctx context.Context, db *sql.DB, tr translate.Translator, target string, force bool) (int, error) {
	rows, err := selectCandidates(ctx, db, force, target,
		"SELECT id, title, content FROM lessons",
		"SELECT id, title, content FROM lessons WHERE id NOT IN (SELECT lesson_id FROM lesson_translations WHERE locale = ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("query lessons: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query, nothing to clean up on error

	type lessonRow struct{ id, title, content string }
	var candidates []lessonRow
	for rows.Next() {
		var r lessonRow
		if err := rows.Scan(&r.id, &r.title, &r.content); err != nil {
			return 0, fmt.Errorf("scan lesson row: %w", err)
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate lesson rows: %w", err)
	}

	for _, r := range candidates {
		title, err := translateText(ctx, tr, r.title, target)
		if err != nil {
			return 0, fmt.Errorf("translate lesson %s title: %w", r.id, err)
		}
		content, err := translateText(ctx, tr, r.content, target)
		if err != nil {
			return 0, fmt.Errorf("translate lesson %s content: %w", r.id, err)
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO lesson_translations (lesson_id, locale, title, content, needs_review)
			VALUES (?, ?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE title = VALUES(title), content = VALUES(content), needs_review = TRUE`,
			r.id, target, title, content,
		); err != nil {
			return 0, fmt.Errorf("upsert lesson %s translation: %w", r.id, err)
		}
	}

	return len(candidates), nil
}

func translateQuizQuestions(ctx context.Context, db *sql.DB, tr translate.Translator, target string, force bool) (int, error) {
	rows, err := selectCandidates(ctx, db, force, target,
		"SELECT id, question FROM quiz_questions",
		"SELECT id, question FROM quiz_questions WHERE id NOT IN (SELECT question_id FROM quiz_question_translations WHERE locale = ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("query quiz questions: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query, nothing to clean up on error

	type questionRow struct{ id, text string }
	var candidates []questionRow
	for rows.Next() {
		var r questionRow
		if err := rows.Scan(&r.id, &r.text); err != nil {
			return 0, fmt.Errorf("scan quiz question row: %w", err)
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate quiz question rows: %w", err)
	}

	for _, r := range candidates {
		text, err := translateText(ctx, tr, r.text, target)
		if err != nil {
			return 0, fmt.Errorf("translate quiz question %s: %w", r.id, err)
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO quiz_question_translations (question_id, locale, text, needs_review)
			VALUES (?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE text = VALUES(text), needs_review = TRUE`,
			r.id, target, text,
		); err != nil {
			return 0, fmt.Errorf("upsert quiz question %s translation: %w", r.id, err)
		}
	}

	return len(candidates), nil
}

func translateQuizOptions(ctx context.Context, db *sql.DB, tr translate.Translator, target string, force bool) (int, error) {
	rows, err := selectCandidates(ctx, db, force, target,
		"SELECT id, text FROM quiz_options",
		"SELECT id, text FROM quiz_options WHERE id NOT IN (SELECT option_id FROM quiz_option_translations WHERE locale = ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("query quiz options: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query, nothing to clean up on error

	type optionRow struct{ id, text string }
	var candidates []optionRow
	for rows.Next() {
		var r optionRow
		if err := rows.Scan(&r.id, &r.text); err != nil {
			return 0, fmt.Errorf("scan quiz option row: %w", err)
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate quiz option rows: %w", err)
	}

	for _, r := range candidates {
		text, err := translateText(ctx, tr, r.text, target)
		if err != nil {
			return 0, fmt.Errorf("translate quiz option %s: %w", r.id, err)
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO quiz_option_translations (option_id, locale, text, needs_review)
			VALUES (?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE text = VALUES(text), needs_review = TRUE`,
			r.id, target, text,
		); err != nil {
			return 0, fmt.Errorf("upsert quiz option %s translation: %w", r.id, err)
		}
	}

	return len(candidates), nil
}

func translateBadges(ctx context.Context, db *sql.DB, tr translate.Translator, target string, force bool) (int, error) {
	rows, err := selectCandidates(ctx, db, force, target,
		"SELECT id, name, description FROM badges",
		"SELECT id, name, description FROM badges WHERE id NOT IN (SELECT badge_id FROM badge_translations WHERE locale = ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("query badges: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query, nothing to clean up on error

	type badgeRow struct{ id, name, description string }
	var candidates []badgeRow
	for rows.Next() {
		var r badgeRow
		if err := rows.Scan(&r.id, &r.name, &r.description); err != nil {
			return 0, fmt.Errorf("scan badge row: %w", err)
		}
		candidates = append(candidates, r)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate badge rows: %w", err)
	}

	for _, r := range candidates {
		name, err := translateText(ctx, tr, r.name, target)
		if err != nil {
			return 0, fmt.Errorf("translate badge %s name: %w", r.id, err)
		}
		description, err := translateText(ctx, tr, r.description, target)
		if err != nil {
			return 0, fmt.Errorf("translate badge %s description: %w", r.id, err)
		}

		if _, err := db.ExecContext(ctx, `
			INSERT INTO badge_translations (badge_id, locale, name, description, needs_review)
			VALUES (?, ?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description), needs_review = TRUE`,
			r.id, target, name, description,
		); err != nil {
			return 0, fmt.Errorf("upsert badge %s translation: %w", r.id, err)
		}
	}

	return len(candidates), nil
}
