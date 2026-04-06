package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadSessionRepository struct {
	db *pgxpool.Pool
}

func NewUploadSessionRepository(db *pgxpool.Pool) *UploadSessionRepository {
	return &UploadSessionRepository{db: db}
}

func (r *UploadSessionRepository) Create(ctx context.Context, session *model.UploadSession) error {
	query := `INSERT INTO upload_sessions (id, user_id, video_id, object_key, upload_status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query,
		session.ID, session.UserID, session.VideoID, session.ObjectKey, session.UploadStatus, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("create upload session: %w", err)
	}
	return nil
}

func (r *UploadSessionRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*model.UploadSession, error) {
	s := &model.UploadSession{}
	query := `SELECT id, user_id, video_id, object_key, upload_status, expires_at, created_at, completed_at
		FROM upload_sessions WHERE video_id = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRow(ctx, query, videoID).Scan(
		&s.ID, &s.UserID, &s.VideoID, &s.ObjectKey, &s.UploadStatus, &s.ExpiresAt, &s.CreatedAt, &s.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get upload session: %w", err)
	}
	return s, nil
}

func (r *UploadSessionRepository) Update(ctx context.Context, session *model.UploadSession) error {
	query := `UPDATE upload_sessions SET upload_status=$2, completed_at=$3 WHERE id=$1`
	_, err := r.db.Exec(ctx, query, session.ID, session.UploadStatus, session.CompletedAt)
	if err != nil {
		return fmt.Errorf("update upload session: %w", err)
	}
	return nil
}
