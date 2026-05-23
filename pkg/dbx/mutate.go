package dbx

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// ─────────────────────────────────────────────────────────────
// InsertBuilder
// ─────────────────────────────────────────────────────────────

// InsertBuilder is an immutable INSERT query builder.
type InsertBuilder struct {
	db *DB
	sq sq.InsertBuilder
}

// Columns sets the column list explicitly.
func (b InsertBuilder) Columns(cols ...string) InsertBuilder {
	b.sq = b.sq.Columns(cols...)
	return b
}

// Values adds a row of values matching the Columns order.
func (b InsertBuilder) Values(vals ...any) InsertBuilder {
	b.sq = b.sq.Values(vals...)
	return b
}

// SetMap inserts a single row from a map[column]value.
func (b InsertBuilder) SetMap(m map[string]any) InsertBuilder {
	b.sq = b.sq.SetMap(m)
	return b
}

// OnConflict appends an ON CONFLICT clause.
// Example: OnConflict("(email) DO UPDATE SET name = EXCLUDED.name")
func (b InsertBuilder) OnConflict(clause string) InsertBuilder {
	b.sq = b.sq.Suffix("ON CONFLICT " + clause)
	return b
}

// Returning appends a RETURNING clause.
// Use .One() to scan the returned row into a struct.
func (b InsertBuilder) Returning(cols ...string) InsertBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the INSERT and returns rows affected.
func (b InsertBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("insert.Exec build: %w", err)
	}
	res, err := b.db.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("insert.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// One runs INSERT … RETURNING and scans the single result row into dest.
func (b InsertBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("insert.One build: %w", err)
	}
	if err := b.db.db.GetContext(ctx, dest, sql, args...); err != nil {
		return fmt.Errorf("insert.One exec: %w", err)
	}
	return nil
}

// ToSQL returns the raw SQL and args for inspection or logging.
func (b InsertBuilder) ToSQL() (string, []any, error) { return b.sq.ToSql() }

// ─────────────────────────────────────────────────────────────
// UpdateBuilder
// ─────────────────────────────────────────────────────────────

// UpdateBuilder is an immutable UPDATE query builder.
type UpdateBuilder struct {
	db *DB
	sq sq.UpdateBuilder
}

// Set sets col = val.
func (b UpdateBuilder) Set(col string, val any) UpdateBuilder {
	b.sq = b.sq.Set(col, val)
	return b
}

// SetMap sets multiple columns from a map[column]value.
func (b UpdateBuilder) SetMap(m map[string]any) UpdateBuilder {
	b.sq = b.sq.SetMap(m)
	return b
}

// SetExpr sets col to a raw SQL expression with optional bound args.
// Example: SetExpr("salary", "salary * ?", 1.05)
func (b UpdateBuilder) SetExpr(col, expr string, args ...any) UpdateBuilder {
	b.sq = b.sq.Set(col, sq.Expr(expr, args...))
	return b
}

// Where appends a predicate.
func (b UpdateBuilder) Where(pred Predicate) UpdateBuilder {
	b.sq = b.sq.Where(pred)
	return b
}

// WhereIf appends the predicate only when cond is true.
func (b UpdateBuilder) WhereIf(cond bool, pred Predicate) UpdateBuilder {
	if cond {
		b.sq = b.sq.Where(pred)
	}
	return b
}

// Returning appends a RETURNING clause.
func (b UpdateBuilder) Returning(cols ...string) UpdateBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the UPDATE and returns rows affected.
func (b UpdateBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("update.Exec build: %w", err)
	}
	res, err := b.db.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("update.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ExecMustAffect runs the UPDATE and returns ErrNoRows if 0 rows changed.
func (b UpdateBuilder) ExecMustAffect(ctx context.Context) error {
	n, err := b.Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

// One runs UPDATE … RETURNING and scans the result row into dest.
// Returns ErrNotFound if no row was updated.
func (b UpdateBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("update.One build: %w", err)
	}
	if err := b.db.db.GetContext(ctx, dest, sql, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return fmt.Errorf("update.One exec: %w", err)
	}
	return nil
}

// ToSQL returns the raw SQL and args for inspection or logging.
func (b UpdateBuilder) ToSQL() (string, []any, error) { return b.sq.ToSql() }

// ─────────────────────────────────────────────────────────────
// DeleteBuilder
// ─────────────────────────────────────────────────────────────

// DeleteBuilder is an immutable DELETE query builder.
type DeleteBuilder struct {
	db *DB
	sq sq.DeleteBuilder
}

// Where appends a predicate.
func (b DeleteBuilder) Where(pred Predicate) DeleteBuilder {
	b.sq = b.sq.Where(pred)
	return b
}

// WhereIf appends the predicate only when cond is true.
func (b DeleteBuilder) WhereIf(cond bool, pred Predicate) DeleteBuilder {
	if cond {
		b.sq = b.sq.Where(pred)
	}
	return b
}

// Returning appends a RETURNING clause.
func (b DeleteBuilder) Returning(cols ...string) DeleteBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the DELETE and returns rows affected.
func (b DeleteBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("delete.Exec build: %w", err)
	}
	res, err := b.db.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("delete.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ExecMustAffect runs DELETE and returns ErrNoRows if 0 rows were deleted.
func (b DeleteBuilder) ExecMustAffect(ctx context.Context) error {
	n, err := b.Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

// ToSQL returns the raw SQL and args for inspection or logging.
func (b DeleteBuilder) ToSQL() (string, []any, error) { return b.sq.ToSql() }
