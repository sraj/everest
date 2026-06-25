package config

import (
	"fmt"
	"strings"

	"github.com/sraj/everest/pkg/configx"
)

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

	ZitadelIssuer        string
	ZitadelClientID      string
	ZitadelRedirectURI   string
	ZitadelPostLogoutURI string
	ZitadelSessionSecret string
}

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// Validate checks all config values and returns an error if any are invalid.
func (c *Config) Validate() error {
	var errs []string

	if c.AppName == "" {
		errs = append(errs, "APP_NAME is required")
	}
	if c.LogLevel != "" && !validLogLevels[strings.ToLower(c.LogLevel)] {
		errs = append(errs, "LOG_LEVEL must be one of: debug, info, warn, error")
	}
	if c.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if c.MinIOEndpoint == "" {
		errs = append(errs, "MINIO_ENDPOINT is required")
	}
	if c.MinIOAccessKey == "" {
		errs = append(errs, "MINIO_ACCESS_KEY is required")
	}
	if c.MinIOSecretKey == "" {
		errs = append(errs, "MINIO_SECRET_KEY is required")
	}
	if c.MinIOBucket == "" {
		errs = append(errs, "MINIO_BUCKET is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func Load() (*Config, error) {
	l := configx.New(configx.WithDotEnv())

	cfg := &Config{
		AppName:        l.String("APP_NAME", "everest"),
		Port:           l.String("PORT", "8080"),
		LogLevel:       l.String("LOG_LEVEL", "info"),
		CORSOrigins:    l.String("CORS_ORIGINS", "http://localhost:5173"),
		DatabaseURL:    l.String("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/everest?sslmode=disable"),
		MinIOEndpoint:  l.String("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: l.String("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: l.String("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    l.String("MINIO_BUCKET", "documents"),
		MinIOUseSSL:    l.Bool("MINIO_USE_SSL", false),
		GRPCPort:       l.String("GRPC_PORT", ""),
		ZitadelIssuer:        l.String("ZITADEL_ISSUER", ""),
		ZitadelClientID:      l.String("ZITADEL_CLIENT_ID", ""),
		ZitadelRedirectURI:   l.String("ZITADEL_REDIRECT_URI", "http://localhost:8080/auth/callback"),
		ZitadelPostLogoutURI: l.String("ZITADEL_POST_LOGOUT_URI", "http://localhost:5173"),
		ZitadelSessionSecret: l.String("ZITADEL_SESSION_SECRET", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
