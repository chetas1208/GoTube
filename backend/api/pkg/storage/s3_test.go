package storage

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestPresignedURLsUsePublicEndpointAndPathStyle(t *testing.T) {
	store, err := NewS3Storage(Config{
		Endpoint:       "http://minio:9000",
		PublicEndpoint: "http://localhost:9000",
		AccessKey:      "minio-access",
		SecretKey:      "minio-secret",
		UsePathStyle:   true,
	})
	if err != nil {
		t.Fatalf("NewS3Storage() error = %v", err)
	}

	putURL, err := store.GeneratePresignedPutURL(context.Background(), "videos", "raw/demo/video.mp4", "video/mp4", time.Minute)
	if err != nil {
		t.Fatalf("GeneratePresignedPutURL() error = %v", err)
	}

	getURL, err := store.GeneratePresignedGetURL(context.Background(), "videos", "thumbs/demo.jpg", time.Minute)
	if err != nil {
		t.Fatalf("GeneratePresignedGetURL() error = %v", err)
	}

	assertURLParts(t, putURL, "localhost:9000", "/videos/raw/demo/video.mp4")
	assertURLParts(t, getURL, "localhost:9000", "/videos/thumbs/demo.jpg")
}

func assertURLParts(t *testing.T, rawURL, expectedHost, expectedPath string) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	if parsed.Host != expectedHost {
		t.Fatalf("unexpected host: got %q want %q", parsed.Host, expectedHost)
	}
	if parsed.Path != expectedPath {
		t.Fatalf("unexpected path: got %q want %q", parsed.Path, expectedPath)
	}
	if parsed.Query().Get("X-Amz-Signature") == "" {
		t.Fatalf("expected presigned URL query params, got %q", rawURL)
	}
}
