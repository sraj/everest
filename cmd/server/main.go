package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sraj/everest/internal/config"
	handlerhttp "github.com/sraj/everest/internal/handler/http"
	handlergrpc "github.com/sraj/everest/internal/handler/grpc"
	"github.com/sraj/everest/internal/infrastructure/minio"
	"github.com/sraj/everest/internal/infrastructure/postgres"
	"github.com/sraj/everest/internal/infrastructure/zitadel"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/internal/store"
	"github.com/sraj/everest/internal/version"
	"github.com/sraj/everest/pkg/dbx"
	"github.com/sraj/everest/pkg/logger"
	"github.com/sraj/everest/pkg/server"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		os.Exit(1)
	}
	log := logger.New(cfg.LogLevel, cfg.AppName)

	log.Info("Starting server",
		"version", version.Version,
		"commit", version.Commit,
		"build_date", version.BuildDate,
	)

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

	contentStore, err := minio.NewContentStore(minio.Config{
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

	minioClient, err := minio.NewClient(minio.Config{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		log.Error("Failed to create MinIO client for health checks", "error", err)
		os.Exit(1)
	}

	docStore := postgres.NewDocumentStore(db)
	st := store.New(docStore, contentStore, db.Close, db.Ping)

	thumbnailSvc := service.NewThumbnailService(service.DefaultThumbnailConfig(), log)
	defer thumbnailSvc.Close()
	docService := service.NewDocumentService(st, thumbnailSvc, log)

	var authMiddleware fiber.Handler
	if cfg.ZitadelClientID != "" && cfg.ZitadelClientID != "<created-in-zitadel-console>" {
		verifier, err := zitadel.NewVerifier(cfg.ZitadelIssuer, log)
		if err != nil {
			log.Error("Zitadel verifier initialization failed, continuing without auth", "error", err)
		} else {
			authMiddleware = verifier.Middleware()
			log.Info("Zitadel authentication enabled", "issuer", cfg.ZitadelIssuer)
		}
	} else {
		log.Info("Zitadel authentication disabled (no ZITADEL_CLIENT_ID configured)")
	}

	httpHandler := handlerhttp.New(docService, log, authMiddleware)
	httpHandler.AddHealthCheck("database", func(ctx context.Context) error {
		return db.Ping(ctx)
	})
	httpHandler.AddHealthCheck("storage", func(ctx context.Context) error {
		_, err := minioClient.BucketExists(ctx, cfg.MinIOBucket)
		return err
	})

	grpcHandler := handlergrpc.New(docService, log)

	httpServer := server.NewHTTP(server.HTTPConfig{
		AppName:        cfg.AppName,
		Port:           cfg.Port,
		CORSOrigins:    cfg.CORSOrigins,
		RateLimitMax:   100,
		RequestTimeout: 30 * time.Second,
		ErrorHandler:   handlerhttp.ErrorHandler(log),
	}, httpHandler, log)

	grpcServer := server.NewGRPC(server.GRPCConfig{
		Port: cfg.GRPCPort,
	}, grpcHandler)

	if err := server.Run(log, httpServer, grpcServer); err != nil {
		log.Error("Server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
