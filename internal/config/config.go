package config

import "github.com/sraj/everest/pkg/configx"

// Config holds all configuration for the application
type Config struct {
	AppName  string
	Port     string
	LogLevel string

	CORSOrigins string

	DatabaseURL string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	GRPCPort       string

	ZitadelIssuer   string
	ZitadelClientID string
}

func Load() *Config {
	l := configx.New(configx.WithDotEnv())

	return &Config{
		AppName:        l.String("APP_NAME", "everest"),
		Port:           l.String("PORT", "8080"),
		LogLevel:       l.String("LOG_LEVEL", "info"),
		CORSOrigins:    l.String("CORS_ORIGINS", "*"),
		DatabaseURL:    l.String("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/everest?sslmode=disable"),
		MinIOEndpoint:  l.String("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: l.String("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: l.String("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    l.String("MINIO_BUCKET", "documents"),
		MinIOUseSSL:    l.Bool("MINIO_USE_SSL", false),
		GRPCPort:       l.String("GRPC_PORT", ""),
		ZitadelIssuer:  l.String("ZITADEL_ISSUER", ""),
		ZitadelClientID: l.String("ZITADEL_CLIENT_ID", ""),
	}
}
