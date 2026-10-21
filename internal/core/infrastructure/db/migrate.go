package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/nopalli-io/nopalli/internal/migrations"
)

// Migrate applies all pending migrations to the database at dsn (goose over the
// embedded migration FS). It opens its own database/sql connection because
// goose is built on database/sql; the app itself uses the pgx pool.
func Migrate(ctx context.Context, dsn string) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("db: open for migrate: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("db: goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("db: migrate up: %w", err)
	}
	return nil
}
