# Everest Development Guide

Every developer working on this project must follow these conventions.

---

## Project Structure

```
cmd/                          # Binary entrypoints (main.go only)
  server/                     #   Everest server
  migrate/                    #   Database migration CLI
internal/                     # Private application code (not importable externally)
  config/                     #   Environment-based configuration
  store/                      #   Data access interfaces (Store, DocumentStore, ContentStore)
  domain/model/               #   Domain entities + value objects (Document, User, Page, PageResult)
  datastore/                  #   Data store implementations
    postgres/                 #     PostgreSQL-backed stores
    minio/                    #     MinIO-backed content stores
    zitadel/                  #     Zitadel auth integration (zitadel_authentication branch only)
  service/                    #   Business logic (DocumentService, ThumbnailService)
  handler/                    #   Transport layer
    http/                     #     REST handlers (Fiber)
    grpc/                     #     gRPC handlers
  apperror/                   #   Application error types
  version/                    #   Embedded version info (ldflags)
api/                          # API contracts
  proto/                      #   Protobuf definitions
  buf.yaml                    #   Buf module config
  buf.gen.yaml                #   Code generation config (Go + gRPC + OpenAPI)
  gen/                        #   Generated code (committed, not gitignored)
db/                           # Database state
  migrations/                 #   Schema migrations
  seeds/                      #   Seed data
pkg/                          # Reusable libraries (ZERO internal/ imports)
  bind/                       #   Request binding + validation
  configx/                    #   Typed env config loader
  dbx/                        #   PostgreSQL query builder (wraps squirrel + sqlx)
  logger/                     #   Structured logging
  server/                     #   HTTP + gRPC server lifecycle
web/                          # React frontend (Vite + TypeScript)
```

## Code Organization Rules

### Package dependency direction
```
handler → service → store ← infrastructure
    ↓                   ↓
 apperror              domain/model
```
- **Handlers** call services. Never call store directly.
- **Services** call store accessors (`s.store.Document().GetByID(...)`).
- **Store** defines interfaces. Infrastructure packages implement them.
- **pkg/** packages must NEVER import `internal/`. They are extractable libraries.
- **domain/model/** has zero dependencies on any other internal package.

### Store pattern
- `store.Store` is the single data access entry point.
- Sub-accessors: `store.Document()`, `store.Content()`.
- Constructor: `store.New(doc, content, closer, pinger)`.
- Services inject `store.Store`, not individual repos.

```go
// main.go — wiring
docStore := postgres.NewDocumentStore(db)
contentStore, _ := minio.NewContentStore(cfg)
st := store.New(docStore, contentStore, db.Close, db.Ping)
docService := service.NewDocumentService(st, thumbnailSvc, log)
```

### Adding a new feature — the flow

1. **Define the API contract**: Add RPC + HTTP annotations to `.proto` file
2. **Generate code**: `make proto`
3. **Define domain model**: Add entity/value object in `internal/domain/model/`
4. **Define store interface**: Add substore interface in `internal/store/`
5. **Implement store**: Add implementation in `internal/infrastructure/postgres/`
6. **Write service**: Business logic in `internal/service/`, inject `store.Store`
7. **Add HTTP handler**: Handler calls service, uses `bindBody()`/`bindQuery()` for input
8. **Add gRPC handler**: Implements generated server interface
9. **Write tests**: Mock-based unit tests for service, httptest for HTTP handlers

---

## Error Handling

### Three-layer error system

```
Infrastructure (postgres) → Store sentinel (store) → Service (apperror) → Handler
```

**Layer 1 — Infrastructure**: Return typed domain errors, never raw DB errors.
```go
// GOOD
if errors.Is(err, dbx.ErrNotFound) {
    return nil, store.ErrNotFound{Resource: "document", ID: id}
}

// BAD
return nil, err  // raw sql.ErrNoRows leaks to callers
```

**Layer 2 — Service**: Translate store errors to `apperror` using `errors.As`.
```go
func (s *documentService) translateError(err error, id string) error {
    var nfErr store.ErrNotFound
    if errors.As(err, &nfErr) {
        return apperror.NotFound("document %s not found", nfErr.ID)
    }
    var cfErr store.ErrConflict
    if errors.As(err, &cfErr) {
        return apperror.Conflict("document %s: conflict", cfErr.Field)
    }
    return err
}
```

**Layer 3 — Handler**: Propagate service errors. The error handler middleware auto-detects `*apperror.AppError` and sets HTTP status.

### AppError types
```go
apperror.BadRequest(...)     // 400
apperror.Unauthorized(...)   // 401
apperror.Forbidden(...)      // 403
apperror.NotFound(...)       // 404
apperror.Conflict(...)       // 409
apperror.ValidationError(err) // 422 (use with struct-tag validation)
apperror.Internal(...)       // 500
```

### Logging errors
Always pass `err.Error()` not raw `err` to `slog`:
```go
// GOOD
s.log.Error("failed to create document", "error", err.Error())

// BAD
s.log.Error("failed to create document", "error", err)  // prints { } for AppError
```

---

## Request Binding + Validation

### Define request types with struct tags
```go
// request.go
type CreateDocumentRequest struct {
    Title   string `json:"title" validate:"max=500"`
    Content string `json:"content"`
}
```

### Bind + validate in one call
```go
// handler
var req CreateDocumentRequest
if err := bindBody(c, &req); err != nil {
    return err  // 400 for parse errors, 422 for validation errors
}
```

### Bind query params
```go
type ListDocumentsQuery struct {
    Page int `query:"page" validate:"omitempty,gte=1"`
    Size int `query:"size" validate:"omitempty,gte=1,lte=100"`
}

var q ListDocumentsQuery
if err := bindQuery(c, &q); err != nil {
    return err
}
```

---

## API Response Types

### Separate from domain models
- Domain models in `internal/domain/model/`
- API response types in `internal/handler/http/response.go`
- Conversion via `toXResponse()` functions

```go
// GOOD — API types decoupled from domain
type DocumentResponse struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    CreatedAt string `json:"created_at"`
}

func toDocumentResponse(doc *model.Document) DocumentResponse { ... }

// BAD — returning domain model directly (couples API to DB schema)
c.JSON(doc)
```

---

## Background Jobs

Use `jobs.Pool` for all async work. Never use raw `go func()` in services.

```go
// GOOD — uses worker pool with graceful shutdown and retries
s.pool.Submit(func(ctx context.Context) error {
    return s.generateAndSaveThumbnail(ctx, docID, content)
})

// BAD — raw goroutine, no lifecycle control, no retries
go s.generateAndSaveThumbnail(context.Background(), docID, content)
```

### Pool configuration
```go
pool := jobs.New(jobs.Config{
    Workers:     4,     // concurrent goroutines
    QueueSize:   100,   // buffered jobs before backpressure
    MaxAttempts: 2,     // retry count for failed 5xx jobs (0 = no retry)
    Log:         log,
})
defer pool.Shutdown(30 * time.Second)  // drain queue before exit
```

### Retry behavior
- **Retries** only `5xx` errors — client errors (`4xx`) are never retried.
- **Backoff**: 500ms, then 1s between attempts.
- **Max attempts**: configurable, defaults to 2 retries (3 total attempts).
- **Job timeout**: each job gets a 120-second context deadline.

### Adding a new background job
1. Inject `*jobs.Pool` into the service constructor.
2. Call `s.pool.Submit(func(ctx context.Context) error { ... })` in the service method.
3. Always respect the `ctx` — check `ctx.Err()` for cancellation in long-running work.

---

## Testing

### Service tests
- Mock the `store.Store` interface. Do NOT mock individual repos.
- Co-located: `document_test.go` beside `document.go`
- Use `testify/assert` and `testify/require`

```go
docStore := &mockDocumentStore{...}
contentStore := &mockContentStore{...}
st := store.New(docStore, contentStore, noopClose, noopPing)
svc := NewDocumentService(st, nil, testLogger)
```

### HTTP handler tests
- Use Fiber's `app.Test()` for route testing
- Use `pkg/server/server_test.go` for server lifecycle tests

### What needs tests
- Every service method
- Every HTTP handler (happy path + error cases)
- `pkg/` packages
- Config validation

---

## Proto + API Generation

```bash
# After modifying .proto files:
make proto

# This generates:
#   api/gen/go/           — Go + gRPC stubs
#   api/gen/openapiv2/    — OpenAPI 2.0 spec
```

- Generated files ARE committed to the repo.
- `buf dep update` runs automatically as part of `make proto`.
- OpenAPI spec served at `GET /api/docs/openapi.json`.

---

## Branches

| Branch | Purpose |
|---|---|
| `main` | Production code, no auth |
| `zitadel_authentication` | Zitadel OIDC auth |
| Feature branches | Named `feature/description` |

**Sync policy**: `main → zitadel_authentication` only. Never reverse. Feature branches merge to `main`.

---

## Build & Run

```bash
make install-tools    # one-time: install go, air, buf, golangci-lint
make docker-infra     # start PostgreSQL + MinIO
make migrate-up       # apply schema
make seed             # seed default data
make dev              # hot-reload server
make web-dev          # hot-reload frontend (separate terminal)
make build            # production build
make test             # run all tests
make proto            # generate from .proto files
```

---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):
```
feat: add document sharing via permissions
fix: handle nil thumbnail_id from database scan
refactor: extract store pattern into single Store interface
test: add unit tests for document service
docs: update README with local dev setup
```
