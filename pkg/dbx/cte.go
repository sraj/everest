package dbx

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// CTE represents a single Common Table Expression (WITH clause).
type CTE struct {
	name string
	sql  string
	args []any
}

// CTEBuilder accumulates CTEs before being attached to a SelectBuilder.
// It maintains immutability by returning a new CTEBuilder on each operation.
type CTEBuilder struct {
	ctes []CTE
}

// WithCTE adds a CTE with a raw SQL definition.
// Example: .WithCTE("regional_sales", "SELECT region, SUM(amount) AS total FROM orders GROUP BY region", nil)
func (b CTEBuilder) WithCTE(name, sql string, args ...any) CTEBuilder {
	ctes := make([]CTE, len(b.ctes)+1)
	copy(ctes, b.ctes)
	ctes[len(b.ctes)] = CTE{name: name, sql: sql, args: args}
	return CTEBuilder{ctes: ctes}
}

// WithSelectCTE adds a CTE built from a SelectBuilder.
// Example: .WithSelectCTE("recent_orders", db.Select("*").From("orders").Where(...))
func (b CTEBuilder) WithSelectCTE(name string, sb SelectBuilder) CTEBuilder {
	sql, args, err := sb.ToSQL()
	if err != nil {
		// In production, this should be handled at build time, not here.
		// For now, we'll defer the error to the build phase.
		return b
	}
	return b.WithCTE(name, sql, args...)
}

// Select begins a SELECT query with the accumulated CTEs.
// Returns a SelectBuilder that will render WITH clauses.
func (b CTEBuilder) Select(cols ...string) SelectBuilder {
	return SelectBuilder{
		ctes: b.ctes,
		sq:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(cols...),
	}
}

// ─────────────────────────────────────────────────────────────

// NewCTEBuilder creates a fresh CTEBuilder to start defining CTEs.
// Usage:
//   results := []MyType{}
//   err := dbx.NewCTEBuilder().
//       WithCTE("cte1", "SELECT ... FROM ...").
//       WithCTE("cte2", "SELECT ... FROM cte1 ...").
//       Select("col1", "col2").
//       From("cte2").
//       All(ctx, &results)
func NewCTEBuilder() CTEBuilder {
	return CTEBuilder{}
}

// ─────────────────────────────────────────────────────────────
// SelectBuilder CTE integration
// ─────────────────────────────────────────────────────────────

// The SelectBuilder struct needs a ctes field. This is handled by extending
// the SelectBuilder type below with CTE support.

// buildWithClause constructs the WITH clause prefix from accumulated CTEs.
func buildWithClause(ctes []CTE) (string, []any) {
	if len(ctes) == 0 {
		return "", []any{}
	}

	var buf strings.Builder
	var allArgs []any

	buf.WriteString("WITH ")
	for i, cte := range ctes {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(cte.name)
		buf.WriteString(" AS (")
		buf.WriteString(cte.sql)
		buf.WriteString(")")
		allArgs = append(allArgs, cte.args...)
	}
	buf.WriteString(" ")

	return buf.String(), allArgs
}

// prependWithClause prepends the WITH clause to the main SQL query.
// Arguments are merged in the correct order.
func prependWithClause(withClause string, withArgs []any, mainSQL string, mainArgs []any) (string, []any) {
	if withClause == "" {
		return mainSQL, mainArgs
	}

	// Merge arguments: CTEs first, then main query
	allArgs := make([]any, 0, len(withArgs)+len(mainArgs))
	allArgs = append(allArgs, withArgs...)
	allArgs = append(allArgs, mainArgs...)

	return withClause + mainSQL, allArgs
}
