package repository

import "context"

// ContentRepository defines the interface for document content storage (MinIO)
type ContentRepository interface {
	Save(ctx context.Context, key string, content []byte, contentType string) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
