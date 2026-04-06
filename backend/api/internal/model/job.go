package model

import (
	"time"

	"github.com/google/uuid"
)

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
