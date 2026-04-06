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

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, job *model.VideoProcessingJob) error {
	query := `INSERT INTO video_processing_jobs (id, video_id, job_type, status, attempts, queued_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query,
		job.ID, job.VideoID, job.JobType, job.Status, job.Attempts, job.QueuedAt, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
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
