package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sraj/everest/internal/domain/repository"
)

type contentRepository struct {
	client *minio.Client
	bucket string
}

// Config holds MinIO configuration
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewClient creates a raw MinIO client for health checks and direct access.
func NewClient(cfg Config) (*minio.Client, error) {
	return minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
}

// NewContentRepository creates a new MinIO content repository
func NewContentRepository(cfg Config) (repository.ContentRepository, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &contentRepository{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (r *contentRepository) Save(ctx context.Context, key string, content []byte, contentType string) error {
	reader := bytes.NewReader(content)
	_, err := r.client.PutObject(ctx, r.bucket, key, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *contentRepository) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

func (r *contentRepository) Delete(ctx context.Context, key string) error {
	return r.client.RemoveObject(ctx, r.bucket, key, minio.RemoveObjectOptions{})
}
