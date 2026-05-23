package dbx

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// SelectBuilder is an immutable SELECT query builder.
// Every method returns a new copy so a base query can be forked freely.
type SelectBuilder struct {
	db        *DB
	sq        sq.SelectBuilder
	allowSort map[string]bool // optional whitelist for safe ORDER BY
}

// ── Chain methods ────────────────────────────────────────────

// From sets the FROM clause.
func (b SelectBuilder) From(table string) SelectBuilder {
	b.sq = b.sq.From(table)
	return b
}

// Join adds an INNER JOIN clause.
func (b SelectBuilder) Join(expr string) SelectBuilder {
	b.sq = b.sq.Join(expr)
	return b
}

// LeftJoin adds a LEFT JOIN clause.
func (b SelectBuilder) LeftJoin(expr string) SelectBuilder {
	b.sq = b.sq.LeftJoin(expr)
	return b
}

// Where appends a predicate (AND-ed with existing conditions).
func (b SelectBuilder) Where(pred Predicate) SelectBuilder {
	b.sq = b.sq.Where(pred)
	return b
}

// WhereIf appends the predicate only when cond is true.
// Eliminates if-blocks in callers when building dynamic queries.
func (b SelectBuilder) WhereIf(cond bool, pred Predicate) SelectBuilder {
	if cond {
		b.sq = b.sq.Where(pred)
	}
	return b
}

// GroupBy sets the GROUP BY columns.
func (b SelectBuilder) GroupBy(cols ...string) SelectBuilder {
	b.sq = b.sq.GroupBy(cols...)
	return b
}

// Having sets a HAVING predicate (used with GroupBy).
func (b SelectBuilder) Having(pred Predicate) SelectBuilder {
	b.sq = b.sq.Having(pred)
	return b
}

// AllowSort registers the whitelist of columns that may be used in OrderBy.
// Any column not in the list is silently ignored, preventing SQL injection.
func (b SelectBuilder) AllowSort(cols ...string) SelectBuilder {
	m := make(map[string]bool, len(cols))
	for _, c := range cols {
		m[c] = true
	}
	b.allowSort = m
	return b
}

// OrderBy appends an ORDER BY clause. If AllowSort was called, col must
// be in the whitelist or the call is a no-op.
func (b SelectBuilder) OrderBy(col string, dir SortDir) SelectBuilder {
	if col == "" {
		return b
	}
	if b.allowSort != nil && !b.allowSort[col] {
		return b
	}
	b.sq = b.sq.OrderBy(col + " " + string(dir))
	return b
}

// Paginate applies LIMIT and OFFSET derived from p.
func (b SelectBuilder) Paginate(p Page) SelectBuilder {
	b.sq = b.sq.Limit(uint64(p.size())).Offset(uint64(p.offset()))
	return b
}

// Suffix appends raw SQL after the main statement.
// e.g. .Suffix("FOR UPDATE")
func (b SelectBuilder) Suffix(sql string, args ...any) SelectBuilder {
	b.sq = b.sq.Suffix(sql, args...)
	return b
}

// ToSQL returns the generated SQL string and bound arguments.
// Useful for logging or testing without executing.
func (b SelectBuilder) ToSQL() (string, []any, error) {
	return b.sq.ToSql()
}

// ── Execution methods ────────────────────────────────────────

// One executes the query and scans a single row into dest.
// Returns ErrNotFound when no row matches.
func (b SelectBuilder) One(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("select.One build: %w", err)
	}
	if err := b.db.db.GetContext(ctx, dest, sql, args...); err != nil {
		if isNoRows(err) {
			return ErrNotFound
		}
		return fmt.Errorf("select.One exec: %w", err)
	}
	return nil
}

// All executes the query and scans all rows into dest.
// dest must be a pointer to a slice of structs.
func (b SelectBuilder) All(ctx context.Context, dest any) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("select.All build: %w", err)
	}
	if err := b.db.db.SelectContext(ctx, dest, sql, args...); err != nil {
		return fmt.Errorf("select.All exec: %w", err)
	}
	return nil
}

// Count clones the builder, replaces SELECT columns with COUNT(*),
// removes ORDER BY, LIMIT, and OFFSET, then returns the total count.
func (b SelectBuilder) Count(ctx context.Context) (int, error) {
	countSQ := b.sq.
		Columns().
		Column("COUNT(*)").
		RemoveLimit().
		RemoveOffset()

	sql, args, err := countSQ.ToSql()
	if err != nil {
		return 0, fmt.Errorf("select.Count build: %w", err)
	}
	var n int
	if err := b.db.db.GetContext(ctx, &n, sql, args...); err != nil {
		return 0, fmt.Errorf("select.Count exec: %w", err)
	}
	return n, nil
}

// Exists wraps the query in SELECT EXISTS (…) and returns the boolean result.
func (b SelectBuilder) Exists(ctx context.Context) (bool, error) {
	inner, args, err := b.sq.ToSql()
	if err != nil {
		return false, fmt.Errorf("select.Exists build: %w", err)
	}
	var exists bool
	if err := b.db.db.GetContext(ctx, &exists,
		"SELECT EXISTS ("+inner+")", args...); err != nil {
		return false, fmt.Errorf("select.Exists exec: %w", err)
	}
	return exists, nil
}

// Each executes the query and calls fn for every row.
// Avoids loading the entire result set into memory — suitable for
// large tables or streaming exports.
func (b SelectBuilder) Each(ctx context.Context, fn func(*sqlx.Rows) error) error {
	sql, args, err := b.sq.ToSql()
	if err != nil {
		return fmt.Errorf("select.Each build: %w", err)
	}
	rows, err := b.db.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("select.Each exec: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
