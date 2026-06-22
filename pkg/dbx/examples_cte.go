package dbx

import (
	"context"
)

// Example: Regional Sales Query (Your Original Use Case)
//
// This example demonstrates how to write the complex CTE query
// using the new dbx CTE support:
//
// WITH regional_sales AS (
//   SELECT region, SUM(amount) AS total_sales FROM orders GROUP BY region
// ),
// top_regions AS (
//   SELECT region FROM regional_sales
//   WHERE total_sales > (SELECT SUM(total_sales)/10 FROM regional_sales)
// )
// SELECT region, product, SUM(quantity) AS product_units, SUM(amount) AS product_sales
// FROM orders
// WHERE region IN (SELECT region FROM top_regions)
// GROUP BY region, product
// ORDER BY product_sales DESC
//
// Usage in code:
//
//	type RegionalProductSales struct {
//		Region       string  `db:"region"`
//		Product      string  `db:"product"`
//		ProductUnits int     `db:"product_units"`
//		ProductSales float64 `db:"product_sales"`
//	}
//
//	var results []RegionalProductSales
//	err := db.WithCTE().
//		WithCTE("regional_sales", `
//			SELECT region, SUM(amount) AS total_sales
//			FROM orders
//			GROUP BY region
//		`).
//		WithCTE("top_regions", `
//			SELECT region
//			FROM regional_sales
//			WHERE total_sales > (SELECT SUM(total_sales)/10 FROM regional_sales)
//		`).
//		Select(
//			"region",
//			"product",
//			"SUM(quantity) AS product_units",
//			"SUM(amount) AS product_sales",
//		).
//		From("orders").
//		Where(dbx.Cond.Raw("region IN (SELECT region FROM top_regions)")).
//		GroupBy("region", "product").
//		OrderBy("product_sales", dbx.DESC).
//		All(ctx, &results)

// Example: Hierarchical Data with Window Functions
//
// Build a manager-employee hierarchy with row numbering:
//
// WITH numbered_employees AS (
//   SELECT id, name, manager_id, department,
//          ROW_NUMBER() OVER (PARTITION BY department ORDER BY name) AS emp_num
//   FROM employees
// )
// SELECT * FROM numbered_employees
// WHERE emp_num <= 5
//
// Usage:
//
//	type EmployeeRanked struct {
//		ID        string `db:"id"`
//		Name      string `db:"name"`
//		ManagerID string `db:"manager_id"`
//		Dept      string `db:"department"`
//		Rank      int    `db:"emp_num"`
//	}
//
//	var employees []EmployeeRanked
//	err := db.WithCTE().
//		WithCTE("numbered_employees", `
//			SELECT id, name, manager_id, department,
//				   ROW_NUMBER() OVER (PARTITION BY department ORDER BY name) AS emp_num
//			FROM employees
//		`).
//		Select("id", "name", "manager_id", "department", "emp_num").
//		From("numbered_employees").
//		Where(dbx.Cond.LtOrEq("emp_num", 5)).
//		All(ctx, &employees)

// Example: Multi-CTE Aggregation Pipeline
//
// Process data through multiple stages (filter → aggregate → rank):
//
// WITH recent_transactions AS (
//   SELECT * FROM transactions WHERE created_at >= NOW() - INTERVAL '30 days'
// ),
// category_totals AS (
//   SELECT category, SUM(amount) AS total, COUNT(*) AS count
//   FROM recent_transactions
//   GROUP BY category
// ),
// ranked_categories AS (
//   SELECT category, total, count,
//          RANK() OVER (ORDER BY total DESC) AS rank
//   FROM category_totals
// )
// SELECT * FROM ranked_categories WHERE rank <= 10
//
// Usage:
//
//	type CategoryRanking struct {
//		Category string  `db:"category"`
//		Total    float64 `db:"total"`
//		Count    int     `db:"count"`
//		Rank     int     `db:"rank"`
//	}
//
//	var rankings []CategoryRanking
//	err := db.WithCTE().
//		WithCTE("recent_transactions", `
//			SELECT * FROM transactions
//			WHERE created_at >= NOW() - INTERVAL '30 days'
//		`).
//		WithCTE("category_totals", `
//			SELECT category, SUM(amount) AS total, COUNT(*) AS count
//			FROM recent_transactions
//			GROUP BY category
//		`).
//		WithCTE("ranked_categories", `
//			SELECT category, total, count,
//				   RANK() OVER (ORDER BY total DESC) AS rank
//			FROM category_totals
//		`).
//		Select("category", "total", "count", "rank").
//		From("ranked_categories").
//		Where(dbx.Cond.LtOrEq("rank", 10)).
//		OrderBy("rank", dbx.ASC).
//		All(ctx, &rankings)

// Example: CTE from SelectBuilder (Composable Subqueries)
//
// Build a CTE from an existing SelectBuilder for better composability:
//
// Usage:
//
//	// Build a reusable CTE definition
//	activeUsersCTE := db.Select("id", "name", "email", "created_at").
//		From("users").
//		Where(dbx.Cond.Eq("status", "active")).
//		Where(dbx.Cond.Gt("last_login_at", time.Now().AddDate(0, 0, -30)))
//
//	// Use it in multiple queries
//	type UserPost struct {
//		UserID   string `db:"user_id"`
//		UserName string `db:"user_name"`
//		PostID   string `db:"post_id"`
//		Title    string `db:"title"`
//	}
//
//	var posts []UserPost
//	err := db.WithCTE().
//		WithSelectCTE("active_users", activeUsersCTE).
//		Select(
//			"u.id AS user_id",
//			"u.name AS user_name",
//			"p.id AS post_id",
//			"p.title",
//		).
//		From("active_users u").
//		Join("posts p ON p.user_id = u.id").
//		Where(dbx.Cond.Eq("p.published", true)).
//		OrderBy("p.published_at", dbx.DESC).
//		All(ctx, &posts)

// Example: Recursive CTE (Organization Hierarchy)
//
// Walk an organization tree to find all descendants of a manager:
//
// WITH RECURSIVE org_tree AS (
//   SELECT id, name, manager_id, 1 AS level
//   FROM employees
//   WHERE id = $1
//   UNION ALL
//   SELECT e.id, e.name, e.manager_id, ot.level + 1
//   FROM employees e
//   JOIN org_tree ot ON e.manager_id = ot.id
//   WHERE ot.level < 10
// )
// SELECT * FROM org_tree ORDER BY level, name
//
// Usage:
//
//	type OrgHierarchy struct {
//		ID        string `db:"id"`
//		Name      string `db:"name"`
//		ManagerID string `db:"manager_id"`
//		Level     int    `db:"level"`
//	}
//
//	managerID := "emp_123"
//	var hierarchy []OrgHierarchy
//	err := db.WithCTE().
//		WithCTE("org_tree", `
//			WITH RECURSIVE org_tree AS (
//				SELECT id, name, manager_id, 1 AS level
//				FROM employees
//				WHERE id = $1
//				UNION ALL
//				SELECT e.id, e.name, e.manager_id, ot.level + 1
//				FROM employees e
//				JOIN org_tree ot ON e.manager_id = ot.id
//				WHERE ot.level < 10
//			)
//			SELECT * FROM org_tree
//		`, managerID).
//		Select("id", "name", "manager_id", "level").
//		From("org_tree").
//		OrderBy("level", dbx.ASC).
//		OrderBy("name", dbx.ASC).
//		All(ctx, &hierarchy)

// Example: Using CTEs for Performance (Materialized Subqueries)
//
// CTEs can help optimize queries by forcing materialization of expensive subqueries:
//
// Usage:
//
//	type SalesReport struct {
//		Region       string  `db:"region"`
//		Q1Sales      float64 `db:"q1_sales"`
//		Q2Sales      float64 `db:"q2_sales"`
//		YTDSales     float64 `db:"ytd_sales"`
//		SalesPerMonth float64 `db:"sales_per_month"`
//	}
//
//	var report []SalesReport
//	err := db.WithCTE().
//		WithCTE("q1_sales", `
//			SELECT region, SUM(amount) AS total
//			FROM orders
//			WHERE date_part('quarter', order_date) = 1
//			GROUP BY region
//		`).
//		WithCTE("q2_sales", `
//			SELECT region, SUM(amount) AS total
//			FROM orders
//			WHERE date_part('quarter', order_date) = 2
//			GROUP BY region
//		`).
//		WithCTE("ytd_totals", `
//			SELECT region, SUM(amount) AS total
//			FROM orders
//			WHERE date_part('year', order_date) = date_part('year', NOW())
//			GROUP BY region
//		`).
//		Select(
//			"COALESCE(q1.region, q2.region, ytd.region) AS region",
//			"COALESCE(q1.total, 0) AS q1_sales",
//			"COALESCE(q2.total, 0) AS q2_sales",
//			"ytd.total AS ytd_sales",
//			"ytd.total / 3 AS sales_per_month",
//		).
//		From("ytd_totals ytd").
//		LeftJoin("q1_sales q1 ON q1.region = ytd.region").
//		LeftJoin("q2_sales q2 ON q2.region = ytd.region").
//		OrderBy("ytd_sales", dbx.DESC).
//		All(ctx, &report)

// ExampleCTEUsage demonstrates CTE builder usage (used in test documentation).
func ExampleCTEUsage(db *DB, ctx context.Context) error {
	type Result struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	var results []Result
	err := db.WithCTE().
		WithCTE("temp", "SELECT 1 AS id, 'test' AS name").
		Select("id", "name").
		From("temp").
		All(ctx, &results)

	return err
}
