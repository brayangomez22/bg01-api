// Package migrations embeds the goose SQL migrations so they ship inside the
// binary and run automatically on startup (no separate migrate step on deploy).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
