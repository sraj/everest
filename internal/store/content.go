package store

import "context"

// ContentStore defines the interface for blob storage (MinIO).
type ContentStore interface {
	Save(ctx context.Context, key string, content []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
