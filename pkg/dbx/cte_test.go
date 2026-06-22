package dbx

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
)

// TestCTEBuilder_WithCTE tests adding raw SQL CTEs.
func TestCTEBuilder_WithCTE(t *testing.T) {
	b := NewCTEBuilder()

	b1 := b.WithCTE("cte1", "SELECT * FROM table1")
	if len(b1.ctes) != 1 {
		t.Errorf("expected 1 CTE, got %d", len(b1.ctes))
	}
	if b1.ctes[0].name != "cte1" {
		t.Errorf("expected CTE name 'cte1', got %q", b1.ctes[0].name)
	}

	b2 := b1.WithCTE("cte2", "SELECT * FROM cte1")
	if len(b2.ctes) != 2 {
		t.Errorf("expected 2 CTEs, got %d", len(b2.ctes))
	}
}

// TestCTEBuilder_Select creates a SelectBuilder with CTEs.
func TestCTEBuilder_Select(t *testing.T) {
	b := NewCTEBuilder().
		WithCTE("temp", "SELECT 1 AS id")

	sb := b.Select("id")
	if len(sb.ctes) != 1 {
		t.Errorf("expected 1 CTE on SelectBuilder, got %d", len(sb.ctes))
	}
}

// TestSelectBuilder_WithCTE adds a CTE to an existing SelectBuilder.
func TestSelectBuilder_WithCTE(t *testing.T) {
	sb := SelectBuilder{
		sq:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("id"),
		ctes: []CTE{},
	}

	sb2 := sb.WithCTE("cte1", "SELECT * FROM table1")
	if len(sb2.ctes) != 1 {
		t.Errorf("expected 1 CTE after WithCTE, got %d", len(sb2.ctes))
	}

	sb3 := sb2.WithCTE("cte2", "SELECT * FROM cte1")
	if len(sb3.ctes) != 2 {
		t.Errorf("expected 2 CTEs after second WithCTE, got %d", len(sb3.ctes))
	}
}

// TestSelectBuilder_ToSQL_WithCTE verifies WITH clause is prepended.
func TestSelectBuilder_ToSQL_WithCTE(t *testing.T) {
	sb := SelectBuilder{
		sq:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("id").From("cte1"),
		ctes: []CTE{{name: "cte1", sql: "SELECT 1 AS id", args: nil}},
	}

	sql, args, err := sb.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !contains(sql, "WITH") {
		t.Errorf("expected WITH clause in SQL, got: %s", sql)
	}
	if !contains(sql, "cte1") {
		t.Errorf("expected 'cte1' in SQL, got: %s", sql)
	}
}

// TestBuildWithClause tests the WITH clause builder.
func TestBuildWithClause(t *testing.T) {
	ctes := []CTE{
		{name: "cte1", sql: "SELECT 1", args: []any{}},
		{name: "cte2", sql: "SELECT * FROM cte1", args: []any{}},
	}

	withClause, args := buildWithClause(ctes)

	if withClause == "" {
		t.Error("expected non-empty WITH clause")
	}
	if !contains(withClause, "WITH") {
		t.Errorf("expected 'WITH' in clause, got: %s", withClause)
	}
	if !contains(withClause, "cte1") {
		t.Errorf("expected 'cte1' in clause, got: %s", withClause)
	}
	if !contains(withClause, "cte2") {
		t.Errorf("expected 'cte2' in clause, got: %s", withClause)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
}

// TestBuildWithClause_WithArgs tests argument merging.
func TestBuildWithClause_WithArgs(t *testing.T) {
	ctes := []CTE{
		{name: "cte1", sql: "SELECT * FROM table1 WHERE id = $1", args: []any{123}},
	}

	withClause, args := buildWithClause(ctes)

	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
	if args[0] != 123 {
		t.Errorf("expected arg 123, got %v", args[0])
	}
}

// TestPrependWithClause tests WITH clause prepending.
func TestPrependWithClause(t *testing.T) {
	withClause := "WITH cte1 AS (SELECT 1) "
	withArgs := []any{}
	mainSQL := "SELECT * FROM cte1"
	mainArgs := []any{123}

	finalSQL, finalArgs := prependWithClause(withClause, withArgs, mainSQL, mainArgs)

	expected := withClause + mainSQL
	if finalSQL != expected {
		t.Errorf("expected %q, got %q", expected, finalSQL)
	}

	if len(finalArgs) != 1 || finalArgs[0] != 123 {
		t.Errorf("expected args [123], got %v", finalArgs)
	}
}

// TestPrependWithClause_EmptyWith tests prepending empty WITH clause.
func TestPrependWithClause_EmptyWith(t *testing.T) {
	mainSQL := "SELECT * FROM table1"
	mainArgs := []any{456}

	finalSQL, finalArgs := prependWithClause("", []any{}, mainSQL, mainArgs)

	if finalSQL != mainSQL {
		t.Errorf("expected %q, got %q", mainSQL, finalSQL)
	}
	if len(finalArgs) != 1 || finalArgs[0] != 456 {
		t.Errorf("expected args [456], got %v", finalArgs)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Example: Regional Sales by Product (your original query)
func ExampleCTE_RegionalSales(t *testing.T) {
	// This is a demonstration of how the query would be written.
	// In real use, db would be a connected *DB instance.

	// Pseudo-code example:
	/*
	type RegionalProductSales struct {
		Region       string  `db:"region"`
		Product      string  `db:"product"`
		ProductUnits int     `db:"product_units"`
		ProductSales float64 `db:"product_sales"`
	}

	var results []RegionalProductSales
	err := db.WithCTE().
		WithCTE("regional_sales", `
			SELECT region, SUM(amount) AS total_sales
			FROM orders
			GROUP BY region
		`).
		WithCTE("top_regions", `
			SELECT region
			FROM regional_sales
			WHERE total_sales > (SELECT SUM(total_sales)/10 FROM regional_sales)
		`).
		Select("region", "product", "SUM(quantity) AS product_units", "SUM(amount) AS product_sales").
		From("orders").
		Where(dbx.Cond.Raw("region IN (SELECT region FROM top_regions)")).
		GroupBy("region", "product").
		OrderBy("product_sales", dbx.DESC).
		All(ctx, &results)
	*/
}

// Example: CTE from SelectBuilder
func ExampleCTE_FromSelectBuilder(t *testing.T) {
	// Pseudo-code example:
	/*
	activeUsersCTE := db.Select("id", "name", "email").
		From("users").
		Where(dbx.Cond.Eq("status", "active"))

	type UserPost struct {
		UserID   string `db:"user_id"`
		UserName string `db:"user_name"`
		PostID   string `db:"post_id"`
		Title    string `db:"title"`
	}

	var posts []UserPost
	err := db.WithCTE().
		WithSelectCTE("active_users", activeUsersCTE).
		Select("u.id AS user_id", "u.name AS user_name", "p.id AS post_id", "p.title").
		From("active_users u").
		Join("posts p ON p.user_id = u.id").
		Where(dbx.Cond.Eq("p.published", true)).
		All(ctx, &posts)
	*/
}
