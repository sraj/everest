package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sraj/everest/internal/handler"
	"github.com/sraj/everest/internal/infrastructure/minio"
	"github.com/sraj/everest/internal/infrastructure/postgres"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/pkg/config"
	"github.com/sraj/everest/pkg/dbx"
	"github.com/sraj/everest/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.AppName)

	log.Info("Starting server...")

	// Connect to PostgreSQL
	db, err := dbx.New(dbx.DBConfig{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to PostgreSQL")

	// Connect to MinIO
	contentRepo, err := minio.NewContentRepository(minio.Config{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Error("Failed to connect to MinIO", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to MinIO")

	// Create repositories
	docRepo := postgres.NewDocumentRepository(db)

	// Create thumbnail service
	thumbnailSvc := service.NewThumbnailService(service.DefaultThumbnailConfig(), log)

	// Create services
	docService := service.NewDocumentService(docRepo, contentRepo, thumbnailSvc, log)

	// Create handler
	h := handler.New(docService, log)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: handler.ErrorHandler(log),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH, OPTIONS",
	}))
	app.Use(logger.Middleware(log))

	// Register routes
	h.RegisterRoutes(app)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("Server started", "port", cfg.Port)

	<-ctx.Done()
	log.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Error("Error during shutdown", "error", err)
	}

	log.Info("Server stopped")
}
