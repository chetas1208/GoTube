package model

import (
	"time"

	"github.com/google/uuid"
)

type VideoView struct {
	ID           uuid.UUID  `json:"id"`
	VideoID      uuid.UUID  `json:"video_id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	SessionID    *string    `json:"session_id,omitempty"`
	ViewedAt     time.Time  `json:"viewed_at"`
	WatchSeconds *int       `json:"watch_seconds,omitempty"`
}
