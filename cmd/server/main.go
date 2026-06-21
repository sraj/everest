package main

import (
	"context"
	"os"
	"time"

	"github.com/sraj/everest/internal/config"
	handlerhttp "github.com/sraj/everest/internal/handler/http"
	handlergrpc "github.com/sraj/everest/internal/handler/grpc"
	"github.com/sraj/everest/internal/infrastructure/minio"
	"github.com/sraj/everest/internal/infrastructure/postgres"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/pkg/dbx"
	"github.com/sraj/everest/pkg/logger"
	"github.com/sraj/everest/pkg/server"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel, cfg.AppName)

	log.Info("Starting server...")

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

	docRepo := postgres.NewDocumentRepository(db)
	thumbnailSvc := service.NewThumbnailService(service.DefaultThumbnailConfig(), log)
	docService := service.NewDocumentService(docRepo, contentRepo, thumbnailSvc, log)

	httpHandler := handlerhttp.New(docService, log)
	grpcHandler := handlergrpc.New(docService)

	httpServer := server.NewHTTP(server.HTTPConfig{
		AppName:      cfg.AppName,
		Port:         cfg.Port,
		CORSOrigins:  cfg.CORSOrigins,
		ErrorHandler: handlerhttp.ErrorHandler(log),
	}, httpHandler, log)

	grpcServer := server.NewGRPC(server.GRPCConfig{
		Port: cfg.GRPCPort,
	}, grpcHandler)

	if err := server.Run(log, httpServer, grpcServer); err != nil {
		log.Error("Server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
