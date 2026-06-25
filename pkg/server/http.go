package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sraj/everest/pkg/logger"
)

// HTTPConfig contains HTTP server configuration.
type HTTPConfig struct {
	AppName        string
	Port           string
	CORSOrigins    string
	RateLimitMax   int           // max requests per minute, 0 = disabled
	RequestTimeout time.Duration // per-request context timeout, 0 = disabled
	ErrorHandler   fiber.ErrorHandler
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
		BodyLimit:    32 * 1024 * 1024, // 32MB
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-User-ID, X-User-Roles",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))
	app.Use(securityHeaders)
	app.Use(logger.Middleware(log))

	if cfg.RateLimitMax > 0 {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.RateLimitMax,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "rate limit exceeded, try again later",
				})
			},
		}))
	}

	if cfg.RequestTimeout > 0 {
		app.Use(func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(c.Context(), cfg.RequestTimeout)
			defer cancel()
			c.SetUserContext(ctx)
			return c.Next()
		})
	}

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

var securityHeaders = func(c *fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: blob:; connect-src 'self'")
	if c.Protocol() == "https" {
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	return c.Next()
}
