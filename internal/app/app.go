package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/config"
	handlerhttp "github.com/sraj/everest/internal/handler/http"
	handlergrpc "github.com/sraj/everest/internal/handler/grpc"
	"github.com/sraj/everest/internal/auth"
	"github.com/sraj/everest/internal/datastore/minio"
	"github.com/sraj/everest/internal/datastore/postgres"
	"github.com/sraj/everest/internal/service"
	"github.com/sraj/everest/internal/jobs"
	"github.com/sraj/everest/internal/store"
	"github.com/sraj/everest/internal/version"
	"github.com/sraj/everest/pkg/dbx"
	"github.com/sraj/everest/pkg/logger"
	"github.com/sraj/everest/pkg/server"
)

// Server is the application lifecycle interface.
type Server interface {
	Run() error
	Close()
}

// App holds all bootstrapped dependencies and manages the server lifecycle.
type App struct {
	cfg *config.Config
	log *slog.Logger
	db  *dbx.DB
}

// New boots config, logger, and database. Returns an unconfigured App ready for Run().
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		return nil, err
	}
	log := logger.New(cfg.LogLevel, cfg.AppName)
	log.Info("starting server",
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
		return nil, fmt.Errorf("database: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}
	log.Info("connected to PostgreSQL")

	return &App{cfg: cfg, log: log, db: db}, nil
}

// Run wires all services and starts the HTTP + gRPC servers. Blocks until shutdown.
func (a *App) Run() error {
	defer a.db.Close()

	contentStore, err := minio.NewContentStore(minio.Config{
		Endpoint:  a.cfg.MinIOEndpoint,
		AccessKey: a.cfg.MinIOAccessKey,
		SecretKey: a.cfg.MinIOSecretKey,
		Bucket:    a.cfg.MinIOBucket,
		UseSSL:    a.cfg.MinIOUseSSL,
	})
	if err != nil {
		return fmt.Errorf("minio: %w", err)
	}
	a.log.Info("connected to MinIO")

	minioClient, err := minio.NewClient(minio.Config{
		Endpoint:  a.cfg.MinIOEndpoint,
		AccessKey: a.cfg.MinIOAccessKey,
		SecretKey: a.cfg.MinIOSecretKey,
		Bucket:    a.cfg.MinIOBucket,
		UseSSL:    a.cfg.MinIOUseSSL,
	})
	if err != nil {
		return fmt.Errorf("minio client: %w", err)
	}

	docStore := postgres.NewDocumentStore(a.db)
	profileStore := postgres.NewUserProfileStore(a.db)
	st := store.New(docStore, contentStore, profileStore, a.db.Close, a.db.Ping)

	thumbnailSvc := service.NewThumbnailService(service.DefaultThumbnailConfig(), a.log)
	defer thumbnailSvc.Close()

	tagger := service.NewTagger(service.DefaultTaggerConfig(), st, a.log)
	defer tagger.Close()

	pool := jobs.New(jobs.DefaultConfig(a.log))
	defer pool.Shutdown(30 * time.Second)

	docService := service.NewDocumentService(st, thumbnailSvc, tagger, pool, a.log)

	var authMiddleware fiber.Handler
	var bffHandler *auth.BFFHandler
	if a.cfg.ZitadelClientID != "" && a.cfg.ZitadelClientID != "<created-in-zitadel-console>" {
		bffHandler, err = auth.NewBFFHandler(auth.BFFConfig{
			Issuer:        a.cfg.ZitadelIssuer,
			ClientID:      a.cfg.ZitadelClientID,
			RedirectURI:   a.cfg.ZitadelRedirectURI,
			PostLogoutURI: a.cfg.ZitadelPostLogoutURI,
			SessionSecret: a.cfg.ZitadelSessionSecret,
			Log:           a.log,
		})
		if err != nil {
			a.log.Error("BFF auth initialization failed, continuing without auth", "error", err)
		} else {
			authMiddleware = bffHandler.BFFMiddleware()
			a.log.Info("BFF authentication enabled", "issuer", a.cfg.ZitadelIssuer)
		}
	} else {
		a.log.Info("Zitadel authentication disabled (no ZITADEL_CLIENT_ID configured)")
	}

	httpHandler := handlerhttp.New(docService, a.log, authMiddleware)
	if bffHandler != nil {
		httpHandler.SetBFFHandler(bffHandler)
	}
	httpHandler.AddHealthCheck("database", func(ctx context.Context) error {
		return a.db.Ping(ctx)
	})
	httpHandler.AddHealthCheck("storage", func(ctx context.Context) error {
		_, err := minioClient.BucketExists(ctx, a.cfg.MinIOBucket)
		return err
	})

	grpcHandler := handlergrpc.New(docService, a.log)

	httpSrv := server.NewHTTP(server.HTTPConfig{
		AppName:        a.cfg.AppName,
		Port:           a.cfg.Port,
		CORSOrigins:    a.cfg.CORSOrigins,
		RateLimitMax:   100,
		RequestTimeout: 30 * time.Second,
		ErrorHandler:   handlerhttp.ErrorHandler(a.log),
	}, httpHandler, a.log)

	grpcSrv := server.NewGRPC(server.GRPCConfig{
		Port: a.cfg.GRPCPort,
	}, grpcHandler)

	return server.Run(a.log, httpSrv, grpcSrv)
}

// Close shuts down database connections.
func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
