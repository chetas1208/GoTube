# GoTube Lite

A production-style mini YouTube clone with asynchronous video processing. Built with Go, Next.js, PostgreSQL, Redis, and S3-compatible object storage.

## Architecture

```
Browser → Next.js (SSR) → Go API → PostgreSQL + Redis + MinIO / S3-compatible storage
                                ↓
                          Go Worker (FFmpeg) → processed video + early thumbnail upload
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 14 (App Router) + TypeScript + Tailwind CSS |
| Backend API | Go + chi router |
| Worker | Go + FFmpeg |
| Database | PostgreSQL 16 |
| Queue/Cache | Redis 7 (Streams) |
| Object Storage | MinIO locally, any S3-compatible backend in deployment |
| Auth | JWT access + refresh token rotation |
| Observability | Prometheus + zerolog |
| Containerization | Docker Compose |

## Features

- User registration and authentication
- Video upload with presigned URLs (direct to object storage)
- Async video processing (H.264 transcode + early best-effort thumbnail generation)
- Progressive MP4 playback via signed URLs
- Video metadata CRUD
- Comments, likes, view tracking
- Search (trigram-based title + tag search)
- Trending (weighted recent views)
- Video status lifecycle (uploaded → queued → processing → ready/failed)
- Rate limiting, validation, CORS
- Prometheus metrics, health/readiness endpoints

## Quick Start

### Prerequisites

- Docker & Docker Compose
- `golang-migrate` CLI (for running migrations locally)

### 1. Setup Environment

```bash
cp .env.example .env
# Edit .env with a strong JWT_SECRET. Local MinIO defaults are already included.
```

### 2. Start Local Object Storage

`docker compose` now starts MinIO on `localhost:9000` and bootstraps these buckets automatically:
- `gotube-raw-videos-dev`
- `gotube-processed-videos-dev`
- `gotube-thumbnails-dev`

### 3. Start Services

```bash
make up
```

This starts: PostgreSQL, Redis, MinIO, bucket bootstrap, Go API, Go Worker, Next.js frontend, Prometheus.

### 4. Run Migrations

```bash
# With golang-migrate installed:
./infra/scripts/migrate.sh up

# Or manually:
migrate -path ./backend/api/migrations \
  -database "postgres://gotube:gotube_secret@localhost:5432/gotube_lite?sslmode=disable" \
  -verbose up
```

### 5. Seed Data (Optional)

```bash
./infra/scripts/seed.sh
```

### 6. Access

- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **MinIO API**: http://localhost:9000
- **MinIO Console**: http://localhost:9001
- **Prometheus**: http://localhost:9090

## Project Structure

```
gotube-lite/
├── frontend/              # Next.js frontend
├── backend/
│   ├── api/               # Go REST API
│   │   ├── cmd/api/       # Entry point
│   │   ├── internal/      # Business logic
│   │   └── migrations/    # SQL migrations
│   └── worker/            # Go background processor
│       ├── cmd/worker/    # Entry point
│       └── internal/      # Processing logic
├── infra/
│   ├── docker/            # Docker Compose
│   ├── prometheus/        # Prometheus config
│   └── scripts/           # Migration/seed scripts
└── docs/                  # Documentation
```

## API Endpoints

See [docs/api-spec.md](docs/api-spec.md) for the complete API specification.

## Development

```bash
# Run tests
make test

# Lint
make lint-api
make lint-web

# Build
make build-api
make build-worker
make build-web
```

## Video Processing Pipeline

1. User initiates upload → API creates video record + presigned PUT URL
2. Frontend uploads directly to object storage via a presigned PUT URL
3. API validates object, creates processing job, enqueues to Redis
4. Worker consumes job, downloads raw video from object storage
5. Worker extracts multiple thumbnail candidates from the source video, scores them, and uploads the best thumbnail immediately
6. FFmpeg transcodes to H.264 MP4 (CRF 23, medium preset, faststart)
7. Worker uploads processed MP4, then marks the video `ready`
8. Studio polls while uploads are in flight so the thumbnail appears before the video reaches `ready`
9. Watch page fetches a signed playback URL, then streams from object storage

## License

MIT
