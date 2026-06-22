package dbx

import (
	"strings"
	"testing"
)

func TestCond_Eq(t *testing.T) {
	pred := Cond.Eq("id", 123)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "id") {
		t.Errorf("expected 'id' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 123 {
		t.Errorf("expected args [123], got %v", args)
	}
}

func TestCond_NotEq(t *testing.T) {
	pred := Cond.NotEq("status", "inactive")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "!=") && !strings.Contains(sql, "<>") {
		t.Errorf("expected != or <> in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "inactive" {
		t.Errorf("expected args ['inactive'], got %v", args)
	}
}

func TestCond_Gt(t *testing.T) {
	pred := Cond.Gt("age", 18)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, ">") {
		t.Errorf("expected '>' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 18 {
		t.Errorf("expected args [18], got %v", args)
	}
}

func TestCond_Lt(t *testing.T) {
	pred := Cond.Lt("price", 100)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "<") {
		t.Errorf("expected '<' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 100 {
		t.Errorf("expected args [100], got %v", args)
	}
}

func TestCond_GtOrEq(t *testing.T) {
	pred := Cond.GtOrEq("score", 50)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, ">=") {
		t.Errorf("expected '>=' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 50 {
		t.Errorf("expected args [50], got %v", args)
	}
}

func TestCond_LtOrEq(t *testing.T) {
	pred := Cond.LtOrEq("value", 75)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "<=") {
		t.Errorf("expected '<=' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != 75 {
		t.Errorf("expected args [75], got %v", args)
	}
}

func TestCond_Like(t *testing.T) {
	pred := Cond.Like("name", "%John%")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "LIKE") {
		t.Errorf("expected 'LIKE' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "%John%" {
		t.Errorf("expected args ['%%John%%'], got %v", args)
	}
}

func TestCond_ILike(t *testing.T) {
	pred := Cond.ILike("email", "%gmail%")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "ILIKE") {
		t.Errorf("expected 'ILIKE' in SQL, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "%gmail%" {
		t.Errorf("expected args ['%%gmail%%'], got %v", args)
	}
}

func TestCond_IsNull(t *testing.T) {
	pred := Cond.IsNull("deleted_at")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "IS NULL") {
		t.Errorf("expected 'IS NULL' in SQL, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected no args for IS NULL, got %v", args)
	}
}

func TestCond_IsNotNull(t *testing.T) {
	pred := Cond.IsNotNull("updated_at")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "IS NOT NULL") {
		t.Errorf("expected 'IS NOT NULL' in SQL, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected no args for IS NOT NULL, got %v", args)
	}
}

func TestCond_In_Single(t *testing.T) {
	pred := Cond.In("status", "active", "pending", "completed")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "IN") {
		t.Errorf("expected 'IN' in SQL, got: %s", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestCond_In_Empty(t *testing.T) {
	pred := Cond.In("id")
	_, args, err := pred.ToSql()
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if len(args) != 0 {
		t.Errorf("expected 0 args for empty IN, got %d", len(args))
	}
}

func TestCond_Between(t *testing.T) {
	pred := Cond.Between("age", 18, 65)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, ">=") || !strings.Contains(sql, "<=") {
		t.Errorf("expected >= and <= in SQL, got: %s", sql)
	}
	if len(args) != 2 || args[0] != 18 || args[1] != 65 {
		t.Errorf("expected args [18, 65], got %v", args)
	}
}

func TestCond_And(t *testing.T) {
	pred := Cond.And(
		Cond.Eq("status", "active"),
		Cond.Gt("age", 18),
	)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "AND") {
		t.Errorf("expected 'AND' in SQL, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestCond_And_Empty(t *testing.T) {
	pred := Cond.And()
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if len(args) != 0 {
		t.Errorf("expected 0 args for empty AND, got %d", len(args))
	}
}

func TestCond_Or(t *testing.T) {
	pred := Cond.Or(
		Cond.Eq("role", "admin"),
		Cond.Eq("role", "moderator"),
	)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "OR") {
		t.Errorf("expected 'OR' in SQL, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestCond_Or_Empty(t *testing.T) {
	pred := Cond.Or()
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if len(args) != 0 {
		t.Errorf("expected 0 args for empty OR, got %d", len(args))
	}
}

func TestCond_Search_Single(t *testing.T) {
	pred := Cond.Search("john", "name", "email")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "ILIKE") {
		t.Errorf("expected 'ILIKE' in SQL for search, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args for 2 columns, got %d", len(args))
	}
	// Check that args contain the search term with wildcards
	if !strings.Contains(args[0].(string), "john") {
		t.Errorf("expected search term 'john' in args, got %v", args)
	}
}

func TestCond_Search_NoColumns(t *testing.T) {
	pred := Cond.Search("term")
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if len(args) != 0 {
		t.Errorf("expected 0 args when no columns provided, got %d", len(args))
	}
}

func TestCond_Raw(t *testing.T) {
	pred := Cond.Raw("col @> ?", `{"tag"}`)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "@>") {
		t.Errorf("expected '@>' operator in SQL, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestCond_Nested_And_Or(t *testing.T) {
	pred := Cond.And(
		Cond.Eq("status", "active"),
		Cond.Or(
			Cond.Eq("role", "admin"),
			Cond.Eq("role", "moderator"),
		),
	)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if !strings.Contains(sql, "AND") || !strings.Contains(sql, "OR") {
		t.Errorf("expected 'AND' and 'OR' in SQL, got: %s", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestCond_Complex_Condition(t *testing.T) {
	pred := Cond.And(
		Cond.Gt("age", 18),
		Cond.LtOrEq("age", 65),
		Cond.In("status", "active", "pending"),
		Cond.IsNotNull("email"),
	)
	sql, args, err := pred.ToSql()
	_ = sql
	if err != nil {
		t.Fatalf("ToSql failed: %v", err)
	}

	if len(args) != 4 {
		t.Errorf("expected 4 args, got %d", len(args))
	}
}
