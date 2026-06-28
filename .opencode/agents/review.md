# Everest Code Review Checklist

Every PR must pass these checks. Reviewers: confirm each item.

---

## Package Dependencies

- [ ] **No `pkg/` imports `internal/`** — `pkg/` packages are extractable. Verify with:
  ```bash
  grep -r "internal/" pkg/ --include="*.go" | grep -v "_test.go"
  ```
  Should produce no output.

- [ ] **Layer isolation — no reverse dependencies**:
  ```bash
  # Handlers must NOT import store/ or datastore/ directly
  grep -r "internal/store\|internal/datastore" internal/handler/ --include="*.go"
  
  # Services must NOT import datastore/ or handler/
  grep -r "internal/datastore\|internal/handler" internal/service/ --include="*.go"
  
  # domain/model must NOT import anything from internal/ except model itself
  grep -r "internal/" internal/domain/model/ --include="*.go" | grep -v "model"
  ```
  All should produce no output.

- [ ] **No handler calls store directly** — handlers call services, never `store.Store()`.
- [ ] **No domain/model imports other internal packages** — model types are dependency-free.
- [ ] **Constructors return interfaces** for injectable dependencies (services, stores). Concrete types are acceptable for transport handlers and infrastructure pools.

---

## Error Handling

- [ ] **Repo returns typed sentinel errors**, never raw `sql.ErrNoRows` or `*pq.Error`:
  ```go
  // REQUIRED
  return store.ErrNotFound{Resource: "document", ID: id}
  return store.ErrConflict{Resource: "document", Field: "title", Err: err}

  // FORBIDDEN
  return err  // leaks DB error to service layer
  ```

- [ ] **Service translates sentinels to `apperror`** using `errors.As`:
  ```go
  var nfErr store.ErrNotFound
  if errors.As(err, &nfErr) {
      return apperror.NotFound(...)
  }
  ```

- [ ] **DB constraint violations are detected** using `dbx.IsUniqueViolation()` / `dbx.IsForeignKeyError()`:
  ```go
  if dbx.IsUniqueViolation(err) {
      return store.ErrConflict{...}
  }
  ```

- [ ] **All `slog` error calls use `err.Error()`** not raw `err`:
  ```bash
  grep -r '"error", err\b' internal/ --include="*.go"
  ```
  Should produce no output (AppError has `json:"-"` fields, raw err prints `{}`).

- [ ] **Errors are wrapped with context** at every layer:
  ```go
  return fmt.Errorf("get document: %w", err)  // GOOD
  return err                                   // BAD — loses context
  ```

- [ ] **No panic** outside `main.go` startup failures. Always return errors.

---

## Request Binding + Validation

- [ ] **All request bodies use typed structs** with validation tags:
  ```go
  type XRequest struct {
      Field string `json:"field" validate:"required,max=100"`
  }
  ```

- [ ] **Binding uses `bindBody()` / `bindQuery()`** — never manual `c.BodyParser()` + inline if-checks.

- [ ] **Pointer fields for optional updates** — `*string` so nil means "not provided":
  ```go
  type UpdateXRequest struct {
      Title *string `json:"title" validate:"omitempty,max=500"`
  }
  ```

---

## API Design

- [ ] **Response types are separate from domain models** — defined in `response.go`, not handler body.
- [ ] **Conversion uses `toXResponse()` functions** — no inline struct literals in handlers.
- [ ] **Proto file updated** if adding/modifying endpoints — check `api/proto/`.
- [ ] **`make proto` ran** and generated files committed — check `api/gen/` for changes.
- [ ] **OpenAPI spec committed** — `api/gen/openapiv2/api.swagger.json` matches proto annotations.
- [ ] **New route registered** in `handler.go` → `RegisterRoutes()`.

---

## Security

- [ ] **No secrets in code** — config values only, never hardcoded credentials.
- [ ] **OwnerID from auth context** — never hardcoded `"00000000-0000-0000-0000-000000000001"` in new features.
- [ ] **Rate limiting applies** to new endpoints — check `RateLimitMax` in `HTTPConfig`.
- [ ] **Request timeout applies** — check `RequestTimeout` in `HTTPConfig`.
- [ ] **Input validation on all user-supplied values** — no blind `fmt.Sprintf` with untrusted input.

---

## Testing

- [ ] **Service tests use mock `store.Store`** — not individual mock repos:
  ```go
  st := store.New(docStore, contentStore, noopClose, noopPing)
  svc := NewDocumentService(st, nil, log)
  ```

- [ ] **Happy path + error cases covered** for every new handler/service method.
- [ ] **Validation errors tested** — missing fields, invalid values, boundary cases.
- [ ] **No skipped tests** without `t.Skip("reason")` and a documented reason.
- [ ] **No test depends on external services** — no real DB, no real MinIO, no real Chrome.

---

## Database

- [ ] **Migrations have both `.up.sql` and `.down.sql`**.
- [ ] **Migrations are additive** — never modify existing migration files.
- [ ] **New store methods use `dbx` builders** — no raw SQL strings (use `r.db.Select()`, `r.db.Insert()`, etc.).
- [ ] **Pagination uses `dbx.Page` + `Paginate()`** — no manual `LIMIT/OFFSET` strings.
- [ ] **COUNT queries use `r.db.Select().From().Count()`** — builder method, not raw SQL.

---

## Performance

- [ ] **DB queries use specific columns** — never `SELECT *`.
- [ ] **List endpoints are paginated** — no unbounded results.
- [ ] **No N+1 queries** — batch-load related data.
- [ ] **Thumbnail generation is async** — never blocks the HTTP response.
- [ ] **Long-running operations use goroutines + context deadlines**.

---

## General

- [ ] **`make build` passes** — no compilation errors.
- [ ] **`make test` passes** — all tests green.
- [ ] **No dead code** — removed unused functions, types, imports.
- [ ] **No commented-out code** — delete it or add a TODO with a reason.
- [ ] **Commit message follows Conventional Commits** — `feat:`, `fix:`, `refactor:`, `test:`, `docs:`.
- [ ] **No merge conflicts** with `main`.
