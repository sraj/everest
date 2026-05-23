# dbx

**dbx** is a PostgreSQL query builder for Go. It wraps [`sqlx`](https://github.com/jmoiron/sqlx) and [`squirrel`](https://github.com/Masterminds/squirrel) behind a fluent, immutable builder API with first-class support for transactions, pagination, and safe conditions.

## Quick start

```go
import "github.com/sraj/everest/pkg/dbx"

db, err := dbx.New(dbx.DBConfig{
    DSN:             "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 1 * time.Minute,
})
defer db.Close()
```

## SELECT

```go
type User struct {
    ID    string `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

// Single row — returns dbx.ErrNotFound if no match.
var u User
err := db.Select("id", "name", "email").
    From("users").
    Where(dbx.Cond.Eq("id", userID)).
    One(ctx, &u)

// Multiple rows.
var users []User
err := db.Select("*").
    From("users").
    Where(dbx.Cond.Gt("age", 18)).
    OrderBy("name", dbx.ASC).
    All(ctx, &users)

// COUNT(*).
count, err := db.Select("*").
    From("users").
    Where(dbx.Cond.Eq("active", true)).
    Count(ctx)

// EXISTS.
exists, err := db.Select("1").
    From("users").
    Where(dbx.Cond.Eq("email", email)).
    Exists(ctx)
```

### Joins

| Relationship | Example | Key pattern |
| --- | --- | --- |
| **1:M** | Orders → Users → Items (aggregated) | `Join` + `GROUP BY` + aggregation |
| **M:1** | Posts → Authors (author nullable) | `LeftJoin` with `*type` fields |
| **1:M** | Employees → Managers (self-join) | Self `LeftJoin` on same table |
| **1:M** | Categories → LATERAL Transactions | `LeftJoin` with `LATERAL` subquery |
| **M:N** | Articles ↔ Articles (similarity) | Self `Join` with cross-condition |
| **1:1** | Users ↔ Profiles (required) | `Join` (both rows always exist) |
| **1:1** | Users ↔ Profiles (optional) | `LeftJoin` + `*string` for optional side |
| **M:N** | Students → Courses (via enrollments) | Junction table with two `Join` |
| **M:N** | Courses → Student counts (aggregated) | Junction table + `LeftJoin` + aggregate |

```go
type OrderSummary struct {
    OrderID   string  `db:"order_id"`
    UserName  string  `db:"user_name"`
    ItemCount int     `db:"item_count"`
    Total     float64 `db:"total"`
}

// 1:M — one user has many orders, each order has many items.
var summaries []OrderSummary
err := db.Select(
    "o.id AS order_id",
    "u.name AS user_name",
    "COUNT(oi.id) AS item_count",
    "SUM(oi.quantity * oi.unit_price) AS total",
).
    From("orders o").
    Join("users u ON u.id = o.user_id").
    Join("order_items oi ON oi.order_id = o.id").
    Where(dbx.Cond.Eq("o.status", "shipped")).
    Where(dbx.Cond.Gt("o.created_at", since)).
    GroupBy("o.id", "u.name").
    Having(dbx.Cond.Gt("SUM(oi.quantity * oi.unit_price)", 100)).
    OrderBy("total", dbx.DESC).
    All(ctx, &summaries)

// M:1 — many posts belong to one author (author may be deleted).
type BlogPostWithAuthor struct {
    PostID    string  `db:"post_id"`
    Title     string  `db:"title"`
    AuthorID  *string `db:"author_id"`  // nil when author deleted
    AuthorName *string `db:"author_name"`
}
var posts []BlogPostWithAuthor
err := db.Select(
    "p.id AS post_id",
    "p.title",
    "a.id AS author_id",
    "a.name AS author_name",
).
    From("posts p").
    LeftJoin("authors a ON a.id = p.author_id").
    Where(dbx.Cond.Eq("p.published", true)).
    OrderBy("p.published_at", dbx.DESC).
    All(ctx, &posts)

// 1:M — one manager has many employees (self-join).
type EmployeeWithManager struct {
    EmployeeID   string `db:"employee_id"`
    EmployeeName string `db:"employee_name"`
    ManagerID    string `db:"manager_id"`
    ManagerName  string `db:"manager_name"`
}
var emps []EmployeeWithManager
err := db.Select(
    "e.id AS employee_id",
    "e.name AS employee_name",
    "m.id AS manager_id",
    "m.name AS manager_name",
).
    From("employees e").
    LeftJoin("employees m ON m.id = e.manager_id").
    OrderBy("e.name", dbx.ASC).
    All(ctx, &emps)

// 1:M — one category has many transactions (lateral subquery).
type DashboardRow struct {
    Category string  `db:"category"`
    Revenue  float64 `db:"revenue"`
    Count    int     `db:"count"`
}
var rows []DashboardRow
err := db.Select("c.name AS category", "s.revenue", "s.count").
    From("categories c").
    LeftJoin("LATERAL (SELECT SUM(amount) AS revenue, COUNT(*) AS count "+
        "FROM transactions t WHERE t.category_id = c.id AND t.status = 'completed'"+
        ") s ON true").
    Where(dbx.Cond.Gt("s.revenue", 0)).
    OrderBy("s.revenue", dbx.DESC).
    All(ctx, &rows)

// M:N — cross-matching articles by similarity score (self-referencing M:N).
var refs []CrossRef
err := db.Select("a.id AS left_id", "b.id AS right_id", "similarity(a.title, b.title) AS score").
    From("articles a").
    Join("articles b ON a.id < b.id").
    Where(dbx.Cond.Gt("similarity(a.title, b.title)", 0.8)).
    OrderBy("score", dbx.DESC).
    Paginate(dbx.Page{Size: 50}).
    All(ctx, &refs)

// 1:1 — one user has exactly one profile.
type UserWithProfile struct {
    UserID    string `db:"user_id"`
    UserName  string `db:"user_name"`
    AvatarURL string `db:"avatar_url"`
    Bio       string `db:"bio"`
}
var profiles []UserWithProfile
err := db.Select(
    "u.id AS user_id",
    "u.name AS user_name",
    "p.avatar_url",
    "p.bio",
).
    From("users u").
    Join("user_profiles p ON p.user_id = u.id").
    OrderBy("u.name", dbx.ASC).
    All(ctx, &profiles)

// 1:1 with optional profile (LEFT JOIN, profile fields are nilable).
type UserOptionalProfile struct {
    UserID    string  `db:"user_id"`
    UserName  string  `db:"user_name"`
    AvatarURL *string `db:"avatar_url"`
    Bio       *string `db:"bio"`
}
var users []UserOptionalProfile
err := db.Select(
    "u.id AS user_id",
    "u.name AS user_name",
    "p.avatar_url",
    "p.bio",
).
    From("users u").
    LeftJoin("user_profiles p ON p.user_id = u.id").
    OrderBy("u.name", dbx.ASC).
    All(ctx, &users)

// M:N — students enrolled in many courses via junction table.
type EnrolledStudent struct {
    StudentID   string `db:"student_id"`
    StudentName string `db:"student_name"`
    CourseID    string `db:"course_id"`
    CourseName  string `db:"course_name"`
    EnrolledAt  string `db:"enrolled_at"`
}
var enrollments []EnrolledStudent
err := db.Select(
    "s.id AS student_id",
    "s.name AS student_name",
    "c.id AS course_id",
    "c.name AS course_name",
    "e.created_at AS enrolled_at",
).
    From("enrollments e").
    Join("students s ON s.id = e.student_id").
    Join("courses c ON c.id = e.course_id").
    Where(dbx.Cond.Eq("e.status", "active")).
    OrderBy("e.created_at", dbx.DESC).
    All(ctx, &enrollments)

// M:N — aggregated: count of enrolled students per course.
type CourseEnrollmentCount struct {
    CourseID      string `db:"course_id"`
    CourseName    string `db:"course_name"`
    StudentCount  int    `db:"student_count"`
}
var courseCounts []CourseEnrollmentCount
err := db.Select(
    "c.id AS course_id",
    "c.name AS course_name",
    "COUNT(e.student_id) AS student_count",
).
    From("courses c").
    LeftJoin("enrollments e ON e.course_id = c.id AND e.status = 'active'").
    GroupBy("c.id", "c.name").
    OrderBy("student_count", dbx.DESC).
    All(ctx, &courseCounts)
```

## INSERT

```go
// Standard insert.
db.Insert("users").
    Columns("id", "name", "email").
    Values(uuid.New().String(), "Alice", "alice@example.com").
    Exec(ctx)

// With RETURNING.
var u User
db.Insert("users").
    Columns("id", "name", "email").
    Values(uuid.New().String(), "Alice", "alice@example.com").
    Returning("id", "name", "email").
    One(ctx, &u)

// SetMap with ON CONFLICT.
db.Insert("users").
    SetMap(map[string]any{"id": id, "name": "Alice", "email": email}).
    OnConflict("(email) DO UPDATE SET name = EXCLUDED.name").
    Exec(ctx)
```

## UPDATE

```go
db.Update("users").
    Set("name", "Bob").
    Set("updated_at", time.Now()).
    Where(dbx.Cond.Eq("id", userID)).
    Exec(ctx)

// Fail if no row matched.
err := db.Update("users").
    Set("name", "Bob").
    Where(dbx.Cond.Eq("id", userID)).
    ExecMustAffect(ctx)  // returns dbx.ErrNoRows if 0 rows updated

// With RETURNING.
var u User
db.Update("users").
    Set("name", "Bob").
    Where(dbx.Cond.Eq("id", userID)).
    Returning("id", "name").
    One(ctx, &u)

// Raw SQL expression.
db.Update("products").
    SetExpr("price", "price * ?", 1.10).
    Where(dbx.Cond.Eq("category", "electronics")).
    Exec(ctx)

// Conditional where.
db.Update("users").
    Set("name", name).
    Where(dbx.Cond.Eq("id", id)).
    WhereIf(role != "", dbx.Cond.Eq("role", role)).
    Exec(ctx)
```

## DELETE

```go
db.Delete("users").
    Where(dbx.Cond.Eq("id", userID)).
    Exec(ctx)

db.Delete("users").
    Where(dbx.Cond.Eq("id", userID)).
    ExecMustAffect(ctx)  // returns dbx.ErrNoRows if 0 rows deleted

db.Delete("sessions").
    Where(dbx.Cond.Lt("expires_at", time.Now())).
    Returning("id").
    Exec(ctx)
```

## Conditions

```go
dbx.Cond.Eq("col", val)         // col = $1
dbx.Cond.NotEq("col", val)      // col != $1
dbx.Cond.Gt("col", val)         // col > $1
dbx.Cond.Lt("col", val)         // col < $1
dbx.Cond.GtOrEq("col", val)     // col >= $1
dbx.Cond.LtOrEq("col", val)     // col <= $1
dbx.Cond.In("col", 1, 2, 3)     // col IN ($1, $2, $3)
dbx.Cond.Like("col", "%foo%")   // col LIKE $1
dbx.Cond.ILike("col", "%foo%")  // col ILIKE $1
dbx.Cond.IsNull("col")          // col IS NULL
dbx.Cond.IsNotNull("col")       // col IS NOT NULL
dbx.Cond.Between("col", lo, hi) // col >= $1 AND col <= $2
dbx.Cond.Search("term", "name", "email")  // (name ILIKE '%term%' OR email ILIKE '%term%')
dbx.Cond.Raw("col @> ?", `{"tag"}`)       // raw SQL expression

// Compound conditions.
dbx.Cond.And(
    dbx.Cond.Eq("active", true),
    dbx.Cond.Gt("age", 18),
)
dbx.Cond.Or(
    dbx.Cond.Eq("role", "admin"),
    dbx.Cond.Eq("role", "moderator"),
)
```

## Pagination

```go
type Post struct {
    ID      string `db:"id"`
    Title   string `db:"title"`
}

// Paginate returns total count, items, page metadata.
result, err := dbx.Paginate[Post](ctx,
    db.Select("id", "title").From("posts"),
    dbx.Page{Number: 1, Size: 20},
)
// result.Total, result.Items, result.Page, result.PageSize, result.TotalPages
```

## Transactions

```go
err := db.Tx(ctx, func(tx *dbx.TxDB) error {
    // All builder methods available on tx.
    tx.Update("accounts").
        Set("balance", sq.Expr("balance - ?", 100)).
        Where(dbx.Cond.Eq("id", fromID)).
        ExecMustAffect(ctx)

    tx.Update("accounts").
        Set("balance", sq.Expr("balance + ?", 100)).
        Where(dbx.Cond.Eq("id", toID)).
        ExecMustAffect(ctx)

    return nil // commits
})
// Auto-rollback on error or panic.
```

## Raw SQL (escape hatch)

For queries squirrel can't express (CTEs, window functions, etc.):

```go
var count int
err := db.RawOne(ctx, &count, "SELECT COUNT(*) FROM users WHERE active = $1", true)

var users []User
err := db.RawAll(ctx, &users, "SELECT * FROM users WHERE created_at > $1", since)

n, err := db.RawExec(ctx, "DELETE FROM sessions WHERE expires_at < $1", time.Now())
```

The same escape hatches exist on `TxDB`.

## Test helpers

```go
// Wrap an existing *sqlx.DB (e.g., from a test container).
db := dbx.Wrap(sqlxDB)

// Inspect generated SQL.
sql, args, _ := db.Select("id").From("users").Where(dbx.Cond.Eq("id", 1)).ToSQL()
// sql:  "SELECT id FROM users WHERE id = $1"
// args: []any{1}
```

## Sentinels

| Error | When returned |
|---|---|
| `dbx.ErrNotFound` | `One()` finds no matching row |
| `dbx.ErrNoRows` | `ExecMustAffect()` affects zero rows |

## Conventions

- All builders are **immutable** — every chain method returns a new copy. Fork freely.
- Placeholders use PostgreSQL `$N` syntax.
- `Page` is 1-based. Default page size is 20.
- Use `WhereIf(cond, pred)` to build dynamic filters without `if` blocks.
- `AllowSort(cols...)` whitelists sort columns to prevent SQL injection through user-controlled `OrderBy`.

## Production readiness

The following improvements would make `dbx` suitable for large-scale enterprise use:

### Observability
- **OpenTelemetry instrumentation** — automatic span creation around every query (build + exec) with SQL, args, and duration as attributes
- **Metrics** — expose pool stats (`MaxOpen`, `InUse`, `Idle`, `WaitCount`, `WaitDuration`) and per-query histograms via `prometheus` or `otel`
- **Structured query logging** — optional slog hook that logs slow queries (>threshold) or all queries at debug level
- **Tracing context propagation** — extract `trace_id`/`span_id` from context and inject into database session for PostgreSQL `pg_stat_activity` correlation

### Resiliency
- **Circuit breaker** — wrap pool methods; trip on connection timeouts or `ErrConBadConn`; half-open probe with `Ping`
- **Retry with backoff** — configurable retry for serialization failures (`40001`/`40P01`) and transient network errors; exponential backoff with jitter
- **Connection health reactor** — goroutine that periodically pings idle connections; evicts and reconnects stale ones
- **Read-write splitting** — primary (`*DB`) + replicas (`*DB`) with round-robin or latency-aware routing; writes and `RETURNING` forced to primary

### Schema & migrations
- **Embedded migrator** — `dbx.Migrate(ctx, fs.Sub(migrations))` that runs `golang-migrate` under the hood; expose status, up, down, and version bump as methods on `DB`
- **Schema introspection** — `Table(name)` helper that returns columns, types, nullable, and foreign keys for codegen and dynamic queries

### Query features
- **Bulk insert** — `Insert("table").Values(row1, row2, ...)` that batches into chunks of 500 with `COPY` fallback for >1000 rows
- **Upsert** — first-class `OnConflict` builder (not a `Suffix` hack) with `DoNothing()` and `DoUpdate(setMap)` helpers
- **CTE builder** — `With("cte_name", subQuery)` that prefixes the main statement with `WITH …`; chainable, composable
- **Window function helpers** — `Over(partition, order)` on aggregate columns
- **JSONB operators** — `Cond.JsonContains`, `Cond.JsonPath`, `Cond.HasKey` that generate `@>`, `#>>`, `?` etc.
- **Array operators** — `Cond.Overlaps`, `Cond.Contains` for PostgreSQL array `&&` and `@>` operators

### Advanced features
- **Soft delete** — `SoftDelete() SelectBuilder` that injects `WHERE deleted_at IS NULL`; `Deleted()` shows soft-deleted rows
- **Optimistic locking** — `Where(dbx.Cond.Eq("version", version))` on update with version auto-increment; return `ErrConflict` on stale write
- **Audit columns** — optional auto-population of `created_at`, `updated_at`, `created_by`, `updated_by` on insert/update
- **Multi-tenant isolation** — `WithTenant(tenantID)` that injects `AND tenant_id = $N` on every query via a wrapping builder
- **Encrypted columns** — transparent `SetEncrypted(col, plaintext, key)` / `DecryptedColumn(col, key)` for pii columns at rest
- **Batch processing** — `Each()` cursor that yields rows one-by-one with a configurable fetch-ahead buffer (avoids loading entire result set)

### Performance
- **Prepared statement caching** — expose `sqlx.Prepare` / `PrepareNamed` through the builder; cache parsed statements per connection
- **Connection pool tuning** — export pool configuration as live-settable fields; expose `Stats()` for monitoring integration
- **Result streaming** — `Stream(ctx) chan Row` that reads rows from a cursor and pushes them through a channel; cancel via context

### Security
- **Comprehensive sort whitelist** — `AllowSort` is a good start; expand to reject `;`, subqueries, and function calls in `OrderBy` column names
- **Row-level security** — helpers to set `app.current_tenant` and `app.current_user` via `SET LOCAL` in each session or transaction

## Database support

Currently hardcoded to PostgreSQL via `github.com/lib/pq` with dollar `$N` placeholders. With the following changes, `dbx` could support multiple databases:

| Database | Driver | Placeholder | What needs to change |
|---|---|---|---|
| **PostgreSQL** | `lib/pq` or `pgx` | `$N` | Already supported |
| **MySQL / MariaDB** | `go-sql-driver/mysql` | `?` | Swap `sq.Dollar` → `sq.Question`; remove `ILIKE` (use `LIKE` with case-insensitive collation), `RETURNING` (use `LAST_INSERT_ID`), `ON CONFLICT` (use `ON DUPLICATE KEY`), `@>` / `?` operators |
| **SQLite** | `mattn/go-sqlite3` or `modernc.org/sqlite` | `?` | Same placeholder change as MySQL; `ILIKE` works in SQLite 3.44+ but use `LOWER()` for older versions; no `RETURNING` before 3.35 (use `last_insert_rowid()`); no `ON CONFLICT` before 3.24 |
| **SQL Server** | `denisenkom/go-mssqldb` | `@p1` format with `INFORMATION_SCHEMA` | Use `sq.Dollar` (squirrel maps `$N` → `@pN` for `mssql`); replace `ILIKE` with `LOWER(col) LIKE LOWER(pattern)`; `RETURNING` → `OUTPUT INSERTED.*`; remove `ON CONFLICT` (use `MERGE`); no `ISNULL`/`@>` equivalents |
| **CockroachDB** | `lib/pq` or `pgx` | `$N` | Nearly identical to PostgreSQL; `RETURNING` and `ON CONFLICT` work; remove `pg_advisory_lock` and `LISTEN/NOTIFY` if those features are added later |
| **YugabyteDB** | `lib/pq` or `pgx` | `$N` | PostgreSQL-compatible wire protocol; same as PostgreSQL with minor distributed SQL considerations |

### Architecture for multi-DB support

```
dbx.New(cfg)          → chooses driver + placeholder based on cfg.Driver
dbx.Cond.ILIKE(...)   → compiles to ILIKE on PG, LOWER() on others (or omitted via build tags)
Insert.Returning(...) → compiled to RETURNING or OUTPUT INSERTED or skipped with fallback
Update.OnConflict(...)→ compiled to ON CONFLICT / ON DUPLICATE KEY / MERGE
```

A `dbx.Driver` interface with per-dialect implementations (`PostgresDialect`, `MySQLDialect`, `SQLiteDialect`, `MSSQLDialect`) would encapsulate:

- Placeholder format (`sq.Dollar`, `sq.Question`, `sq.AtP`)
- Feature flags (`supportsReturning`, `supportsOnConflict`, `supportsILike`, `supportsCte`, `supportsWindowFunctions`)
- Type mapping (e.g., `TEXT` vs `VARCHAR` vs `NVARCHAR`)
- Error translation (`isNoRows` → `ErrNotFound` mapping per driver)

This would let callers write the same `dbx` code and target any database by changing only the connection string and driver name, similar to the `database/sql` pattern.

## Dependencies

- [`github.com/jmoiron/sqlx`](https://github.com/jmoiron/sqlx)
- [`github.com/Masterminds/squirrel`](https://github.com/Masterminds/squirrel)
- [`github.com/lib/pq`](https://github.com/lib/pq) (PostgreSQL driver)
