# GoTube Lite — API Specification

Base URL: `http://localhost:8080/api/v1`

---

## Auth

### POST /auth/register
Create a new user account.

**Auth**: None

**Request Body**:
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepass123"
}
```

**Validation**: username 3-50 chars, email valid, password 8-128 chars

**Response** `201`:
```json
{
  "access_token": "eyJhbG...",
  "user": { "id": "uuid", "username": "johndoe", "email": "john@example.com" }
}
```

**Set-Cookie**: `refresh_token=...; HttpOnly; Path=/api/v1/auth`

**Errors**: `400` validation, `409` user exists

---

### POST /auth/login

**Auth**: None

**Request Body**:
```json
{ "email": "john@example.com", "password": "securepass123" }
```

**Response** `200`: Same as register

**Errors**: `401` invalid credentials

---

### POST /auth/logout

**Auth**: Required

**Response** `200`:
```json
{ "message": "logged out" }
```

---

### GET /auth/me

**Auth**: Required

**Response** `200`:
```json
{ "id": "uuid", "username": "johndoe", "email": "john@example.com" }
```

---

### POST /auth/refresh

**Auth**: None (uses refresh_token cookie)

**Response** `200`: Same as login (new access + refresh tokens)

---

## Videos

### POST /videos/initiate-upload

**Auth**: Required | **Rate Limited**: 10/min

**Request Body**:
```json
{
  "title": "My Video",
  "description": "A great video",
  "tags": ["tutorial", "go"],
  "filename": "video.mp4",
  "content_type": "video/mp4",
  "file_size": 52428800
}
```

**Validation**: title 1-255 chars, description max 5000, tags max 20, file type/size limits

**Response** `201`:
```json
{
  "video_id": "uuid",
  "upload_url": "https://...(presigned PUT URL)...",
  "object_key": "raw/userId/videoId/original.mp4"
}
```

---

### POST /videos/{id}/complete-upload

**Auth**: Required (owner only)

**Response** `200`:
```json
{ "message": "upload completed, processing started" }
```

**Errors**: `403` not owner, `400` invalid state

---

### GET /videos

**Auth**: Optional

**Query**: `page`, `per_page`

**Response** `200`:
```json
{
  "videos": [{ ...video }],
  "total_count": 100,
  "page": 1,
  "per_page": 20
}
```

---

### GET /videos/{id}

**Auth**: Optional

**Response** `200`: Full video object with tags, username, user_has_liked

---

### PATCH /videos/{id}

**Auth**: Required (owner only)

**Request Body** (all optional):
```json
{
  "title": "Updated Title",
  "description": "New desc",
  "tags": ["updated"],
  "visibility": "unlisted"
}
```

---

### DELETE /videos/{id}

**Auth**: Required (owner only)

**Response** `200`: `{ "message": "video deleted" }`

---

### POST /videos/{id}/view

**Auth**: Optional

**Query**: `session_id` (optional)

**Response** `200`: `{ "message": "view recorded" }`

---

### POST /videos/{id}/like

**Auth**: Required

**Response** `200`:
```json
{ "liked": true }
```

Toggles like state. Returns current state.

---

### DELETE /videos/{id}/like

**Auth**: Required

**Response** `200`: `{ "message": "like removed" }`

---

### GET /videos/{id}/playback

**Auth**: Optional

**Response** `200`:
```json
{
  "playback_url": "https://...(signed R2 GET URL, 2h expiry)...",
  "content_type": "video/mp4"
}
```

**Errors**: `404` if video not ready

---

### GET /videos/my

**Auth**: Required

**Query**: `page`, `per_page`

**Response**: Same as GET /videos (but includes all statuses for the authenticated user)

---

## Comments

### GET /videos/{id}/comments

**Auth**: Optional

**Query**: `page`, `per_page`

**Response** `200`:
```json
{
  "comments": [{ "id": "...", "username": "...", "body": "...", ... }],
  "total_count": 5,
  "page": 1,
  "per_page": 20
}
```

---

### POST /videos/{id}/comments

**Auth**: Required

**Request Body**:
```json
{ "body": "Great video!", "parent_id": null }
```

**Validation**: body 1-2000 chars

**Response** `201`: Comment object

---

## Search

### GET /search

**Auth**: Optional

**Query**: `q` (required), `page`, `per_page`, `sort_by` (relevance|recent|views)

**Response**: Same format as video list

---

## Trending

### GET /trending

**Auth**: Optional

**Query**: `limit` (default 20, max 50)

**Response** `200`: Array of video objects

---

## Health

### GET /health
Returns `{ "status": "ok" }` — always 200

### GET /ready
Checks DB and Redis connectivity. Returns 200 or 503.

### GET /metrics
Prometheus metrics endpoint.
