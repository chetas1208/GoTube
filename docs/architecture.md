# GoTube Lite — Architecture Document

## System Overview

GoTube Lite is a monorepo containing three main services:

1. **Web** (Next.js) — SSR frontend with client interactivity
2. **API** (Go) — REST API for all business logic
3. **Worker** (Go) — Background video processor

## Data Flow

### Upload Sequence

```
User → Frontend (upload form)
  → POST /api/v1/videos/initiate-upload
    → API creates video record (status: uploaded)
    → API creates upload_session
    → API generates presigned PUT URL for R2
    → Returns { video_id, upload_url, object_key }

User → Frontend (XHR PUT to presigned URL)
  → File uploads directly to Cloudflare R2
  → Frontend tracks progress via XHR events

User → Frontend (POST /api/v1/videos/{id}/complete-upload)
  → API verifies object exists via HeadObject
  → API updates upload_session to completed
  → API updates video status to queued
  → API creates video_processing_job
  → API enqueues job to Redis Stream
```

### Processing Sequence

```
Worker → Redis XREADGROUP (blocking poll)
  → Receives job message { job_id, video_id }
  → Fetches job record from DB (idempotency check)
  → Updates job status to running
  → Updates video status to processing
  → Downloads raw video from R2 to temp directory
  → FFmpeg transcode: H.264/AAC, CRF 23, medium, faststart
  → FFmpeg thumbnail: frame at 25% duration, 640px wide
  → Uploads processed MP4 to R2 (processed bucket)
  → Uploads thumbnail to R2 (thumbnails bucket)
  → Extracts duration via ffprobe
  → Updates video record: status=ready, processed_object_key, thumbnail_object_key, duration
  → Updates job status to completed
  → Cleans up temp files
  → ACKs Redis message
```

### Playback Sequence

```
User → GET /watch/{id}
  → Frontend SSR/CSR fetches video metadata
  → Frontend fetches GET /api/v1/videos/{id}/playback
    → API generates presigned GET URL for processed MP4 in R2
    → Returns { playback_url, content_type }
  → HTML5 <video> element loads from signed R2 URL
  → R2 handles byte-range requests natively
  → Frontend fires POST /api/v1/videos/{id}/view
```

## Database Schema

9 tables: users, videos, video_tags, video_processing_jobs, comments, likes, video_views, refresh_tokens, upload_sessions.

See migration files in `backend/api/migrations/`.

## Video Status State Machine

```
uploaded → queued → processing → ready
                              ↘ failed
```

## Trending Algorithm

```sql
score = (views in past 7 days) + (likes_count * 0.5)
ORDER BY score DESC
```

Computed on read via indexed subquery. For scale, would move to periodic materialized view.

## Search Implementation

Uses PostgreSQL `pg_trgm` extension with GIN indexes on `videos.title` and `video_tags.tag`.
Supports partial matching via ILIKE with trigram acceleration.

## Auth Flow

- JWT access token (15min TTL) — sent via Authorization header
- Refresh token (7d TTL) — stored as SHA-256 hash in DB, sent via httpOnly cookie
- Token rotation on refresh (old revoked, new issued)

## FFmpeg Settings

- Codec: libx264 + aac
- Quality: CRF 23 (visually near-lossless, ~40-60% size reduction)
- Preset: medium (good speed/compression balance)
- Container: MP4 with movflags +faststart (progressive playback)
- Audio: 128kbps AAC
- Thumbnail: single JPEG frame, 640px wide, quality 3

## Future Evolution (Post-MVP)

- HLS packaging (multiple renditions)
- Chunked/resumable uploads (tus protocol)
- CDN integration
- Admin panel
- Advanced analytics
- Multiple video qualities
