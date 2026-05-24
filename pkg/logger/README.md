# logger

**logger** provides a structured logger backed by [`zerolog`](https://github.com/rs/zerolog) exposed through the standard library's [`log/slog`](https://pkg.go.dev/log/slog) API, with a pretty-printed console output, JSON output, and a Fiber HTTP middleware.

## Quick start

```go
import "github.com/sraj/everest/pkg/logger"

log := logger.New("debug", "everest")
log.Info("server starting", "port", 8080)
```

Console output:

```
2026-05-23T20:00:00Z INF server starting port=8080 service=everest
```

## JSON output

For production, use `NewJSON` to get structured JSON lines parsable by log aggregators (Datadog, Grafana Loki, ELK, etc.):

```go
log := logger.NewJSON("info", "everest")
log.Info("server starting", "port", 8080)
```

Output:

```json
{"level":"info","service":"everest","time":"2026-05-23T20:00:00Z","message":"server starting","port":8080}
```

## Console vs JSON

| Output | Function | Use case |
|---|---|---|
| Pretty console (colorized) | `logger.New(level, service)` | Local development |
| JSON (structured) | `logger.NewJSON(level, service)` | Production / log aggregators |

## Console output

```go
log := logger.New("debug", "everest")

log.Debug("connecting to database", "host", "localhost", "port", 5432)
log.Info("server started", "port", 8080)
log.Warn("slow query detected", "duration", 250*time.Millisecond)
log.Error("failed to connect", "error", err)
```

Console output:

```
2026-05-23T20:00:00Z DBG connecting to database host=localhost port=5432 service=everest
2026-05-23T20:00:00Z INF server started port=8080 service=everest
2026-05-23T20:00:00Z WRN slow query detected duration=250ms service=everest
2026-05-23T20:00:00Z ERR failed to connect error="connection refused" service=everest
```

## JSON output

```go
log := logger.NewJSON("debug", "everest")

log.Debug("connecting to database", "host", "localhost", "port", 5432)
log.Info("server started", "port", 8080)
log.Warn("slow query detected", "duration", 250*time.Millisecond)
log.Error("failed to connect", "error", err)
```

JSON output:

```json
{"level":"debug","service":"everest","time":"2026-05-23T20:00:00Z","message":"connecting to database","host":"localhost","port":5432}
{"level":"info","service":"everest","time":"2026-05-23T20:00:00Z","message":"server started","port":8080}
{"level":"warn","service":"everest","time":"2026-05-23T20:00:00Z","message":"slow query detected","duration":250000000}
{"level":"error","service":"everest","time":"2026-05-23T20:00:00Z","message":"failed to connect","error":"connection refused"}
```

## Log levels

Available level strings: `"debug"`, `"info"`, `"warn"`, `"error"`. Unknown strings default to `Info`.

## Attribute types

The handler supports all standard `slog` attribute kinds:

```go
log.Info("order created",
    "order_id", "ord_abc123",              // string
    "total", int64(4999),                  // int64
    "quantity", uint64(2),                 // uint64
    "discount", 0.15,                      // float64
    "express", true,                       // bool
    "eta", 30*time.Minute,                 // duration
    "shipped_at", time.Now(),              // time
    "meta", map[string]string{"note": "gift wrap"}, // any
)
```

JSON output:

```json
{"level":"info","service":"everest","time":"...","message":"order created","order_id":"ord_abc123","total":4999,"quantity":2,"discount":0.15,"express":true,"eta":1800000000000,"shipped_at":"...","meta":{"note":"gift wrap"}}
```

## Error logging

```go
if err != nil {
    log.Error("database query failed",
        "query", "SELECT * FROM orders",
        "error", err,
        "retry", 3,
    )
}
```

JSON output:

```json
{"level":"error","service":"everest","time":"...","message":"database query failed","query":"SELECT * FROM orders","error":"connection refused","retry":3}
```

## Contextual logging with `With`

```go
reqLog := log.With("request_id", "req_001", "user_id", "usr_42")
reqLog.Info("processing payment")
reqLog.Warn("payment declined", "reason", "insufficient_funds")
```

Both log lines include `request_id=req_001` and `user_id=usr_42`.

## Logging errors with stack traces

```go
err := doSomething()
if err != nil {
    log.Error("operation failed",
        "error", err,
        "stack", fmt.Sprintf("%+v", errors.WithStack(err)),
    )
}
```

## Switching output per environment

```go
var log *slog.Logger
if os.Getenv("APP_ENV") == "production" {
    log = logger.NewJSON(cfg.LogLevel, cfg.AppName)
} else {
    log = logger.New(cfg.LogLevel, cfg.AppName)
}
```

## Disable colors in console output

```go
output := zerolog.ConsoleWriter{
    Out:        os.Stdout,
    TimeFormat: time.RFC3339,
    NoColor:    true,
}

// To use a custom writer, modify logger.New or build the logger manually.
```

## Fiber middleware

```go
import "github.com/sraj/everest/pkg/logger"

app := fiber.New()
log := logger.New("debug", "everest")
app.Use(logger.Middleware(log))
```

Console output:

```
2026-05-23T20:00:00Z INF request method=GET path=/api/v1/documents status=200 ip=127.0.0.1 request_id=abc latency=12ms service=everest
```

JSON output (with `NewJSON`):

```json
{"level":"info","service":"everest","time":"...","message":"request","method":"GET","path":"/api/v1/documents","status":200,"ip":"127.0.0.1","request_id":"abc","latency":12000000}
```

## Full Fiber app

```go
cfg := config.Load()

log := logger.New(cfg.LogLevel, cfg.AppName)

app := fiber.New(fiber.Config{AppName: cfg.AppName})
app.Use(logger.Middleware(log))
app.Use(recover.New())

app.Get("/health", func(c *fiber.Ctx) error {
    log.Debug("health check pinged", "ip", c.IP())
    return c.SendString("OK")
})

log.Info("listening", "port", cfg.Port)
app.Listen(":" + cfg.Port)
```

## Log level from config

```go
cfg := config.Load()
log := logger.New(cfg.LogLevel, cfg.AppName)
// If LOG_LEVEL=debug, all debug messages are visible.
// If LOG_LEVEL=error, only error messages are printed.
```

## Architecture

Console output:

```
slog.Logger
  └─ zerologHandler (custom slog.Handler)
       └─ zerolog.Logger
            └─ ConsoleWriter (colorized pretty-print)
                 └─ os.Stdout
```

JSON output:

```
slog.Logger
  └─ zerologHandler (custom slog.Handler)
       └─ zerolog.Logger
            └─ os.Stdout (raw JSON)
```

A custom `slog.Handler` translates `slog.Record` into zerolog events, giving you:

- **Pretty console output** — colorized, human-readable output during development
- **JSON output** — structured lines for production log aggregation
- **Standard `slog` interface** — zero coupling to zerolog in your application code
- **`service` tag** — every log line is tagged with the service name for multi-service deployments

## Dependencies

- [`github.com/rs/zerolog`](https://github.com/rs/zerolog)
- [`github.com/gofiber/fiber/v2`](https://github.com/gofiber/fiber) (for the middleware)
