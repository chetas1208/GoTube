package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type VideoStatus string

const (
	VideoStatusUploaded   VideoStatus = "uploaded"
	VideoStatusQueued     VideoStatus = "queued"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusReady      VideoStatus = "ready"
	VideoStatusFailed     VideoStatus = "failed"
)

type VideoVisibility string

const (
	VisibilityPublic   VideoVisibility = "public"
	VisibilityUnlisted VideoVisibility = "unlisted"
	VisibilityPrivate  VideoVisibility = "private"
)

type Video struct {
	ID                 uuid.UUID       `json:"id"`
	UserID             uuid.UUID       `json:"user_id"`
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	Status             VideoStatus     `json:"status"`
	Visibility         VideoVisibility `json:"visibility"`
	DurationSeconds    *int            `json:"duration_seconds,omitempty"`
	ThumbnailObjectKey *string         `json:"thumbnail_object_key,omitempty"`
	SourceObjectKey    string          `json:"source_object_key"`
	ProcessedObjectKey *string         `json:"processed_object_key,omitempty"`
	ViewsCount         int             `json:"views_count"`
	LikesCount         int             `json:"likes_count"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	PublishedAt        *time.Time      `json:"published_at,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	Username           string          `json:"username,omitempty"`
}

type VideoTag struct {
	ID      uuid.UUID `json:"id"`
	VideoID uuid.UUID `json:"video_id"`
	Tag     string    `json:"tag"`
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type JobType string

const (
	JobTypeTranscode JobType = "transcode"
	JobTypeThumbnail JobType = "thumbnail"
)

type VideoProcessingJob struct {
	ID         uuid.UUID  `json:"id"`
	VideoID    uuid.UUID  `json:"video_id"`
	JobType    JobType    `json:"job_type"`
	Status     JobStatus  `json:"status"`
	Attempts   int        `json:"attempts"`
	LastError  *string    `json:"last_error,omitempty"`
	QueuedAt   time.Time  `json:"queued_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID  `json:"id"`
	VideoID   uuid.UUID  `json:"video_id"`
	UserID    uuid.UUID  `json:"user_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Username  string     `json:"username,omitempty"`
}

type Like struct {
	ID        uuid.UUID `json:"id"`
	VideoID   uuid.UUID `json:"video_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type VideoView struct {
	ID           uuid.UUID  `json:"id"`
	VideoID      uuid.UUID  `json:"video_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	SessionID    *string    `json:"session_id,omitempty"`
	ViewedAt     time.Time  `json:"viewed_at"`
	WatchSeconds *int       `json:"watch_seconds,omitempty"`
}

type RefreshToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt time.Time  `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type UploadStatus string

const (
	UploadStatusInitiated UploadStatus = "initiated"
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
	UploadStatusExpired   UploadStatus = "expired"
)

type UploadSession struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	VideoID      uuid.UUID    `json:"video_id"`
	ObjectKey    string       `json:"object_key"`
	UploadStatus UploadStatus `json:"upload_status"`
	ExpiresAt    time.Time    `json:"expires_at"`
	CreatedAt    time.Time    `json:"created_at"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}
