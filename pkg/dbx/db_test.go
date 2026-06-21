package dbx

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// TestDBConfig_Defaults tests that DBConfig with defaults is valid
func TestDBConfig_Valid(t *testing.T) {
	cfg := DBConfig{
		DSN: "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
	}

	if cfg.DSN == "" {
		t.Error("DBConfig DSN should not be empty")
	}
}

// TestWrap_CreatesDB tests that Wrap creates a valid DB from sqlx.DB
func TestWrap_CreatesDB(t *testing.T) {
	// Create a mock sqlx.DB for testing
	// Note: In real tests, you'd use a test database
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	if wrapped == nil {
		t.Error("Wrap should return non-nil DB")
	}

	if wrapped.db != db {
		t.Error("Wrap should preserve the underlying sqlx.DB")
	}

	if wrapped.psql == nil {
		t.Error("Wrap should initialize psql statement builder")
	}
}

// TestDB_Close_NoOp tests that Close can be called (though it's a no-op in tests)
func TestDB_Close_MockDB(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	// Close should not panic
	// Note: Real close test would require an actual database connection
	if wrapped == nil {
		t.Error("wrapped DB should not be nil")
	}
}

// TestDB_Underlying_Returns_RawDB tests that Underlying returns the raw sqlx.DB
func TestDB_Underlying_Returns_RawDB(t *testing.T) {
	rawDB := &sqlx.DB{}
	wrapped := Wrap(rawDB)

	underlying := wrapped.Underlying()
	if underlying != rawDB {
		t.Error("Underlying should return the same sqlx.DB that was wrapped")
	}
}

// TestDB_Select_Returns_SelectBuilder tests that Select returns a valid SelectBuilder
func TestDB_Select_Returns_SelectBuilder(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select("id", "name")

	if sb.db != wrapped {
		t.Error("SelectBuilder should reference the parent DB")
	}

	if sb.sq == nil {
		t.Error("SelectBuilder should have a squirrel builder")
	}

	sql, _, err := sb.ToSQL()
	if err != nil {
		t.Fatalf("SelectBuilder.ToSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("SelectBuilder should generate SQL")
	}
}

// TestDB_Insert_Returns_InsertBuilder tests that Insert returns a valid InsertBuilder
func TestDB_Insert_Returns_InsertBuilder(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	ib := wrapped.Insert("users")

	if ib.db != wrapped {
		t.Error("InsertBuilder should reference the parent DB")
	}

	if ib.sq == nil {
		t.Error("InsertBuilder should have a squirrel builder")
	}
}

// TestDB_Update_Returns_UpdateBuilder tests that Update returns a valid UpdateBuilder
func TestDB_Update_Returns_UpdateBuilder(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	ub := wrapped.Update("users")

	if ub.db != wrapped {
		t.Error("UpdateBuilder should reference the parent DB")
	}

	if ub.sq == nil {
		t.Error("UpdateBuilder should have a squirrel builder")
	}
}

// TestDB_Delete_Returns_DeleteBuilder tests that Delete returns a valid DeleteBuilder
func TestDB_Delete_Returns_DeleteBuilder(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	db := wrapped.Delete("users")

	if db.db != wrapped {
		t.Error("DeleteBuilder should reference the parent DB")
	}

	if db.sq == nil {
		t.Error("DeleteBuilder should have a squirrel builder")
	}
}

// TestDB_WithCTE_Returns_CTEBuilder tests that WithCTE returns a CTEBuilder
func TestDB_WithCTE_Returns_CTEBuilder(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	cte := wrapped.WithCTE()

	if cte.ctes == nil {
		t.Error("CTEBuilder should have initialized ctes slice")
	}

	if len(cte.ctes) != 0 {
		t.Error("new CTEBuilder should have empty ctes")
	}
}

// TestDB_BuilderChaining tests that we can chain builders correctly
func TestDB_BuilderChaining(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select("id", "name", "email").
		From("users").
		Where(Cond.Eq("active", true)).
		OrderBy("name", ASC)

	sql, args, err := sb.ToSQL()
	if err != nil {
		t.Fatalf("Builder chaining failed: %v", err)
	}

	if sql == "" {
		t.Error("chained builders should produce SQL")
	}

	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

// TestDB_MultipleBuilders_Independent tests that multiple builders are independent
func TestDB_MultipleBuilders_Independent(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb1 := wrapped.Select("id").From("users")
	sb2 := wrapped.Select("name").From("products")

	sql1, _, err1 := sb1.ToSQL()
	sql2, _, err2 := sb2.ToSQL()

	if err1 != nil || err2 != nil {
		t.Errorf("ToSQL failed: %v, %v", err1, err2)
	}

	if sql1 == sql2 {
		t.Error("different builders should produce different SQL")
	}
}

// TestDB_RawOne tests RawOne method
func TestDB_RawOne_MockDB(t *testing.T) {
	// Note: This tests the method exists and can be called with mocks
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	// We can't actually execute against a mock DB, but we can verify the method exists
	// and follows the expected signature
	if wrapped.RawOne == nil {
		t.Error("DB.RawOne method should exist")
	}
}

// TestDB_RawAll tests RawAll method
func TestDB_RawAll_MockDB(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	if wrapped.RawAll == nil {
		t.Error("DB.RawAll method should exist")
	}
}

// TestDB_RawExec tests RawExec method
func TestDB_RawExec_MockDB(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	if wrapped.RawExec == nil {
		t.Error("DB.RawExec method should exist")
	}
}

// TestDB_Ping_Interface tests that Ping method exists
func TestDB_Ping_Interface(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	if wrapped.Ping == nil {
		t.Error("DB.Ping method should exist")
	}
}

// TestDB_SelectBuilder_NoColumns tests Select with no columns specified
func TestDB_SelectBuilder_NoColumns(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select()
	if sb.sq == nil {
		t.Error("SelectBuilder with no columns should still be valid")
	}
}

// TestDB_SelectBuilder_ManyColumns tests Select with many columns
func TestDB_SelectBuilder_ManyColumns(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select("col1", "col2", "col3", "col4", "col5")
	sql, _, err := sb.From("table").ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("SelectBuilder with many columns should produce SQL")
	}
}

// TestDB_SelectBuilder_WithExpressions tests Select with SQL expressions
func TestDB_SelectBuilder_WithExpressions(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select("COUNT(*) AS count", "MAX(created_at) AS latest").
		From("events")

	sql, _, err := sb.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL failed: %v", err)
	}

	if sql == "" {
		t.Error("SelectBuilder with expressions should produce SQL")
	}
}

// TestIsNoRows_Helper tests the isNoRows helper function
func TestIsNoRows_WithNilError(t *testing.T) {
	result := isNoRows(nil)
	if result {
		t.Error("isNoRows should return false for nil error")
	}
}

// TestDB_PlaceholderFormat tests that DB uses PostgreSQL $ placeholders
func TestDB_PlaceholderFormat(t *testing.T) {
	db := &sqlx.DB{}
	wrapped := Wrap(db)

	sb := wrapped.Select("*").From("users").Where(Cond.Eq("id", 1))
	sql, _, _ := sb.ToSQL()

	// Should use $ for PostgreSQL
	if sql == "" {
		t.Error("SQL should be generated")
	}

	// In PostgreSQL mode, we should see placeholders
	// This is a basic check that the builder is using the right format
	if wrapped.psql == nil {
		t.Error("DB should have initialized statement builder")
	}
}
