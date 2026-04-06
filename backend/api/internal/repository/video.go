package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoRepository struct {
	db *pgxpool.Pool
}

func NewVideoRepository(db *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(ctx context.Context, video *model.Video) error {
	query := `INSERT INTO videos (id, user_id, title, description, status, visibility, source_object_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query,
		video.ID, video.UserID, video.Title, video.Description, video.Status, video.Visibility,
		video.SourceObjectKey, video.CreatedAt, video.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create video: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Video, error) {
	video := &model.Video{}
	query := `SELECT v.id, v.user_id, v.title, v.description, v.status, v.visibility,
		v.duration_seconds, v.thumbnail_object_key, v.source_object_key, v.processed_object_key,
		v.views_count, v.likes_count, v.created_at, v.updated_at, v.published_at, u.username
		FROM videos v JOIN users u ON v.user_id = u.id WHERE v.id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&video.ID, &video.UserID, &video.Title, &video.Description, &video.Status, &video.Visibility,
		&video.DurationSeconds, &video.ThumbnailObjectKey, &video.SourceObjectKey, &video.ProcessedObjectKey,
		&video.ViewsCount, &video.LikesCount, &video.CreatedAt, &video.UpdatedAt, &video.PublishedAt, &video.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video by id: %w", err)
	}

	tags, err := r.GetTags(ctx, id)
	if err != nil {
		return nil, err
	}
	video.Tags = tags
	return video, nil
}

func (r *VideoRepository) GetTags(ctx context.Context, videoID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT tag FROM video_tags WHERE video_id = $1 ORDER BY tag`, videoID)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *VideoRepository) SetTags(ctx context.Context, videoID uuid.UUID, tags []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM video_tags WHERE video_id = $1`, videoID)
	if err != nil {
		return fmt.Errorf("delete old tags: %w", err)
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		_, err = tx.Exec(ctx, `INSERT INTO video_tags (video_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`, videoID, tag)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	return tx.Commit(ctx)
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

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM videos WHERE id = $1`, id)
	return err
}

func (r *VideoRepository) ListRecent(ctx context.Context, p dto.PaginationParams) ([]model.Video, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM videos WHERE status = 'ready' AND visibility = 'public'`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count videos: %w", err)
	}

	query := `SELECT v.id, v.user_id, v.title, v.description, v.status, v.visibility,
		v.duration_seconds, v.thumbnail_object_key, v.views_count, v.likes_count,
		v.created_at, v.published_at, u.username
		FROM videos v JOIN users u ON v.user_id = u.id
		WHERE v.status = 'ready' AND v.visibility = 'public'
		ORDER BY v.created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, p.PerPage, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list recent: %w", err)
	}
	defer rows.Close()
	return r.scanVideoList(ctx, rows, total)
}

func (r *VideoRepository) ListByUser(ctx context.Context, userID uuid.UUID, p dto.PaginationParams) ([]model.Video, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM videos WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user videos: %w", err)
	}

	query := `SELECT v.id, v.user_id, v.title, v.description, v.status, v.visibility,
		v.duration_seconds, v.thumbnail_object_key, v.views_count, v.likes_count,
		v.created_at, v.published_at, u.username
		FROM videos v JOIN users u ON v.user_id = u.id
		WHERE v.user_id = $1 ORDER BY v.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, p.PerPage, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list user videos: %w", err)
	}
	defer rows.Close()
	return r.scanVideoList(ctx, rows, total)
}

func (r *VideoRepository) Search(ctx context.Context, params dto.SearchParams) ([]model.Video, int, error) {
	likeQuery := "%" + strings.ToLower(params.Query) + "%"

	countQuery := `SELECT count(DISTINCT v.id) FROM videos v
		LEFT JOIN video_tags vt ON v.id = vt.video_id
		WHERE v.status = 'ready' AND v.visibility = 'public'
		AND (v.title ILIKE $1 OR vt.tag ILIKE $1)`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, likeQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("search count: %w", err)
	}

	orderClause := "ORDER BY v.created_at DESC"
	switch params.SortBy {
	case "views":
		orderClause = "ORDER BY v.views_count DESC"
	case "relevance":
		orderClause = "ORDER BY similarity(v.title, $1) DESC"
	}

	searchQuery := fmt.Sprintf(`SELECT DISTINCT v.id, v.user_id, v.title, v.description, v.status, v.visibility,
		v.duration_seconds, v.thumbnail_object_key, v.views_count, v.likes_count,
		v.created_at, v.published_at, u.username
		FROM videos v JOIN users u ON v.user_id = u.id
		LEFT JOIN video_tags vt ON v.id = vt.video_id
		WHERE v.status = 'ready' AND v.visibility = 'public'
		AND (v.title ILIKE $1 OR vt.tag ILIKE $1)
		%s LIMIT $2 OFFSET $3`, orderClause)

	rows, err := r.db.Query(ctx, searchQuery, likeQuery, params.PerPage, (params.Page-1)*params.PerPage)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()
	return r.scanVideoList(ctx, rows, total)
}

func (r *VideoRepository) Trending(ctx context.Context, limit int) ([]model.Video, error) {
	// Weighted score: views in past 7 days + 0.5 * likes_count
	// Recency factor applied via the time window on views
	query := `SELECT v.id, v.user_id, v.title, v.description, v.status, v.visibility,
		v.duration_seconds, v.thumbnail_object_key, v.views_count, v.likes_count,
		v.created_at, v.published_at, u.username,
		(SELECT count(*) FROM video_views vv WHERE vv.video_id = v.id AND vv.viewed_at > now() - interval '7 days') AS recent_views
		FROM videos v JOIN users u ON v.user_id = u.id
		WHERE v.status = 'ready' AND v.visibility = 'public'
		ORDER BY (SELECT count(*) FROM video_views vv WHERE vv.video_id = v.id AND vv.viewed_at > now() - interval '7 days') + v.likes_count * 0.5 DESC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("trending: %w", err)
	}
	defer rows.Close()

	var videos []model.Video
	for rows.Next() {
		var v model.Video
		var recentViews int
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.Status, &v.Visibility,
			&v.DurationSeconds, &v.ThumbnailObjectKey, &v.ViewsCount, &v.LikesCount,
			&v.CreatedAt, &v.PublishedAt, &v.Username, &recentViews)
		if err != nil {
			return nil, fmt.Errorf("scan trending video: %w", err)
		}
		tags, _ := r.GetTags(ctx, v.ID)
		v.Tags = tags
		videos = append(videos, v)
	}
	return videos, nil
}

func (r *VideoRepository) IncrementViews(ctx context.Context, videoID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE videos SET views_count = views_count + 1 WHERE id = $1`, videoID)
	return err
}

func (r *VideoRepository) IncrementLikes(ctx context.Context, videoID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE videos SET likes_count = likes_count + 1 WHERE id = $1`, videoID)
	return err
}

func (r *VideoRepository) DecrementLikes(ctx context.Context, videoID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE videos SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = $1`, videoID)
	return err
}

func (r *VideoRepository) scanVideoList(ctx context.Context, rows pgx.Rows, total int) ([]model.Video, int, error) {
	var videos []model.Video
	for rows.Next() {
		var v model.Video
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.Status, &v.Visibility,
			&v.DurationSeconds, &v.ThumbnailObjectKey, &v.ViewsCount, &v.LikesCount,
			&v.CreatedAt, &v.PublishedAt, &v.Username)
		if err != nil {
			return nil, 0, fmt.Errorf("scan video: %w", err)
		}
		tags, _ := r.GetTags(ctx, v.ID)
		v.Tags = tags
		videos = append(videos, v)
	}
	return videos, total, nil
}
