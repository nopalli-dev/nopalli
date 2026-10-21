// Package migrations embeds nopalli's SQL migrations so they can be applied
// from Go (tests, a migrate command) without shipping the .sql files
// separately. nopalli owns its schema and migrations (it does not share rdt's
// database). The files are goose-format (-- +goose Up / Down).
package migrations

import "embed"

// FS holds the migration files, rooted at this package directory.
//
//go:embed *.sql
var FS embed.FS
