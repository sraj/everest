package dbx

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// Predicate is an alias for squirrel's Sqlizer.
// Callers use dbx.Predicate without importing squirrel directly.
type Predicate = sq.Sqlizer

// Cond is the global condition builder.
// Usage: dbx.Cond.Eq("id", 1)
var Cond = condBuilder{}

type condBuilder struct{}

// Eq generates col = $N.
func (condBuilder) Eq(col string, val any) Predicate { return sq.Eq{col: val} }

// NotEq generates col != $N.
func (condBuilder) NotEq(col string, val any) Predicate { return sq.NotEq{col: val} }

// Gt generates col > $N.
func (condBuilder) Gt(col string, val any) Predicate { return sq.Gt{col: val} }

// GtOrEq generates col >= $N.
func (condBuilder) GtOrEq(col string, val any) Predicate { return sq.GtOrEq{col: val} }

// Lt generates col < $N.
func (condBuilder) Lt(col string, val any) Predicate { return sq.Lt{col: val} }

// LtOrEq generates col <= $N.
func (condBuilder) LtOrEq(col string, val any) Predicate { return sq.LtOrEq{col: val} }

// Like generates col LIKE $N.
func (condBuilder) Like(col, pattern string) Predicate { return sq.Like{col: pattern} }

// ILike generates col ILIKE $N (case-insensitive, PostgreSQL).
func (condBuilder) ILike(col, pattern string) Predicate { return sq.ILike{col: pattern} }

// IsNull generates col IS NULL.
func (condBuilder) IsNull(col string) Predicate { return sq.Eq{col: nil} }

// IsNotNull generates col IS NOT NULL.
func (condBuilder) IsNotNull(col string) Predicate { return sq.NotEq{col: nil} }

// In generates col IN ($N, $M, …).
// Pass values as variadic any: dbx.Cond.In("id", 1, 2, 3)
func (condBuilder) In(col string, vals ...any) Predicate { return sq.Eq{col: vals} }

// Raw creates a predicate from a raw SQL fragment with positional $N args.
// Example: dbx.Cond.Raw("tags && ?", "{go,postgres}")
func (condBuilder) Raw(sql string, args ...any) Predicate { return sq.Expr(sql, args...) }

// And combines multiple predicates with AND.
func (condBuilder) And(preds ...Predicate) Predicate {
	a := make(sq.And, len(preds))
	for i, p := range preds {
		a[i] = p
	}
	return a
}

// Or combines multiple predicates with OR.
func (condBuilder) Or(preds ...Predicate) Predicate {
	o := make(sq.Or, len(preds))
	for i, p := range preds {
		o[i] = p
	}
	return o
}

// Search builds (col1 ILIKE %term% OR col2 ILIKE %term% …).
// Useful for free-text search across multiple columns.
func (condBuilder) Search(term string, cols ...string) Predicate {
	t := "%" + strings.ToLower(term) + "%"
	parts := make(sq.Or, len(cols))
	for i, c := range cols {
		parts[i] = sq.ILike{c: t}
	}
	return parts
}

// Between generates col >= lo AND col <= hi.
func (condBuilder) Between(col string, lo, hi any) Predicate {
	return sq.And{sq.GtOrEq{col: lo}, sq.LtOrEq{col: hi}}
}
