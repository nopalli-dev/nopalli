// Package db holds the cross-cutting database infrastructure: the connection
// pool, the unit-of-work transaction boundary, and the schema migrator. Module
// repositories build on these (rdt: core.infrastructure.db).
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the slice of pgx that repositories use. Both *pgxpool.Pool and pgx.Tx
// satisfy it, so a repository works either directly against the pool (reads) or
// inside a transaction (writes driven by a UnitOfWork).
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Connect opens a pgx connection pool and verifies it with a ping.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return pool, nil
}

// UnitOfWork runs a function inside a single transaction, committing on success
// and rolling back on error or panic (rdt: SqlAlchemyUnitOfWork — the
// transaction boundary that the repositories built for one operation share).
type UnitOfWork struct {
	pool *pgxpool.Pool
}

// NewUnitOfWork builds a UnitOfWork over a pool.
func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

// Do opens a transaction, invokes fn with it, and commits — unless fn returns
// an error or panics, in which case the transaction is rolled back. The tx is
// passed as a DBTX so repositories created inside fn share it.
func (u *UnitOfWork) Do(ctx context.Context, fn func(tx DBTX) error) (err error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("db: commit: %w", err)
	}
	return nil
}
