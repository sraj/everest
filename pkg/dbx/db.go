package dbx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DBConfig holds connection-pool settings passed to New.
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB is the unified entry point: it wraps sqlx for execution and
// exposes squirrel-backed builder methods. Import only this package.
type DB struct {
	db   *sqlx.DB
	psql sq.StatementBuilderType
}

// New connects to PostgreSQL using cfg and returns a ready *DB.
func New(cfg DBConfig) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("dbx.New: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	return wrap(db), nil
}

// Wrap creates a *DB from an existing *sqlx.DB. Useful in tests.
func Wrap(db *sqlx.DB) *DB { return wrap(db) }

func wrap(db *sqlx.DB) *DB {
	return &DB{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// Close closes the underlying connection pool.
func (d *DB) Close() error { return d.db.Close() }

// Ping verifies the connection is alive.
func (d *DB) Ping(ctx context.Context) error { return d.db.PingContext(ctx) }

// Underlying returns the raw *sqlx.DB for escape-hatch use (e.g. named queries).
func (d *DB) Underlying() *sqlx.DB { return d.db }

// ── Builder entry points ─────────────────────────────────────

// Select starts a SELECT builder. Columns may include expressions:
//
//	db.Select("u.*", "COUNT(p.id) AS post_count")
func (d *DB) Select(cols ...string) SelectBuilder {
	return SelectBuilder{db: d, sq: d.psql.Select(cols...)}
}

// Insert starts an INSERT builder for the given table.
func (d *DB) Insert(table string) InsertBuilder {
	return InsertBuilder{db: d, sq: d.psql.Insert(table)}
}

// Update starts an UPDATE builder for the given table.
func (d *DB) Update(table string) UpdateBuilder {
	return UpdateBuilder{db: d, sq: d.psql.Update(table)}
}

// Delete starts a DELETE builder for the given table.
func (d *DB) Delete(table string) DeleteBuilder {
	return DeleteBuilder{db: d, sq: d.psql.Delete(table)}
}

// ── Transaction ──────────────────────────────────────────────

// Tx begins a transaction, calls fn with a *TxDB that exposes the same
// builder API, then commits or rolls back automatically.
//
// If fn returns an error the transaction is rolled back.
// If fn panics the transaction is rolled back and the panic is re-raised.
func (d *DB) Tx(ctx context.Context, fn func(*TxDB) error) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	tdb := &TxDB{tx: tx, psql: d.psql}
	if err := fn(tdb); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback (%v) after: %w", rbErr, err)
		}
		return err
	}
	return tx.Commit()
}

// ── Raw escape hatches ───────────────────────────────────────
// Use these for CTEs, window functions, or any SQL squirrel can't express.

// RawOne executes rawSQL with args and scans a single row into dest.
// Returns ErrNotFound if no row matches.
func (d *DB) RawOne(ctx context.Context, dest any, rawSQL string, args ...any) error {
	if err := d.db.GetContext(ctx, dest, rawSQL, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// RawAll executes rawSQL with args and scans all rows into dest
// (dest must be a pointer to a slice).
func (d *DB) RawAll(ctx context.Context, dest any, rawSQL string, args ...any) error {
	return d.db.SelectContext(ctx, dest, rawSQL, args...)
}

// RawExec executes rawSQL and returns the number of rows affected.
func (d *DB) RawExec(ctx context.Context, rawSQL string, args ...any) (int64, error) {
	res, err := d.db.ExecContext(ctx, rawSQL, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ── Internal helpers ─────────────────────────────────────────

func isNoRows(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows) ||
		strings.Contains(err.Error(), "no rows in result set"))
}
