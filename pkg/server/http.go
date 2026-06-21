package server

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sraj/everest/pkg/logger"
)

// HTTPConfig contains HTTP server configuration.
type HTTPConfig struct {
	AppName      string
	Port         string
	CORSOrigins  string
	ErrorHandler fiber.ErrorHandler
}

// RouteRegistrar registers application routes.
type RouteRegistrar interface {
	RegisterRoutes(app *fiber.App)
}

// HTTPServer wraps Fiber app setup and lifecycle.
type HTTPServer struct {
	app *fiber.App
	cfg HTTPConfig
}

// NewHTTP creates a new HTTP server. Returns nil if cfg.Port is empty.
func NewHTTP(cfg HTTPConfig, routes RouteRegistrar, log *slog.Logger) Server {
	if cfg.Port == "" {
		log.Info("http server disabled (port is empty)")
		return noop{}
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: cfg.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-User-ID, X-User-Roles",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))
	app.Use(logger.Middleware(log))

	routes.RegisterRoutes(app)

	return &HTTPServer{app: app, cfg: cfg}
}

// Start starts listening for incoming HTTP requests.
func (s *HTTPServer) Start() error {
	return s.app.Listen(":" + s.cfg.Port)
}

// Shutdown gracefully shuts down the server.
func (s *HTTPServer) Shutdown(_ context.Context) error {
	return s.app.Shutdown()
}

// Name returns the server name.
func (s *HTTPServer) Name() string {
	return "http"
}
