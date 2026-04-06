package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DB          DBConfig
	Redis       RedisConfig
	Storage     StorageConfig
	Concurrency int
	MaxRetries  int
	FFmpeg      FFmpegConfig
	TempDir     string
}

type DBConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func (c DBConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

type RedisConfig struct {
	Host string
	Port string
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type StorageConfig struct {
	Endpoint         string
	PublicEndpoint   string
	AccessKey        string
	SecretKey        string
	UsePathStyle     bool
	BucketRaw        string
	BucketProcessed  string
	BucketThumbnails string
}

type FFmpegConfig struct {
	CRF    string
	Preset string
}

func Load() (*Config, error) {
	concurrency, _ := strconv.Atoi(getEnv("WORKER_CONCURRENCY", "2"))
	maxRetries, _ := strconv.Atoi(getEnv("WORKER_MAX_RETRIES", "3"))
	storageEndpoint := firstEnv("OBJECT_STORAGE_ENDPOINT", "R2_ENDPOINT")
	storagePublicEndpoint := firstEnv("OBJECT_STORAGE_PUBLIC_ENDPOINT")
	if storagePublicEndpoint == "" {
		storagePublicEndpoint = storageEndpoint
	}

	return &Config{
		DB: DBConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "gotube"),
			Password: getEnv("POSTGRES_PASSWORD", "gotube_secret"),
			DBName:   getEnv("POSTGRES_DB", "gotube_lite"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		Storage: StorageConfig{
			Endpoint:         storageEndpoint,
			PublicEndpoint:   storagePublicEndpoint,
			AccessKey:        firstEnv("OBJECT_STORAGE_ACCESS_KEY", "R2_ACCESS_KEY_ID"),
			SecretKey:        firstEnv("OBJECT_STORAGE_SECRET_KEY", "R2_SECRET_ACCESS_KEY"),
			UsePathStyle:     getEnvBool("OBJECT_STORAGE_USE_PATH_STYLE", shouldUsePathStyle(storageEndpoint, storagePublicEndpoint)),
			BucketRaw:        firstEnvWithDefault("gotube-raw-videos-dev", "OBJECT_STORAGE_BUCKET_RAW", "R2_BUCKET_RAW"),
			BucketProcessed:  firstEnvWithDefault("gotube-processed-videos-dev", "OBJECT_STORAGE_BUCKET_PROCESSED", "R2_BUCKET_PROCESSED"),
			BucketThumbnails: firstEnvWithDefault("gotube-thumbnails-dev", "OBJECT_STORAGE_BUCKET_THUMBNAILS", "R2_BUCKET_THUMBNAILS"),
		},
		Concurrency: concurrency,
		MaxRetries:  maxRetries,
		FFmpeg: FFmpegConfig{
			CRF:    getEnv("FFMPEG_CRF", "23"),
			Preset: getEnv("FFMPEG_PRESET", "medium"),
		},
		TempDir: getEnv("TEMP_DIR", "/tmp/gotube-worker"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func firstEnvWithDefault(fallback string, keys ...string) string {
	if v := firstEnv(keys...); v != "" {
		return v
	}
	return fallback
}

func shouldUsePathStyle(endpoints ...string) bool {
	for _, endpoint := range endpoints {
		endpoint = strings.ToLower(endpoint)
		if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") || strings.Contains(endpoint, "minio:") {
			return true
		}
	}
	return false
}
