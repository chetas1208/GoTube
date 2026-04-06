package processor

import (
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chetasparekh/gotube-lite/api/pkg/metrics"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/chetasparekh/gotube-lite/api/pkg/repository"
	"github.com/chetasparekh/gotube-lite/api/pkg/storage"
	workerCfg "github.com/chetasparekh/gotube-lite/worker/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type VideoProcessor struct {
	videoRepo *repository.VideoRepository
	jobRepo   *repository.JobRepository
	storage   storage.ObjectStorage
	cfg       *workerCfg.Config
}

func NewVideoProcessor(
	videoRepo *repository.VideoRepository,
	jobRepo *repository.JobRepository,
	store storage.ObjectStorage,
	cfg *workerCfg.Config,
) *VideoProcessor {
	return &VideoProcessor{
		videoRepo: videoRepo,
		jobRepo:   jobRepo,
		storage:   store,
		cfg:       cfg,
	}
}

func (p *VideoProcessor) Process(ctx context.Context, jobID, videoID uuid.UUID) error {
	start := time.Now()
	logger := log.With().Str("job_id", jobID.String()).Str("video_id", videoID.String()).Logger()

	// Fetch job record
	job, err := p.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}

	// Idempotency: skip if already completed
	if job.Status == model.JobStatusCompleted {
		logger.Info().Msg("job already completed, skipping")
		return nil
	}

	// Check retry limit
	if job.Attempts >= p.cfg.MaxRetries {
		logger.Warn().Int("attempts", job.Attempts).Msg("max retries exceeded")
		p.failJob(ctx, job, "max retries exceeded")
		p.failVideo(ctx, videoID)
		return nil
	}

	// Mark as running
	now := time.Now()
	job.Status = model.JobStatusRunning
	job.StartedAt = &now
	job.Attempts++
	_ = p.jobRepo.Update(ctx, job)

	// Update video status
	video, err := p.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		p.failJob(ctx, job, fmt.Sprintf("get video: %v", err))
		return fmt.Errorf("get video: %w", err)
	}
	video.Status = model.VideoStatusProcessing
	_ = p.videoRepo.Update(ctx, video)

	logger.Info().Str("source_key", video.SourceObjectKey).Msg("starting video processing")

	// Create temp directory for this job
	workDir := filepath.Join(p.cfg.TempDir, jobID.String())
	if err := os.MkdirAll(workDir, 0755); err != nil {
		p.failJob(ctx, job, fmt.Sprintf("create work dir: %v", err))
		return fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Download source from object storage
	sourceFile := filepath.Join(workDir, "source"+filepath.Ext(video.SourceObjectKey))
	if err := p.downloadObject(ctx, p.cfg.Storage.BucketRaw, video.SourceObjectKey, sourceFile); err != nil {
		p.failJob(ctx, job, fmt.Sprintf("download source: %v", err))
		p.failVideo(ctx, videoID)
		metrics.VideosProcessedTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("download source: %w", err)
	}
	logger.Info().Msg("source downloaded")

	duration := p.extractDuration(ctx, sourceFile)

	var thumbnailKey *string
	if key, err := p.generateBestEffortThumbnail(ctx, video, sourceFile, workDir, duration); err != nil {
		logger.Warn().Err(err).Msg("early thumbnail generation failed, falling back later")
	} else {
		thumbnailKey = &key
		logger.Info().Str("key", key).Msg("early thumbnail uploaded")
	}

	// Transcode to optimized MP4
	processedFile := filepath.Join(workDir, "video.mp4")
	if err := p.transcode(ctx, sourceFile, processedFile); err != nil {
		p.failJob(ctx, job, fmt.Sprintf("transcode: %v", err))
		p.failVideo(ctx, videoID)
		metrics.VideosProcessedTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("transcode: %w", err)
	}
	logger.Info().Msg("transcode completed")

	// Upload processed video to object storage
	processedKey := fmt.Sprintf("processed/%s/%s/video.mp4", video.UserID, videoID)
	if err := p.uploadObject(ctx, p.cfg.Storage.BucketProcessed, processedKey, processedFile, "video/mp4"); err != nil {
		p.failJob(ctx, job, fmt.Sprintf("upload processed: %v", err))
		p.failVideo(ctx, videoID)
		metrics.VideosProcessedTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("upload processed: %w", err)
	}
	logger.Info().Str("key", processedKey).Msg("processed video uploaded")

	if thumbnailKey == nil {
		key, err := p.generateFallbackThumbnail(ctx, video, sourceFile, workDir, duration)
		if err != nil {
			logger.Warn().Err(err).Msg("fallback thumbnail generation failed")
		} else {
			thumbnailKey = &key
			logger.Info().Str("key", key).Msg("fallback thumbnail uploaded")
		}
	}

	// Update video record
	video.Status = model.VideoStatusReady
	video.ProcessedObjectKey = &processedKey
	video.ThumbnailObjectKey = thumbnailKey
	if duration > 0 {
		video.DurationSeconds = &duration
	}
	publishTime := time.Now()
	video.PublishedAt = &publishTime
	if err := p.videoRepo.Update(ctx, video); err != nil {
		p.failJob(ctx, job, fmt.Sprintf("update video: %v", err))
		return fmt.Errorf("update video: %w", err)
	}

	// Mark job complete
	finishedAt := time.Now()
	job.Status = model.JobStatusCompleted
	job.FinishedAt = &finishedAt
	_ = p.jobRepo.Update(ctx, job)

	elapsed := time.Since(start).Seconds()
	metrics.VideosProcessedTotal.WithLabelValues("success").Inc()
	metrics.VideoProcessingDuration.Observe(elapsed)

	logger.Info().Float64("elapsed_seconds", elapsed).Msg("video processing completed")
	return nil
}

// transcode converts source video to an optimized MP4.
// Uses H.264 + AAC with CRF-based quality control and faststart for progressive playback.
func (p *VideoProcessor) transcode(ctx context.Context, input, output string) error {
	args := []string{
		"-i", input,
		"-c:v", "libx264",
		"-preset", p.cfg.FFmpeg.Preset,
		"-crf", p.cfg.FFmpeg.CRF,
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		output,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *VideoProcessor) generateThumbnail(ctx context.Context, input, output, timestamp string) error {
	args := []string{
		"-i", input,
		"-ss", timestamp,
		"-vframes", "1",
		"-vf", "scale=640:-1",
		"-q:v", "3",
		"-y",
		output,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *VideoProcessor) extractDuration(ctx context.Context, input string) int {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	str := strings.TrimSpace(string(out))
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return int(f)
}

func (p *VideoProcessor) generateBestEffortThumbnail(ctx context.Context, video *model.Video, input, workDir string, duration int) (string, error) {
	candidateSeconds := thumbnailCandidateSeconds(duration)
	if len(candidateSeconds) == 0 {
		return "", fmt.Errorf("no thumbnail candidates available")
	}

	bestPath := ""
	bestScore := math.Inf(-1)

	for i, second := range candidateSeconds {
		candidatePath := filepath.Join(workDir, fmt.Sprintf("thumb-candidate-%02d.jpg", i))
		if err := p.generateThumbnail(ctx, input, candidatePath, formatTimestamp(second)); err != nil {
			continue
		}

		score, err := scoreThumbnail(candidatePath)
		if err != nil {
			continue
		}

		if score > bestScore {
			bestScore = score
			bestPath = candidatePath
		}
	}

	if bestPath == "" {
		return "", fmt.Errorf("no valid thumbnail candidates generated")
	}

	return p.persistThumbnail(ctx, video, bestPath)
}

func (p *VideoProcessor) generateFallbackThumbnail(ctx context.Context, video *model.Video, input, workDir string, duration int) (string, error) {
	thumbnailFile := filepath.Join(workDir, "thumb-fallback.jpg")
	if err := p.generateThumbnail(ctx, input, thumbnailFile, fallbackThumbnailTimestamp(duration)); err != nil {
		return "", err
	}
	return p.persistThumbnail(ctx, video, thumbnailFile)
}

func (p *VideoProcessor) persistThumbnail(ctx context.Context, video *model.Video, thumbnailPath string) (string, error) {
	key := fmt.Sprintf("thumbnails/%s/%s/thumb.jpg", video.UserID, video.ID)
	if err := p.uploadObject(ctx, p.cfg.Storage.BucketThumbnails, key, thumbnailPath, "image/jpeg"); err != nil {
		return "", err
	}
	video.ThumbnailObjectKey = &key
	if err := p.videoRepo.Update(ctx, video); err != nil {
		return "", err
	}
	return key, nil
}

func thumbnailCandidateSeconds(duration int) []int {
	fractions := []float64{0.10, 0.20, 0.35, 0.50, 0.65}
	maxSecond := duration - 1
	if maxSecond < 1 {
		maxSecond = 1
	}

	seen := make(map[int]struct{}, len(fractions))
	seconds := make([]int, 0, len(fractions))
	for _, fraction := range fractions {
		second := int(math.Round(float64(duration) * fraction))
		if second < 1 {
			second = 1
		}
		if second > maxSecond {
			second = maxSecond
		}
		if _, ok := seen[second]; ok {
			continue
		}
		seen[second] = struct{}{}
		seconds = append(seconds, second)
	}

	if len(seconds) == 0 {
		return []int{1}
	}
	return seconds
}

func fallbackThumbnailTimestamp(duration int) string {
	second := 1
	if duration > 4 {
		second = int(float64(duration) * 0.25)
		if second < 1 {
			second = 1
		}
	}
	return formatTimestamp(second)
}

func formatTimestamp(second int) string {
	if second < 1 {
		second = 1
	}
	hours := second / 3600
	minutes := (second % 3600) / 60
	seconds := second % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func scoreThumbnail(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		return 0, fmt.Errorf("empty thumbnail image")
	}

	sampleWidth := bounds.Dx()
	if sampleWidth > 128 {
		sampleWidth = 128
	}
	sampleHeight := bounds.Dy()
	if sampleHeight > 72 {
		sampleHeight = 72
	}

	var sum float64
	var sumSquares float64
	var edgeSum float64
	prevRow := make([]float64, sampleWidth)

	for y := 0; y < sampleHeight; y++ {
		sourceY := bounds.Min.Y + (y*bounds.Dy())/sampleHeight
		prevValue := 0.0
		for x := 0; x < sampleWidth; x++ {
			sourceX := bounds.Min.X + (x*bounds.Dx())/sampleWidth
			value := luminance(img.At(sourceX, sourceY))
			sum += value
			sumSquares += value * value

			if x > 0 {
				edgeSum += math.Abs(value - prevValue)
			}
			if y > 0 {
				edgeSum += math.Abs(value - prevRow[x])
			}

			prevRow[x] = value
			prevValue = value
		}
	}

	pixels := float64(sampleWidth * sampleHeight)
	mean := (sum / pixels) / 255.0
	variance := (sumSquares / pixels) - math.Pow(sum/pixels, 2)
	if variance < 0 {
		variance = 0
	}

	contrast := math.Sqrt(variance) / 255.0
	edgeDetail := edgeSum / (pixels * 255.0)
	exposure := 1 - math.Min(1, math.Abs(mean-0.5)/0.5)

	score := edgeDetail*0.55 + contrast*0.25 + exposure*0.20
	if mean < 0.12 || mean > 0.88 {
		score -= 0.40
	}
	if contrast < 0.05 {
		score -= 0.25
	}

	return score, nil
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)
}

func (p *VideoProcessor) downloadObject(ctx context.Context, bucket, key, destPath string) error {
	reader, err := p.storage.GetObject(ctx, bucket, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (p *VideoProcessor) uploadObject(ctx context.Context, bucket, key, srcPath, contentType string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return p.storage.PutObject(ctx, bucket, key, file, contentType)
}

func (p *VideoProcessor) failJob(ctx context.Context, job *model.VideoProcessingJob, errMsg string) {
	job.Status = model.JobStatusFailed
	job.LastError = &errMsg
	now := time.Now()
	job.FinishedAt = &now
	_ = p.jobRepo.Update(ctx, job)
}

func (p *VideoProcessor) failVideo(ctx context.Context, videoID uuid.UUID) {
	video, err := p.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return
	}
	video.Status = model.VideoStatusFailed
	_ = p.videoRepo.Update(ctx, video)
}
