# GoTube Lite — Environment Variable Reference

## PostgreSQL

| Variable | Default | Description |
|----------|---------|-------------|
| DATABASE_URL | | Full Postgres connection string. When set, it overrides the POSTGRES_* variables below. |
| POSTGRES_HOST | localhost | Database host |
| POSTGRES_PORT | 5432 | Database port |
| POSTGRES_USER | gotube | Database user |
| POSTGRES_PASSWORD | gotube_secret | Database password |
| POSTGRES_DB | gotube_lite | Database name |

## Redis

| Variable | Default | Description |
|----------|---------|-------------|
| REDIS_HOST | localhost | Redis host |
| REDIS_PORT | 6379 | Redis port |

## Go API

| Variable | Default | Description |
|----------|---------|-------------|
| API_PORT | 8080 | API server port |
| API_ENV | development | Environment (development/production) |
| JWT_SECRET | (required) | Secret for signing JWTs. Use 64+ random chars |
| JWT_ACCESS_TTL | 15m | Access token lifetime |
| JWT_REFRESH_TTL | 168h | Refresh token lifetime (7 days) |
| CORS_ALLOWED_ORIGINS | http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000,http://127.0.0.1:3001 | Comma-separated CORS origins |

## Object Storage

| Variable | Default | Description |
|----------|---------|-------------|
| OBJECT_STORAGE_ENDPOINT | http://localhost:9000 | Internal S3-compatible endpoint used by the API and worker |
| OBJECT_STORAGE_PUBLIC_ENDPOINT | http://localhost:9000 | Browser-reachable endpoint used when generating presigned URLs |
| OBJECT_STORAGE_ACCESS_KEY | gotube_minio | Access key / MinIO root user |
| OBJECT_STORAGE_SECRET_KEY | gotube_minio_secret | Secret key / MinIO root password |
| OBJECT_STORAGE_USE_PATH_STYLE | true | Enables path-style URLs, required for local MinIO |
| OBJECT_STORAGE_BUCKET_RAW | gotube-raw-videos-dev | Bucket for raw uploads |
| OBJECT_STORAGE_BUCKET_PROCESSED | gotube-processed-videos-dev | Bucket for processed videos |
| OBJECT_STORAGE_BUCKET_THUMBNAILS | gotube-thumbnails-dev | Bucket for thumbnails |

Deprecated fallback variables:

| Variable | Description |
|----------|-------------|
| R2_ACCESS_KEY_ID | Legacy access key fallback |
| R2_SECRET_ACCESS_KEY | Legacy secret key fallback |
| R2_ENDPOINT | Legacy endpoint fallback |
| R2_BUCKET_RAW | Legacy raw bucket fallback |
| R2_BUCKET_PROCESSED | Legacy processed bucket fallback |
| R2_BUCKET_THUMBNAILS | Legacy thumbnail bucket fallback |

## Worker

| Variable | Default | Description |
|----------|---------|-------------|
| WORKER_CONCURRENCY | 2 | Number of concurrent processing goroutines |
| WORKER_MAX_RETRIES | 3 | Max retry attempts per job |
| FFMPEG_CRF | 23 | FFmpeg CRF quality (18-28 range, lower = better) |
| FFMPEG_PRESET | medium | FFmpeg preset (ultrafast to veryslow) |
| TEMP_DIR | /tmp/gotube-worker | Temporary file storage for processing |

## Upload Limits

| Variable | Default | Description |
|----------|---------|-------------|
| MAX_UPLOAD_SIZE_MB | 500 | Maximum upload file size in MB |
| ALLOWED_VIDEO_TYPES | video/mp4,video/quicktime,video/webm | Comma-separated MIME types |

## Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| RATE_LIMIT_AUTH_RPM | 20 | Auth endpoint requests per minute per IP |
| RATE_LIMIT_UPLOAD_RPM | 10 | Upload endpoint requests per minute per IP |

## Next.js Frontend

| Variable | Default | Description |
|----------|---------|-------------|
| NEXT_PUBLIC_API_URL | http://localhost:8080/api/v1 | Backend API URL (client-side) |
| NEXT_PUBLIC_APP_URL | http://localhost:3001 | Frontend app URL |
