package model

import (
	"time"

	"github.com/google/uuid"
)

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
	ID                uuid.UUID       `json:"id"`
	UserID            uuid.UUID       `json:"user_id"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Status            VideoStatus     `json:"status"`
	Visibility        VideoVisibility `json:"visibility"`
	DurationSeconds   *int            `json:"duration_seconds,omitempty"`
	ThumbnailObjectKey *string        `json:"thumbnail_object_key,omitempty"`
	SourceObjectKey   string          `json:"source_object_key"`
	ProcessedObjectKey *string        `json:"processed_object_key,omitempty"`
	ViewsCount        int             `json:"views_count"`
	LikesCount        int             `json:"likes_count"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	PublishedAt       *time.Time      `json:"published_at,omitempty"`

	// Joined fields
	Tags     []string `json:"tags,omitempty"`
	Username string   `json:"username,omitempty"`
}

type VideoTag struct {
	ID      uuid.UUID `json:"id"`
	VideoID uuid.UUID `json:"video_id"`
	Tag     string    `json:"tag"`
}
