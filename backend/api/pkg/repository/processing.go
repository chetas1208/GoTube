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

var ErrNotFound = errors.New("not found")

type VideoRepository struct {
	db *pgxpool.Pool
}

func NewVideoRepository(db *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Video, error) {
	video := &model.Video{}
	query := `SELECT id, user_id, title, description, status, visibility,
		duration_seconds, thumbnail_object_key, source_object_key, processed_object_key,
		views_count, likes_count, created_at, updated_at, published_at
		FROM videos WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&video.ID, &video.UserID, &video.Title, &video.Description, &video.Status, &video.Visibility,
		&video.DurationSeconds, &video.ThumbnailObjectKey, &video.SourceObjectKey, &video.ProcessedObjectKey,
		&video.ViewsCount, &video.LikesCount, &video.CreatedAt, &video.UpdatedAt, &video.PublishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video by id: %w", err)
	}
	return video, nil
}

func (r *VideoRepository) Update(ctx context.Context, video *model.Video) error {
	query := `UPDATE videos SET title=$2, description=$3, status=$4, visibility=$5,
		duration_seconds=$6, thumbnail_object_key=$7, processed_object_key=$8,
		views_count=$9, likes_count=$10, updated_at=now(), published_at=$11
		WHERE id=$1`
	_, err := r.db.Exec(ctx, query,
		video.ID, video.Title, video.Description, video.Status, video.Visibility,
		video.DurationSeconds, video.ThumbnailObjectKey, video.ProcessedObjectKey,
		video.ViewsCount, video.LikesCount, video.PublishedAt)
	if err != nil {
		return fmt.Errorf("update video: %w", err)
	}
	return nil
}

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.VideoProcessingJob, error) {
	job := &model.VideoProcessingJob{}
	query := `SELECT id, video_id, job_type, status, attempts, last_error, queued_at, started_at, finished_at, created_at, updated_at
		FROM video_processing_jobs WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.VideoID, &job.JobType, &job.Status, &job.Attempts, &job.LastError,
		&job.QueuedAt, &job.StartedAt, &job.FinishedAt, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}
	return job, nil
}

func (r *JobRepository) Update(ctx context.Context, job *model.VideoProcessingJob) error {
	query := `UPDATE video_processing_jobs SET status=$2, attempts=$3, last_error=$4, started_at=$5, finished_at=$6, updated_at=now()
		WHERE id=$1`
	_, err := r.db.Exec(ctx, query, job.ID, job.Status, job.Attempts, job.LastError, job.StartedAt, job.FinishedAt)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	return nil
}
