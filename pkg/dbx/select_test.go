package dbx

import (
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestSelectBuilder_From(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("id", "name"),
	}

	sb2 := sb.From("users")
	if sb2.sq == nil {
		t.Error("From should return new SelectBuilder")
	}
}

func TestSelectBuilder_Join(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("orders"),
	}

	sb2 := sb.Join("users u ON u.id = orders.user_id")
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "JOIN") {
		t.Errorf("expected JOIN in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_LeftJoin(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("orders"),
	}

	sb2 := sb.LeftJoin("users u ON u.id = orders.user_id")
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "LEFT JOIN") {
		t.Errorf("expected LEFT JOIN in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_Where(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.Where(Cond.Eq("active", true))
	sql, args, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE in SQL, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestSelectBuilder_WhereIf_True(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.WhereIf(true, Cond.Eq("role", "admin"))
	sql, args, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE when condition is true, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestSelectBuilder_WhereIf_False(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.WhereIf(false, Cond.Eq("role", "admin"))
	sql, args, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if strings.Contains(sql, "WHERE") {
		t.Errorf("expected no WHERE when condition is false, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
}

func TestSelectBuilder_GroupBy(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("category", "COUNT(*)").From("orders"),
	}

	sb2 := sb.GroupBy("category")
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "GROUP BY") {
		t.Errorf("expected GROUP BY in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_GroupBy_Multiple(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("region", "product", "SUM(amount)").From("orders"),
	}

	sb2 := sb.GroupBy("region", "product")
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "GROUP BY") {
		t.Errorf("expected GROUP BY in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_Having(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("category", "COUNT(*)").From("orders").GroupBy("category"),
	}

	sb2 := sb.Having(Cond.Gt("COUNT(*)", 100))
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "HAVING") {
		t.Errorf("expected HAVING in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_OrderBy_Asc(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.OrderBy("name", ASC)
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "ASC") {
		t.Errorf("expected ASC in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_OrderBy_Desc(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.OrderBy("created_at", DESC)
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "DESC") {
		t.Errorf("expected DESC in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_OrderBy_EmptyColumn(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.OrderBy("", DESC)
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected no ORDER BY for empty column, got: %s", sql)
	}
}

func TestSelectBuilder_AllowSort_Whitelist(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.AllowSort("name", "email", "created_at")
	sb3 := sb2.OrderBy("name", ASC)
	sql, _, err := sb3.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY for whitelisted column, got: %s", sql)
	}
}

func TestSelectBuilder_AllowSort_Reject_Unlisted(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.AllowSort("name", "email")
	sb3 := sb2.OrderBy("password", ASC) // Try to sort by non-whitelisted column
	sql, _, err := sb3.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected no ORDER BY for non-whitelisted column, got: %s", sql)
	}
}

func TestSelectBuilder_Paginate(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.Paginate(Page{Number: 2, Size: 20})
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "LIMIT") {
		t.Errorf("expected LIMIT in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "OFFSET") {
		t.Errorf("expected OFFSET in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_Suffix(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.Suffix("FOR UPDATE")
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "FOR UPDATE") {
		t.Errorf("expected FOR UPDATE suffix in SQL, got: %s", sql)
	}
}

func TestSelectBuilder_ToSQL_Simple(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("id", "name").From("users"),
	}

	sql, args, err := sb.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !strings.Contains(sql, "SELECT") {
		t.Errorf("expected SELECT in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "FROM") {
		t.Errorf("expected FROM in SQL, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected no args for simple query, got %v", args)
	}
}

func TestSelectBuilder_Chain_Multiple_Conditions(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.Where(Cond.Eq("status", "active")).
		Where(Cond.Gt("age", 18)).
		OrderBy("name", ASC)

	sql, args, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE in SQL, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestSelectBuilder_Immutability(t *testing.T) {
	original := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	modified := original.Where(Cond.Eq("id", 123))

	// Original should be unchanged
	origSQL, origArgs, _ := original.ToSQL()
	if strings.Contains(origSQL, "WHERE") {
		t.Error("original SelectBuilder was modified")
	}
	if len(origArgs) != 0 {
		t.Error("original SelectBuilder args were modified")
	}

	// Modified should have the new condition
	modSQL, modArgs, _ := modified.ToSQL()
	if !strings.Contains(modSQL, "WHERE") {
		t.Error("modified SelectBuilder should have WHERE clause")
	}
	if len(modArgs) != 1 {
		t.Error("modified SelectBuilder should have args")
	}
}

func TestSelectBuilder_Multiple_OrderBy(t *testing.T) {
	sb := SelectBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("users"),
	}

	sb2 := sb.OrderBy("region", ASC).OrderBy("name", ASC)
	sql, _, err := sb2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL, got: %s", sql)
	}
}
