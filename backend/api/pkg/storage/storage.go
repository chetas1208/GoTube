package storage

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)
	GeneratePresignedPutURL(ctx context.Context, bucket, key, contentType string, expiry time.Duration) (string, error)
	GeneratePresignedGetURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}

type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

type Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	UsePathStyle   bool
}
