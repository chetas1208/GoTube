package repository

import (
	"context"
	"fmt"

	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LikeRepository struct {
	db *pgxpool.Pool
}

func NewLikeRepository(db *pgxpool.Pool) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Create(ctx context.Context, like *model.Like) error {
	query := `INSERT INTO likes (id, video_id, user_id, created_at) VALUES ($1, $2, $3, $4) ON CONFLICT (video_id, user_id) DO NOTHING`
	tag, err := r.db.Exec(ctx, query, like.ID, like.VideoID, like.UserID, like.CreatedAt)
	if err != nil {
		return fmt.Errorf("create like: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDuplicate
	}
	return nil
}

func (r *LikeRepository) Delete(ctx context.Context, videoID, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM likes WHERE video_id = $1 AND user_id = $2`, videoID, userID)
	if err != nil {
		return fmt.Errorf("delete like: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *LikeRepository) HasLiked(ctx context.Context, videoID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM likes WHERE video_id = $1 AND user_id = $2)`, videoID, userID).Scan(&exists)
	return exists, err
}
