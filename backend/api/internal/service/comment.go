package service

import (
	"context"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/repository"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
)

type CommentService struct {
	comments *repository.CommentRepository
}

func NewCommentService(comments *repository.CommentRepository) *CommentService {
	return &CommentService{comments: comments}
}

func (s *CommentService) Create(ctx context.Context, videoID, userID uuid.UUID, req dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	now := time.Now()
	comment := &model.Comment{
		ID:        uuid.New(),
		VideoID:   videoID,
		UserID:    userID,
		ParentID:  req.ParentID,
		Body:      req.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, err
	}

	created, err := s.comments.GetByID(ctx, comment.ID)
	if err != nil {
		return nil, err
	}

	return &dto.CommentResponse{
		ID:        created.ID,
		VideoID:   created.VideoID,
		UserID:    created.UserID,
		Username:  created.Username,
		ParentID:  created.ParentID,
		Body:      created.Body,
		CreatedAt: created.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CommentService) ListByVideo(ctx context.Context, videoID uuid.UUID, p dto.PaginationParams) (*dto.CommentListResponse, error) {
	comments, total, err := s.comments.GetByVideoID(ctx, videoID, p)
	if err != nil {
		return nil, err
	}

	var items []dto.CommentResponse
	for _, c := range comments {
		items = append(items, dto.CommentResponse{
			ID:        c.ID,
			VideoID:   c.VideoID,
			UserID:    c.UserID,
			Username:  c.Username,
			ParentID:  c.ParentID,
			Body:      c.Body,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}
	if items == nil {
		items = []dto.CommentResponse{}
	}

	return &dto.CommentListResponse{
		Comments:   items,
		TotalCount: total,
		Page:       p.Page,
		PerPage:    p.PerPage,
	}, nil
}
