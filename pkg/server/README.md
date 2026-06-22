# server

`server` provides a reusable HTTP + gRPC server lifecycle with graceful shutdown for Go applications.

## Quick Start

```go
package main

import (
	"log/slog"
	"os"

	"your.mod/pkg/server"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	httpCfg := server.HTTPConfig{
		AppName:     "myapp",
		Port:        "8080",
		CORSOrigins: "*",
	}
	httpSrv := server.NewHTTP(httpCfg, myRoutes{}, log)

	if err := server.Run(log, httpSrv); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
```

## HTTP Server

```go
// RouteRegistrar is the interface your handlers must implement.
type myRoutes struct{}

func (myRoutes) RegisterRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
```

## gRPC Server

```go
type myGRPC struct{}

func (myGRPC) RegisterServices(s *grpc.Server) {
	// pb.RegisterMyServiceServer(s, &myImpl{})
}

grpcSrv := server.NewGRPC(server.GRPCConfig{Port: "9090"}, myGRPC{})
server.Run(log, httpSrv, grpcSrv)
```

## Optional Servers

Any server with an empty port is automatically disabled:

| PORT | GRPC_PORT | Result |
|---|---|---|
| `"8080"` | `""` | HTTP only |
| `""` | `"9090"` | gRPC only |
| `"8080"` | `"9090"` | Both |
| `""` | `""` | Nothing |

Disabled servers log a message and are skipped silently.

## Graceful Shutdown

`Run()` handles SIGINT/SIGTERM and shuts down all servers in reverse order with a 10-second timeout.

## API

```go
// Server is the interface for any transport server.
type Server interface {
    Name() string
    Start() error
    Shutdown(ctx context.Context) error
}

// Run starts servers and blocks until shutdown signal or fatal error.
func Run(log *slog.Logger, servers ...Server) error
```
