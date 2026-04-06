package repository

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ViewRepository struct {
	db *pgxpool.Pool
}

func NewViewRepository(db *pgxpool.Pool) *ViewRepository {
	return &ViewRepository{db: db}
}

func (r *ViewRepository) Create(ctx context.Context, view *model.VideoView) error {
	query := `INSERT INTO video_views (id, video_id, user_id, session_id, viewed_at, watch_seconds)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, view.ID, view.VideoID, view.UserID, view.SessionID, view.ViewedAt, view.WatchSeconds)
	if err != nil {
		return fmt.Errorf("create view: %w", err)
	}
	return nil
}

func (r *ViewRepository) CreateUniqueAndIncrement(ctx context.Context, view *model.VideoView, dedupeWindow time.Duration) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin view tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, viewLockKey(view.VideoID, view.UserID, view.SessionID)); err != nil {
		return false, fmt.Errorf("lock unique view: %w", err)
	}

	if dedupeWindow > 0 {
		exists, err := hasRecentViewTx(ctx, tx, view.VideoID, view.UserID, view.SessionID, view.ViewedAt.Add(-dedupeWindow))
		if err != nil {
			return false, err
		}
		if exists {
			if err := tx.Commit(ctx); err != nil {
				return false, fmt.Errorf("commit duplicate view tx: %w", err)
			}
			return false, nil
		}
	}

	query := `INSERT INTO video_views (id, video_id, user_id, session_id, viewed_at, watch_seconds)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, query, view.ID, view.VideoID, view.UserID, view.SessionID, view.ViewedAt, view.WatchSeconds); err != nil {
		return false, fmt.Errorf("create unique view: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE videos SET views_count = views_count + 1 WHERE id = $1`, view.VideoID); err != nil {
		return false, fmt.Errorf("increment views: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit view tx: %w", err)
	}

	return true, nil
}

func hasRecentViewTx(ctx context.Context, tx pgx.Tx, videoID uuid.UUID, userID *uuid.UUID, sessionID *string, since time.Time) (bool, error) {
	var exists bool

	switch {
	case userID != nil:
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM video_views
				WHERE video_id = $1 AND user_id = $2 AND viewed_at >= $3
			)
		`, videoID, *userID, since).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("check recent user view: %w", err)
		}
		return exists, nil
	case sessionID != nil && *sessionID != "":
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1
				FROM video_views
				WHERE video_id = $1 AND session_id = $2 AND viewed_at >= $3
			)
		`, videoID, *sessionID, since).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("check recent session view: %w", err)
		}
		return exists, nil
	default:
		return false, nil
	}
}

func viewLockKey(videoID uuid.UUID, userID *uuid.UUID, sessionID *string) int64 {
	viewerKey := "anonymous"
	switch {
	case userID != nil:
		viewerKey = "user:" + userID.String()
	case sessionID != nil && *sessionID != "":
		viewerKey = "session:" + *sessionID
	}

	hash := fnv.New64a()
	_, _ = hash.Write([]byte(videoID.String()))
	_, _ = hash.Write([]byte("|"))
	_, _ = hash.Write([]byte(viewerKey))
	return int64(hash.Sum64())
}
