// Package migrations embeds the project's versioned SQL migration files so
// the compiled binary carries its own schema history — no separate asset
// deployment step is required.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
