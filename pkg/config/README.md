# config

**config** loads the application configuration from environment variables with `.env` file support via [`godotenv`](https://github.com/joho/godotenv).

## Quick start

```go
import "github.com/sraj/everest/pkg/config"

cfg := config.Load()

fmt.Println("App:", cfg.AppName)       // "everest"
fmt.Println("Port:", cfg.Port)         // "8080"
fmt.Println("Log level:", cfg.LogLevel) // "info"
```

## Initializing the app with config

```go
cfg := config.Load()

app := fiber.New(fiber.Config{
    AppName: cfg.AppName,
})

log := logger.New(cfg.LogLevel, cfg.AppName)
```

## Database connection

```go
cfg := config.Load()

db, err := dbx.New(dbx.DBConfig{
    DSN: cfg.DatabaseURL,
})
```

## MinIO client

```go
cfg := config.Load()

minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
    Secure: cfg.MinIOUseSSL,
})
```

## CORS middleware

```go
cfg := config.Load()

app.Use(cors.New(cors.Config{
    AllowOrigins: cfg.CORSOrigins,
}))
```

## Using the full config

```go
cfg := config.Load()

app := fiber.New(fiber.Config{AppName: cfg.AppName})

app.Use(logger.Middleware(logger.New(cfg.LogLevel, cfg.AppName)))
app.Use(cors.New(cors.Config{AllowOrigins: cfg.CORSOrigins}))

db, _ := dbx.New(dbx.DBConfig{DSN: cfg.DatabaseURL})

mc, _ := minio.New(cfg.MinIOEndpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
    Secure: cfg.MinIOUseSSL,
})
```

## Env file

`Load()` calls `godotenv.Load()` automatically, so a `.env` file in the working directory is picked up without extra setup:

```env
APP_NAME=everest
PORT=8080
LOG_LEVEL=debug
CORS_ORIGINS=http://localhost:5173
DATABASE_URL=postgres://postgres:postgres@localhost:5432/everest?sslmode=disable
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=documents
MINIO_USE_SSL=false
```

## Override with environment variables

Environment variables always take precedence over `.env` values:

```bash
export PORT=3000
export LOG_LEVEL=debug
go run .
```

This starts the app on port 3000 with debug logging, while the rest of the config comes from `.env`.

## Multiple env files

```go
// Load a specific env file before config.Load
_ = godotenv.Load("config/production.env")
cfg := config.Load()
```

Or load several with cascading overrides:

```go
_ = godotenv.Load(".env.defaults", ".env.overrides")
cfg := config.Load()
```

## Custom defaults

For tests, you can create configs with non-standard defaults:

```go
cfg := &config.Config{
    AppName:       "everest-test",
    Port:          "0", // random port
    LogLevel:      "debug",
    DatabaseURL:   "postgres://test:test@localhost:5432/everest_test?sslmode=disable",
    MinIOEndpoint: "localhost:9000",
    MinIOBucket:   "test-bucket",
}
```

## Validation

The current `Config` struct does no validation at load time — you can add it at the call site:

```go
cfg := config.Load()

if cfg.DatabaseURL == "" {
    log.Fatal("DATABASE_URL is required")
}

port, err := strconv.Atoi(cfg.Port)
if err != nil {
    log.Fatalf("invalid PORT: %v", err)
}

if cfg.MinIOEndpoint == "" {
    log.Fatal("MINIO_ENDPOINT is required")
}
```

## Dependencies

- [`github.com/joho/godotenv`](https://github.com/joho/godotenv)
