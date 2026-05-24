# logger

**logger** provides a structured logger backed by [`zerolog`](https://github.com/rs/zerolog) exposed through the standard library's [`log/slog`](https://pkg.go.dev/log/slog) API, with a pretty-printed console output and a Fiber HTTP middleware.

## Quick start

```go
import "github.com/sraj/everest/pkg/logger"

log := logger.New("debug", "everest")
log.Info("server starting", "port", 8080)
```

Output:

```
2026-05-23T20:00:00Z INF server starting port=8080 service=everest
```

## Log levels

```go
log := logger.New("debug", "everest")

log.Debug("connecting to database", "host", "localhost", "port", 5432)
log.Info("server started", "port", 8080)
log.Warn("slow query detected", "duration", 250*time.Millisecond)
log.Error("failed to connect", "error", err)
```

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

## Contextual logging with `With`

```go
// Create a child logger with request-scoped attributes.
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

## Fiber middleware

```go
import "github.com/sraj/everest/pkg/logger"

app := fiber.New()
log := logger.New("debug", "everest")
app.Use(logger.Middleware(log))
```

Each request is logged with:

```
method=GET path=/api/v1/documents status=200 ip=127.0.0.1 request_id=abc latency=12ms
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
import "github.com/sraj/everest/pkg/config"
import "github.com/sraj/everest/pkg/logger"

cfg := config.Load()
log := logger.New(cfg.LogLevel, cfg.AppName)
// If LOG_LEVEL=debug, all debug messages are visible.
// If LOG_LEVEL=error, only error messages are printed.
```

## Disable colors

The zerolog console writer can be customized for production use:

```go
output := zerolog.ConsoleWriter{
    Out:        os.Stdout,
    TimeFormat: time.RFC3339,
    NoColor:    true,
}

// To use a custom writer, modify logger.New or build the logger manually.
```

## Architecture

```
slog.Logger
  └─ zerologHandler (custom slog.Handler)
       └─ zerolog.Logger
            └─ ConsoleWriter (pretty-printed)
                 └─ os.Stdout
```

A custom `slog.Handler` translates `slog.Record` into zerolog events, giving you:

- **zerolog's pretty console writer** — colorized, human-readable output during development
- **Standard `slog` interface** — zero coupling to zerolog in your application code
- **`service` tag** — every log line is tagged with the service name for multi-service deployments

## Dependencies

- [`github.com/rs/zerolog`](https://github.com/rs/zerolog)
- [`github.com/gofiber/fiber/v2`](https://github.com/gofiber/fiber) (for the middleware)
