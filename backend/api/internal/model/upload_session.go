package model

import (
	"time"

	"github.com/google/uuid"
)

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
