package config

import "testing"

func TestLoadPrefersObjectStorageVars(t *testing.T) {
	t.Setenv("OBJECT_STORAGE_ENDPOINT", "http://minio:9000")
	t.Setenv("OBJECT_STORAGE_PUBLIC_ENDPOINT", "http://localhost:9000")
	t.Setenv("OBJECT_STORAGE_ACCESS_KEY", "minio-access")
	t.Setenv("OBJECT_STORAGE_SECRET_KEY", "minio-secret")
	t.Setenv("OBJECT_STORAGE_USE_PATH_STYLE", "true")
	t.Setenv("OBJECT_STORAGE_BUCKET_RAW", "raw-bucket")
	t.Setenv("OBJECT_STORAGE_BUCKET_PROCESSED", "processed-bucket")
	t.Setenv("OBJECT_STORAGE_BUCKET_THUMBNAILS", "thumb-bucket")
	t.Setenv("R2_ENDPOINT", "https://legacy.example.com")
	t.Setenv("R2_ACCESS_KEY_ID", "legacy-access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "legacy-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Storage.Endpoint != "http://minio:9000" {
		t.Fatalf("unexpected endpoint: got %q", cfg.Storage.Endpoint)
	}
	if cfg.Storage.PublicEndpoint != "http://localhost:9000" {
		t.Fatalf("unexpected public endpoint: got %q", cfg.Storage.PublicEndpoint)
	}
	if cfg.Storage.AccessKey != "minio-access" || cfg.Storage.SecretKey != "minio-secret" {
		t.Fatalf("unexpected credentials: got %q / %q", cfg.Storage.AccessKey, cfg.Storage.SecretKey)
	}
	if !cfg.Storage.UsePathStyle {
		t.Fatalf("expected path style to be enabled")
	}
	if cfg.Storage.BucketRaw != "raw-bucket" || cfg.Storage.BucketProcessed != "processed-bucket" || cfg.Storage.BucketThumbnails != "thumb-bucket" {
		t.Fatalf("unexpected bucket config: %+v", cfg.Storage)
	}
}

func TestLoadFallsBackToLegacyR2Vars(t *testing.T) {
	t.Setenv("OBJECT_STORAGE_ENDPOINT", "")
	t.Setenv("OBJECT_STORAGE_PUBLIC_ENDPOINT", "")
	t.Setenv("OBJECT_STORAGE_ACCESS_KEY", "")
	t.Setenv("OBJECT_STORAGE_SECRET_KEY", "")
	t.Setenv("OBJECT_STORAGE_USE_PATH_STYLE", "")
	t.Setenv("OBJECT_STORAGE_BUCKET_RAW", "")
	t.Setenv("OBJECT_STORAGE_BUCKET_PROCESSED", "")
	t.Setenv("OBJECT_STORAGE_BUCKET_THUMBNAILS", "")
	t.Setenv("R2_ENDPOINT", "https://demo-account.r2.cloudflarestorage.com")
	t.Setenv("R2_ACCESS_KEY_ID", "legacy-access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "legacy-secret")
	t.Setenv("R2_BUCKET_RAW", "legacy-raw")
	t.Setenv("R2_BUCKET_PROCESSED", "legacy-processed")
	t.Setenv("R2_BUCKET_THUMBNAILS", "legacy-thumbs")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Storage.Endpoint != "https://demo-account.r2.cloudflarestorage.com" {
		t.Fatalf("unexpected endpoint: got %q", cfg.Storage.Endpoint)
	}
	if cfg.Storage.PublicEndpoint != "https://demo-account.r2.cloudflarestorage.com" {
		t.Fatalf("unexpected public endpoint: got %q", cfg.Storage.PublicEndpoint)
	}
	if cfg.Storage.AccessKey != "legacy-access" || cfg.Storage.SecretKey != "legacy-secret" {
		t.Fatalf("unexpected credentials: got %q / %q", cfg.Storage.AccessKey, cfg.Storage.SecretKey)
	}
	if cfg.Storage.UsePathStyle {
		t.Fatalf("expected path style to stay disabled for R2 fallback")
	}
	if cfg.Storage.BucketRaw != "legacy-raw" || cfg.Storage.BucketProcessed != "legacy-processed" || cfg.Storage.BucketThumbnails != "legacy-thumbs" {
		t.Fatalf("unexpected bucket config: %+v", cfg.Storage)
	}
}

func TestLoadParsesRefreshIdleTTL(t *testing.T) {
	t.Setenv("JWT_REFRESH_IDLE_TTL", "45m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JWT.RefreshIdleTTL.String() != "45m0s" {
		t.Fatalf("unexpected refresh idle ttl: got %s", cfg.JWT.RefreshIdleTTL)
	}
}
