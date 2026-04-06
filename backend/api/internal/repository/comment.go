package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `INSERT INTO comments (id, video_id, user_id, parent_id, body, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query,
		comment.ID, comment.VideoID, comment.UserID, comment.ParentID, comment.Body, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID, p dto.PaginationParams) ([]model.Comment, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM comments WHERE video_id = $1`, videoID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	query := `SELECT c.id, c.video_id, c.user_id, c.parent_id, c.body, c.created_at, c.updated_at, u.username
		FROM comments c JOIN users u ON c.user_id = u.id
		WHERE c.video_id = $1
		ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, videoID, p.PerPage, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("get comments: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.VideoID, &c.UserID, &c.ParentID, &c.Body, &c.CreatedAt, &c.UpdatedAt, &c.Username); err != nil {
			return nil, 0, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, total, nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	c := &model.Comment{}
	query := `SELECT c.id, c.video_id, c.user_id, c.parent_id, c.body, c.created_at, c.updated_at, u.username
		FROM comments c JOIN users u ON c.user_id = u.id WHERE c.id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.VideoID, &c.UserID, &c.ParentID, &c.Body, &c.CreatedAt, &c.UpdatedAt, &c.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	return c, nil
}
