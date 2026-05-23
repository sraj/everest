package dbx

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// TxDB mirrors *DB but all operations run inside the same transaction.
// Obtained exclusively via DB.Tx().
type TxDB struct {
	tx   *sqlx.Tx
	psql sq.StatementBuilderType
}

// Select starts a SELECT builder on the transaction.
func (t *TxDB) Select(cols ...string) TxSelectBuilder {
	return TxSelectBuilder{tx: t.tx, sq: t.psql.Select(cols...)}
}

// Insert starts an INSERT builder on the transaction.
func (t *TxDB) Insert(table string) TxInsertBuilder {
	return TxInsertBuilder{tx: t.tx, sq: t.psql.Insert(table)}
}

// Update starts an UPDATE builder on the transaction.
func (t *TxDB) Update(table string) TxUpdateBuilder {
	return TxUpdateBuilder{tx: t.tx, sq: t.psql.Update(table)}
}

// Delete starts a DELETE builder on the transaction.
func (t *TxDB) Delete(table string) TxDeleteBuilder {
	return TxDeleteBuilder{tx: t.tx, sq: t.psql.Delete(table)}
}

// RawOne executes rawSQL on the transaction and scans a single row into dest.
func (t *TxDB) RawOne(ctx context.Context, dest any, rawSQL string, args ...any) error {
	if err := t.tx.GetContext(ctx, dest, rawSQL, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// RawAll executes rawSQL on the transaction and scans all rows into dest.
func (t *TxDB) RawAll(ctx context.Context, dest any, rawSQL string, args ...any) error {
	return t.tx.SelectContext(ctx, dest, rawSQL, args...)
}

// RawExec executes rawSQL on the transaction and returns rows affected.
func (t *TxDB) RawExec(ctx context.Context, rawSQL string, args ...any) (int64, error) {
	res, err := t.tx.ExecContext(ctx, rawSQL, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ─────────────────────────────────────────────────────────────
// TxSelectBuilder
// ─────────────────────────────────────────────────────────────

// TxSelectBuilder is a SELECT builder bound to a transaction.
type TxSelectBuilder struct {
	tx *sqlx.Tx
	sq sq.SelectBuilder
}

func (b TxSelectBuilder) From(table string) TxSelectBuilder     { b.sq = b.sq.From(table); return b }
func (b TxSelectBuilder) Join(expr string) TxSelectBuilder      { b.sq = b.sq.Join(expr); return b }
func (b TxSelectBuilder) LeftJoin(expr string) TxSelectBuilder  { b.sq = b.sq.LeftJoin(expr); return b }
func (b TxSelectBuilder) GroupBy(cols ...string) TxSelectBuilder { b.sq = b.sq.GroupBy(cols...); return b }

func (b TxSelectBuilder) Where(p Predicate) TxSelectBuilder { b.sq = b.sq.Where(p); return b }
func (b TxSelectBuilder) WhereIf(cond bool, p Predicate) TxSelectBuilder {
	if cond {
		b.sq = b.sq.Where(p)
	}
	return b
}

func (b TxSelectBuilder) OrderBy(col string, dir SortDir) TxSelectBuilder {
	b.sq = b.sq.OrderBy(col + " " + string(dir))
	return b
}
func (b TxSelectBuilder) Limit(n uint64) TxSelectBuilder { b.sq = b.sq.Limit(n); return b }
func (b TxSelectBuilder) Suffix(sql string, args ...any) TxSelectBuilder {
	b.sq = b.sq.Suffix(sql, args...)
	return b
}

// One scans a single row into dest. Returns ErrNotFound when no row matches.
func (b TxSelectBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("tx.select.One build: %w", err)
	}
	if err := b.tx.GetContext(ctx, dest, sql, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return fmt.Errorf("tx.select.One exec: %w", err)
	}
	return nil
}

// All scans all rows into dest (pointer to slice).
func (b TxSelectBuilder) All(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("tx.select.All build: %w", err)
	}
	return b.tx.SelectContext(ctx, dest, sql, args...)
}

// Exists returns true if at least one row matches.
func (b TxSelectBuilder) Exists(ctx context.Context) (bool, error) {
	inner, args, err := b.sq.ToSql()
	if err != nil {
		return false, fmt.Errorf("tx.select.Exists build: %w", err)
	}
	var exists bool
	err = b.tx.GetContext(ctx, &exists, "SELECT EXISTS ("+inner+")", args...)
	return exists, err
}

// ─────────────────────────────────────────────────────────────
// TxInsertBuilder
// ─────────────────────────────────────────────────────────────

// TxInsertBuilder is an INSERT builder bound to a transaction.
type TxInsertBuilder struct {
	tx *sqlx.Tx
	sq sq.InsertBuilder
}

func (b TxInsertBuilder) Columns(cols ...string) TxInsertBuilder { b.sq = b.sq.Columns(cols...); return b }
func (b TxInsertBuilder) Values(vals ...any) TxInsertBuilder     { b.sq = b.sq.Values(vals...); return b }
func (b TxInsertBuilder) SetMap(m map[string]any) TxInsertBuilder { b.sq = b.sq.SetMap(m); return b }

func (b TxInsertBuilder) OnConflict(clause string) TxInsertBuilder {
	b.sq = b.sq.Suffix("ON CONFLICT " + clause)
	return b
}
func (b TxInsertBuilder) Returning(cols ...string) TxInsertBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the INSERT and returns rows affected.
func (b TxInsertBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("tx.insert.Exec build: %w", err)
	}
	res, err := b.tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("tx.insert.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// One runs INSERT … RETURNING and scans the row into dest.
func (b TxInsertBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("tx.insert.One build: %w", err)
	}
	return b.tx.GetContext(ctx, dest, sql, args...)
}

// ─────────────────────────────────────────────────────────────
// TxUpdateBuilder
// ─────────────────────────────────────────────────────────────

// TxUpdateBuilder is an UPDATE builder bound to a transaction.
type TxUpdateBuilder struct {
	tx *sqlx.Tx
	sq sq.UpdateBuilder
}

func (b TxUpdateBuilder) Set(col string, val any) TxUpdateBuilder { b.sq = b.sq.Set(col, val); return b }
func (b TxUpdateBuilder) SetMap(m map[string]any) TxUpdateBuilder  { b.sq = b.sq.SetMap(m); return b }
func (b TxUpdateBuilder) SetExpr(col, expr string, args ...any) TxUpdateBuilder {
	b.sq = b.sq.Set(col, sq.Expr(expr, args...))
	return b
}
func (b TxUpdateBuilder) Where(p Predicate) TxUpdateBuilder { b.sq = b.sq.Where(p); return b }
func (b TxUpdateBuilder) WhereIf(cond bool, p Predicate) TxUpdateBuilder {
	if cond {
		b.sq = b.sq.Where(p)
	}
	return b
}
func (b TxUpdateBuilder) Returning(cols ...string) TxUpdateBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the UPDATE and returns rows affected.
func (b TxUpdateBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("tx.update.Exec build: %w", err)
	}
	res, err := b.tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("tx.update.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ExecMustAffect runs UPDATE and returns ErrNoRows if 0 rows changed.
func (b TxUpdateBuilder) ExecMustAffect(ctx context.Context) error {
	n, err := b.Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}

// One runs UPDATE … RETURNING and scans the result into dest.
func (b TxUpdateBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("tx.update.One build: %w", err)
	}
	if err := b.tx.GetContext(ctx, dest, sql, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return fmt.Errorf("tx.update.One exec: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// TxDeleteBuilder
// ─────────────────────────────────────────────────────────────

// TxDeleteBuilder is a DELETE builder bound to a transaction.
type TxDeleteBuilder struct {
	tx *sqlx.Tx
	sq sq.DeleteBuilder
}

func (b TxDeleteBuilder) Where(p Predicate) TxDeleteBuilder { b.sq = b.sq.Where(p); return b }
func (b TxDeleteBuilder) WhereIf(cond bool, p Predicate) TxDeleteBuilder {
	if cond {
		b.sq = b.sq.Where(p)
	}
	return b
}
func (b TxDeleteBuilder) Returning(cols ...string) TxDeleteBuilder {
	b.sq = b.sq.Suffix("RETURNING " + strings.Join(cols, ", "))
	return b
}

// Exec runs the DELETE and returns rows affected.
func (b TxDeleteBuilder) Exec(ctx context.Context) (int64, error) {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return 0, fmt.Errorf("tx.delete.Exec build: %w", err)
	}
	res, err := b.tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("tx.delete.Exec exec: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ExecMustAffect runs DELETE and returns ErrNoRows if 0 rows were removed.
func (b TxDeleteBuilder) ExecMustAffect(ctx context.Context) error {
	n, err := b.Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRows
	}
	return nil
}
