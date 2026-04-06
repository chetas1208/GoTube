package dto

import "github.com/google/uuid"

type CreateCommentRequest struct {
	Body     string     `json:"body"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

func (r CreateCommentRequest) Validate() map[string]string {
	errs := map[string]string{}
	if len(r.Body) < 1 || len(r.Body) > 2000 {
		errs["body"] = "must be between 1 and 2000 characters"
	}
	return errs
}

type CommentResponse struct {
	ID        uuid.UUID  `json:"id"`
	VideoID   uuid.UUID  `json:"video_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Username  string     `json:"username"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Body      string     `json:"body"`
	CreatedAt string     `json:"created_at"`
}

type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
}
