package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	API       APIConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Storage   StorageConfig
	Upload    UploadConfig
	RateLimit RateLimitConfig
}

type APIConfig struct {
	Port               string
	Env                string
	CORSAllowedOrigins string
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

type JWTConfig struct {
	Secret         string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	RefreshIdleTTL time.Duration
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

type UploadConfig struct {
	MaxSizeMB         int64
	AllowedVideoTypes string
}

type RateLimitConfig struct {
	AuthRPM   int
	UploadRPM int
}

func Load() (*Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	refreshIdleTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_IDLE_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_IDLE_TTL: %w", err)
	}
	maxSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE_MB", "500"), 10, 64)
	authRPM, _ := strconv.Atoi(getEnv("RATE_LIMIT_AUTH_RPM", "20"))
	uploadRPM, _ := strconv.Atoi(getEnv("RATE_LIMIT_UPLOAD_RPM", "10"))
	storageEndpoint := firstEnv("OBJECT_STORAGE_ENDPOINT", "R2_ENDPOINT")
	storagePublicEndpoint := firstEnv("OBJECT_STORAGE_PUBLIC_ENDPOINT")
	if storagePublicEndpoint == "" {
		storagePublicEndpoint = storageEndpoint
	}

	return &Config{
		API: APIConfig{
			Port:               getEnv("API_PORT", "8080"),
			Env:                getEnv("API_ENV", "development"),
			CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
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
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", ""),
			AccessTTL:      accessTTL,
			RefreshTTL:     refreshTTL,
			RefreshIdleTTL: refreshIdleTTL,
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
		Upload: UploadConfig{
			MaxSizeMB:         maxSize,
			AllowedVideoTypes: getEnv("ALLOWED_VIDEO_TYPES", "video/mp4,video/quicktime,video/webm"),
		},
		RateLimit: RateLimitConfig{
			AuthRPM:   authRPM,
			UploadRPM: uploadRPM,
		},
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
