package dbx

import (
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

// ==================== INSERT TESTS ====================

func TestInsertBuilder_Columns(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users"),
	}

	ib2 := ib.Columns("id", "name", "email")
	if ib2.sq == nil {
		t.Error("Columns should return new InsertBuilder")
	}
}

func TestInsertBuilder_Values(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users").Columns("id", "name", "email"),
	}

	ib2 := ib.Values("123", "Alice", "alice@example.com")
	sql, args, err := ib2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT") {
		t.Errorf("expected INSERT in SQL, got: %s", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestInsertBuilder_SetMap(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users"),
	}

	m := map[string]any{
		"id":    "123",
		"name":  "Alice",
		"email": "alice@example.com",
	}
	ib2 := ib.SetMap(m)
	sql, args, err := ib2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT") {
		t.Errorf("expected INSERT in SQL, got: %s", sql)
	}
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
}

func TestInsertBuilder_OnConflict(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users").Columns("id", "name").Values("123", "Alice"),
	}

	ib2 := ib.OnConflict("(id) DO UPDATE SET name = EXCLUDED.name")
	sql, _, err := ib2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ON CONFLICT") {
		t.Errorf("expected ON CONFLICT in SQL, got: %s", sql)
	}
}

func TestInsertBuilder_Returning(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users").Columns("id", "name").Values("123", "Alice"),
	}

	ib2 := ib.Returning("id", "name")
	sql, _, err := ib2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "RETURNING") {
		t.Errorf("expected RETURNING in SQL, got: %s", sql)
	}
}

func TestInsertBuilder_ToSQL(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("users").
			Columns("id", "name", "email").
			Values("123", "Alice", "alice@example.com"),
	}

	sql, args, err := ib.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT INTO") {
		t.Errorf("expected INSERT INTO in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "users") {
		t.Errorf("expected table name in SQL, got: %s", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

// ==================== UPDATE TESTS ====================

func TestUpdateBuilder_Set(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users"),
	}

	ub2 := ub.Set("name", "Bob").Set("email", "bob@example.com")
	sql, args, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "UPDATE") {
		t.Errorf("expected UPDATE in SQL, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestUpdateBuilder_SetMap(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users"),
	}

	m := map[string]any{
		"name":       "Bob",
		"email":      "bob@example.com",
		"updated_at": "2024-01-01",
	}
	ub2 := ub.SetMap(m)
	sql, args, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "UPDATE") {
		t.Errorf("expected UPDATE in SQL, got: %s", sql)
	}
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
}

func TestUpdateBuilder_Where(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("name", "Charlie"),
	}

	ub2 := ub.Where(Cond.Eq("id", "123"))
	sql, args, err := ub2.ToSQL()
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

func TestUpdateBuilder_WhereIf_True(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("status", "active"),
	}

	ub2 := ub.WhereIf(true, Cond.Eq("id", "123"))
	sql, args, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE when condition is true, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestUpdateBuilder_WhereIf_False(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("status", "active"),
	}

	ub2 := ub.WhereIf(false, Cond.Eq("id", "123"))
	sql, args, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if strings.Contains(sql, "WHERE") {
		t.Errorf("expected no WHERE when condition is false, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg (only SET value), got %d", len(args))
	}
}

func TestUpdateBuilder_Returning(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("name", "Dave"),
	}

	ub2 := ub.Returning("id", "name")
	sql, _, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "RETURNING") {
		t.Errorf("expected RETURNING in SQL, got: %s", sql)
	}
}

func TestUpdateBuilder_OrderBy(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("status", "archived"),
	}

	ub2 := ub.OrderBy("created_at", DESC)
	sql, _, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL, got: %s", sql)
	}
}

func TestUpdateBuilder_Limit(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").Set("status", "reviewed"),
	}

	ub2 := ub.Limit(10)
	sql, _, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "LIMIT") {
		t.Errorf("expected LIMIT in SQL, got: %s", sql)
	}
}

func TestUpdateBuilder_ToSQL(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users").
			Set("name", "Eve").
			Set("email", "eve@example.com"),
	}

	sql, args, err := ub.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "UPDATE users") {
		t.Errorf("expected UPDATE users in SQL, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

// ==================== DELETE TESTS ====================

func TestDeleteBuilder_Where(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("users"),
	}

	db2 := db.Where(Cond.Eq("id", "123"))
	sql, args, err := db2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "DELETE") {
		t.Errorf("expected DELETE in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE in SQL, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestDeleteBuilder_OrderBy(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("sessions"),
	}

	db2 := db.OrderBy("expires_at", ASC)
	sql, _, err := db2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "ORDER BY") {
		t.Errorf("expected ORDER BY in SQL, got: %s", sql)
	}
}

func TestDeleteBuilder_Limit(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("logs"),
	}

	db2 := db.Limit(100)
	sql, _, err := db2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "LIMIT") {
		t.Errorf("expected LIMIT in SQL, got: %s", sql)
	}
}

func TestDeleteBuilder_Returning(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("users"),
	}

	db2 := db.Returning("id", "name")
	sql, _, err := db2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "RETURNING") {
		t.Errorf("expected RETURNING in SQL, got: %s", sql)
	}
}

func TestDeleteBuilder_ToSQL(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("users").Where(Cond.Eq("status", "inactive")),
	}

	sql, args, err := db.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "DELETE FROM users") {
		t.Errorf("expected DELETE FROM users in SQL, got: %s", sql)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestDeleteBuilder_Multiple_Conditions(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("sessions"),
	}

	db2 := db.Where(Cond.Lt("expires_at", "2024-01-01")).
		Where(Cond.Eq("status", "expired"))

	sql, args, err := db2.ToSQL()
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

// ==================== INTEGRATION TESTS ====================

func TestMutateBuilders_ComplexInsert(t *testing.T) {
	ib := InsertBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("orders"),
	}

	ib2 := ib.Columns("id", "user_id", "amount", "status").
		Values("ord-123", "user-456", 99.99, "pending").
		Returning("id", "user_id", "amount")

	sql, args, err := ib2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "INSERT INTO orders") {
		t.Errorf("expected INSERT INTO orders, got: %s", sql)
	}
	if !strings.Contains(sql, "RETURNING") {
		t.Errorf("expected RETURNING, got: %s", sql)
	}
	if len(args) != 4 {
		t.Errorf("expected 4 args, got %d", len(args))
	}
}

func TestMutateBuilders_ComplexUpdate(t *testing.T) {
	ub := UpdateBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update("users"),
	}

	ub2 := ub.Set("name", "Updated Name").
		Set("email", "new@example.com").
		Where(Cond.Eq("id", "user-123")).
		Returning("id", "name", "email")

	sql, args, err := ub2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "UPDATE users") {
		t.Errorf("expected UPDATE users, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE, got: %s", sql)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestMutateBuilders_ComplexDelete(t *testing.T) {
	db := DeleteBuilder{
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete("logs"),
	}

	db2 := db.Where(Cond.Lt("created_at", "2023-01-01")).
		Where(Cond.Eq("level", "debug")).
		Limit(1000).
		OrderBy("created_at", ASC)

	sql, args, err := db2.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if !strings.Contains(sql, "DELETE FROM logs") {
		t.Errorf("expected DELETE FROM logs, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") {
		t.Errorf("expected WHERE, got: %s", sql)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}
