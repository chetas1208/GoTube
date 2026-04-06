package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/config"
	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/repository"
	"github.com/chetasparekh/gotube-lite/api/pkg/metrics"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/chetasparekh/gotube-lite/api/pkg/queue"
	"github.com/chetasparekh/gotube-lite/api/pkg/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	ErrForbidden     = errors.New("forbidden")
	ErrVideoNotReady = errors.New("video not ready for playback")
)

const viewDedupeWindow = 30 * time.Minute

type VideoService struct {
	videos     *repository.VideoRepository
	uploads    *repository.UploadSessionRepository
	jobs       *repository.JobRepository
	likes      *repository.LikeRepository
	views      *repository.ViewRepository
	storage    storage.ObjectStorage
	queue      *queue.Queue
	storageCfg config.StorageConfig
	uploadCfg  config.UploadConfig
}

func NewVideoService(
	videos *repository.VideoRepository,
	uploads *repository.UploadSessionRepository,
	jobs *repository.JobRepository,
	likes *repository.LikeRepository,
	views *repository.ViewRepository,
	store storage.ObjectStorage,
	q *queue.Queue,
	storageCfg config.StorageConfig,
	uploadCfg config.UploadConfig,
) *VideoService {
	return &VideoService{
		videos:     videos,
		uploads:    uploads,
		jobs:       jobs,
		likes:      likes,
		views:      views,
		storage:    store,
		queue:      q,
		storageCfg: storageCfg,
		uploadCfg:  uploadCfg,
	}
}

func (s *VideoService) InitiateUpload(ctx context.Context, userID uuid.UUID, req dto.InitiateUploadRequest) (*dto.InitiateUploadResponse, error) {
	if !s.isAllowedContentType(req.ContentType) {
		return nil, fmt.Errorf("unsupported content type: %s", req.ContentType)
	}
	maxBytes := s.uploadCfg.MaxSizeMB * 1024 * 1024
	if req.FileSize > maxBytes {
		return nil, fmt.Errorf("file too large: max %dMB", s.uploadCfg.MaxSizeMB)
	}

	videoID := uuid.New()
	ext := filepath.Ext(req.Filename)
	objectKey := fmt.Sprintf("raw/%s/%s/original%s", userID, videoID, ext)

	now := time.Now()
	video := &model.Video{
		ID:              videoID,
		UserID:          userID,
		Title:           req.Title,
		Description:     req.Description,
		Status:          model.VideoStatusUploaded,
		Visibility:      model.VisibilityPublic,
		SourceObjectKey: objectKey,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.videos.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("create video record: %w", err)
	}

	if len(req.Tags) > 0 {
		if err := s.videos.SetTags(ctx, videoID, req.Tags); err != nil {
			log.Warn().Err(err).Msg("failed to set tags during upload initiation")
		}
	}

	session := &model.UploadSession{
		ID:           uuid.New(),
		UserID:       userID,
		VideoID:      videoID,
		ObjectKey:    objectKey,
		UploadStatus: model.UploadStatusInitiated,
		ExpiresAt:    now.Add(1 * time.Hour),
		CreatedAt:    now,
	}
	if err := s.uploads.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create upload session: %w", err)
	}

	uploadURL, err := s.storage.GeneratePresignedPutURL(ctx, s.storageCfg.BucketRaw, objectKey, req.ContentType, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generate presigned URL: %w", err)
	}

	metrics.UploadsTotal.WithLabelValues("initiated").Inc()

	return &dto.InitiateUploadResponse{
		VideoID:   videoID,
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}, nil
}

func (s *VideoService) CompleteUpload(ctx context.Context, userID, videoID uuid.UUID) error {
	video, err := s.videos.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video.UserID != userID {
		return ErrForbidden
	}
	if video.Status != model.VideoStatusUploaded {
		return fmt.Errorf("video not in uploaded state")
	}

	// Verify object exists in object storage before queueing work.
	_, err = s.storage.HeadObject(ctx, s.storageCfg.BucketRaw, video.SourceObjectKey)
	if err != nil {
		return fmt.Errorf("source object not found in storage: %w", err)
	}

	session, err := s.uploads.GetByVideoID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("get upload session: %w", err)
	}
	now := time.Now()
	session.UploadStatus = model.UploadStatusCompleted
	session.CompletedAt = &now
	if err := s.uploads.Update(ctx, session); err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	video.Status = model.VideoStatusQueued
	if err := s.videos.Update(ctx, video); err != nil {
		return fmt.Errorf("update video status: %w", err)
	}

	jobID := uuid.New()
	job := &model.VideoProcessingJob{
		ID:        jobID,
		VideoID:   videoID,
		JobType:   model.JobTypeTranscode,
		Status:    model.JobStatusPending,
		Attempts:  0,
		QueuedAt:  now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.jobs.Create(ctx, job); err != nil {
		return fmt.Errorf("create processing job: %w", err)
	}

	if err := s.queue.Enqueue(ctx, queue.JobMessage{
		JobID:   jobID,
		VideoID: videoID,
		JobType: string(model.JobTypeTranscode),
	}); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}

	metrics.UploadsTotal.WithLabelValues("completed").Inc()
	log.Info().Str("video_id", videoID.String()).Msg("upload completed, job enqueued")
	return nil
}

func (s *VideoService) GetVideo(ctx context.Context, videoID uuid.UUID, viewerID *uuid.UUID) (*dto.VideoResponse, error) {
	video, err := s.videos.GetByID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return s.toVideoResponse(ctx, video, viewerID)
}

func (s *VideoService) UpdateVideo(ctx context.Context, userID, videoID uuid.UUID, req dto.UpdateVideoRequest) (*dto.VideoResponse, error) {
	video, err := s.videos.GetByID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	if video.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if req.Visibility != nil {
		video.Visibility = model.VideoVisibility(*req.Visibility)
	}
	if err := s.videos.Update(ctx, video); err != nil {
		return nil, err
	}
	if req.Tags != nil {
		if err := s.videos.SetTags(ctx, videoID, req.Tags); err != nil {
			return nil, err
		}
		video.Tags = req.Tags
	}
	return s.toVideoResponse(ctx, video, &userID)
}

func (s *VideoService) DeleteVideo(ctx context.Context, userID, videoID uuid.UUID) error {
	video, err := s.videos.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video.UserID != userID {
		return ErrForbidden
	}
	// Clean up storage objects (best effort)
	_ = s.storage.DeleteObject(ctx, s.storageCfg.BucketRaw, video.SourceObjectKey)
	if video.ProcessedObjectKey != nil {
		_ = s.storage.DeleteObject(ctx, s.storageCfg.BucketProcessed, *video.ProcessedObjectKey)
	}
	if video.ThumbnailObjectKey != nil {
		_ = s.storage.DeleteObject(ctx, s.storageCfg.BucketThumbnails, *video.ThumbnailObjectKey)
	}
	return s.videos.Delete(ctx, videoID)
}

func (s *VideoService) ListRecent(ctx context.Context, p dto.PaginationParams, viewerID *uuid.UUID) (*dto.VideoListResponse, error) {
	videos, total, err := s.videos.ListRecent(ctx, p)
	if err != nil {
		return nil, err
	}
	return s.toVideoListResponse(ctx, videos, total, p, viewerID)
}

func (s *VideoService) ListByUser(ctx context.Context, userID uuid.UUID, p dto.PaginationParams) (*dto.VideoListResponse, error) {
	videos, total, err := s.videos.ListByUser(ctx, userID, p)
	if err != nil {
		return nil, err
	}
	return s.toVideoListResponse(ctx, videos, total, p, &userID)
}

func (s *VideoService) GetPlaybackURL(ctx context.Context, videoID uuid.UUID) (*dto.PlaybackResponse, error) {
	video, err := s.videos.GetByID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	if video.Status != model.VideoStatusReady || video.ProcessedObjectKey == nil {
		return nil, ErrVideoNotReady
	}

	url, err := s.storage.GeneratePresignedGetURL(ctx, s.storageCfg.BucketProcessed, *video.ProcessedObjectKey, 2*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("generate playback URL: %w", err)
	}
	return &dto.PlaybackResponse{PlaybackURL: url, ContentType: "video/mp4"}, nil
}

func (s *VideoService) RecordView(ctx context.Context, videoID uuid.UUID, userID *uuid.UUID, sessionID *string) error {
	view := &model.VideoView{
		ID:        uuid.New(),
		VideoID:   videoID,
		UserID:    userID,
		SessionID: sessionID,
		ViewedAt:  time.Now(),
	}
	if _, err := s.views.CreateUniqueAndIncrement(ctx, view, viewDedupeWindow); err != nil {
		return err
	}
	return nil
}

func (s *VideoService) ToggleLike(ctx context.Context, videoID, userID uuid.UUID) (bool, error) {
	liked, err := s.likes.HasLiked(ctx, videoID, userID)
	if err != nil {
		return false, err
	}
	if liked {
		if err := s.likes.Delete(ctx, videoID, userID); err != nil {
			return false, err
		}
		_ = s.videos.DecrementLikes(ctx, videoID)
		return false, nil
	}
	like := &model.Like{
		ID:        uuid.New(),
		VideoID:   videoID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	if err := s.likes.Create(ctx, like); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return true, nil
		}
		return false, err
	}
	_ = s.videos.IncrementLikes(ctx, videoID)
	return true, nil
}

func (s *VideoService) RemoveLike(ctx context.Context, videoID, userID uuid.UUID) error {
	if err := s.likes.Delete(ctx, videoID, userID); err != nil {
		return err
	}
	return s.videos.DecrementLikes(ctx, videoID)
}

func (s *VideoService) Search(ctx context.Context, params dto.SearchParams, viewerID *uuid.UUID) (*dto.VideoListResponse, error) {
	videos, total, err := s.videos.Search(ctx, params)
	if err != nil {
		return nil, err
	}
	p := dto.PaginationParams{Page: params.Page, PerPage: params.PerPage}
	return s.toVideoListResponse(ctx, videos, total, p, viewerID)
}

func (s *VideoService) Trending(ctx context.Context, limit int, viewerID *uuid.UUID) ([]dto.VideoResponse, error) {
	videos, err := s.videos.Trending(ctx, limit)
	if err != nil {
		return nil, err
	}
	var result []dto.VideoResponse
	for _, v := range videos {
		vCopy := v
		r, err := s.toVideoResponse(ctx, &vCopy, viewerID)
		if err != nil {
			continue
		}
		result = append(result, *r)
	}
	return result, nil
}

func (s *VideoService) GetThumbnailURL(ctx context.Context, key string) (string, error) {
	url, err := s.storage.GeneratePresignedGetURL(ctx, s.storageCfg.BucketThumbnails, key, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("generate thumbnail URL: %w", err)
	}
	return url, nil
}

func (s *VideoService) toVideoResponse(ctx context.Context, v *model.Video, viewerID *uuid.UUID) (*dto.VideoResponse, error) {
	resp := &dto.VideoResponse{
		ID:              v.ID,
		UserID:          v.UserID,
		Username:        v.Username,
		Title:           v.Title,
		Description:     v.Description,
		Status:          string(v.Status),
		Visibility:      string(v.Visibility),
		DurationSeconds: v.DurationSeconds,
		ViewsCount:      v.ViewsCount,
		LikesCount:      v.LikesCount,
		Tags:            v.Tags,
		CreatedAt:       v.CreatedAt.Format(time.RFC3339),
	}
	if v.PublishedAt != nil {
		t := v.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &t
	}
	if v.ThumbnailObjectKey != nil {
		url, err := s.GetThumbnailURL(ctx, *v.ThumbnailObjectKey)
		if err == nil {
			resp.ThumbnailURL = url
		}
	}
	if resp.Tags == nil {
		resp.Tags = []string{}
	}
	if viewerID != nil {
		liked, _ := s.likes.HasLiked(ctx, v.ID, *viewerID)
		resp.UserHasLiked = liked
	}
	return resp, nil
}

func (s *VideoService) toVideoListResponse(ctx context.Context, videos []model.Video, total int, p dto.PaginationParams, viewerID *uuid.UUID) (*dto.VideoListResponse, error) {
	var items []dto.VideoResponse
	for _, v := range videos {
		vCopy := v
		r, err := s.toVideoResponse(ctx, &vCopy, viewerID)
		if err != nil {
			continue
		}
		items = append(items, *r)
	}
	if items == nil {
		items = []dto.VideoResponse{}
	}
	return &dto.VideoListResponse{
		Videos:     items,
		TotalCount: total,
		Page:       p.Page,
		PerPage:    p.PerPage,
	}, nil
}

func (s *VideoService) isAllowedContentType(ct string) bool {
	allowed := strings.Split(s.uploadCfg.AllowedVideoTypes, ",")
	for _, a := range allowed {
		if strings.TrimSpace(a) == ct {
			return true
		}
	}
	return false
}
