package dto

import "github.com/google/uuid"

type InitiateUploadRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Filename    string   `json:"filename"`
	ContentType string   `json:"content_type"`
	FileSize    int64    `json:"file_size"`
}

func (r InitiateUploadRequest) Validate() map[string]string {
	errs := map[string]string{}
	if len(r.Title) < 1 || len(r.Title) > 255 {
		errs["title"] = "must be between 1 and 255 characters"
	}
	if len(r.Description) > 5000 {
		errs["description"] = "must be at most 5000 characters"
	}
	if len(r.Tags) > 20 {
		errs["tags"] = "maximum 20 tags allowed"
	}
	for _, t := range r.Tags {
		if len(t) > 100 {
			errs["tags"] = "each tag must be at most 100 characters"
			break
		}
	}
	if r.Filename == "" {
		errs["filename"] = "required"
	}
	if r.ContentType == "" {
		errs["content_type"] = "required"
	}
	return errs
}

type InitiateUploadResponse struct {
	VideoID   uuid.UUID `json:"video_id"`
	UploadURL string    `json:"upload_url"`
	ObjectKey string    `json:"object_key"`
}

type CompleteUploadRequest struct {
	ObjectKey string `json:"object_key"`
}

type UpdateVideoRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Visibility  *string  `json:"visibility,omitempty"`
}

func (r UpdateVideoRequest) Validate() map[string]string {
	errs := map[string]string{}
	if r.Title != nil && (len(*r.Title) < 1 || len(*r.Title) > 255) {
		errs["title"] = "must be between 1 and 255 characters"
	}
	if r.Description != nil && len(*r.Description) > 5000 {
		errs["description"] = "must be at most 5000 characters"
	}
	if r.Visibility != nil {
		switch *r.Visibility {
		case "public", "unlisted", "private":
		default:
			errs["visibility"] = "must be public, unlisted, or private"
		}
	}
	return errs
}

type VideoResponse struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Username          string    `json:"username"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Status            string    `json:"status"`
	Visibility        string    `json:"visibility"`
	DurationSeconds   *int      `json:"duration_seconds,omitempty"`
	ThumbnailURL      string    `json:"thumbnail_url,omitempty"`
	ViewsCount        int       `json:"views_count"`
	LikesCount        int       `json:"likes_count"`
	Tags              []string  `json:"tags"`
	CreatedAt         string    `json:"created_at"`
	PublishedAt       *string   `json:"published_at,omitempty"`
	UserHasLiked      bool      `json:"user_has_liked"`
}

type PlaybackResponse struct {
	PlaybackURL string `json:"playback_url"`
	ContentType string `json:"content_type"`
}

type VideoListResponse struct {
	Videos     []VideoResponse `json:"videos"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
}
