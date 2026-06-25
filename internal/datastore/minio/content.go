package minio

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sraj/everest/internal/store"
)

// Config holds MinIO configuration.
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type contentStore struct {
	client *minio.Client
	bucket string
}

// NewContentStore creates a new MinIO-backed ContentStore.
func NewContentStore(cfg Config) (store.ContentStore, error) {
	return NewContentStoreFromClient(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.UseSSL, cfg.Bucket)
}

// NewContentStoreFromClient creates a ContentStore from explicit client parameters.
func NewContentStoreFromClient(endpoint, accessKey, secretKey string, useSSL bool, bucket string) (store.ContentStore, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &contentStore{client: client, bucket: bucket}, nil
}

// NewClient creates a raw MinIO client for health checks etc.
func NewClient(cfg Config) (*minio.Client, error) {
	return minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
}

func (r *contentStore) Save(ctx context.Context, key string, content []byte, contentType string) error {
	reader := bytes.NewReader(content)
	_, err := r.client.PutObject(ctx, r.bucket, key, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *contentStore) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (r *contentStore) Delete(ctx context.Context, key string) error {
	return r.client.RemoveObject(ctx, r.bucket, key, minio.RemoveObjectOptions{})
}
