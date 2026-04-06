package consumer

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/chetasparekh/gotube-lite/api/pkg/metrics"
	"github.com/chetasparekh/gotube-lite/api/pkg/queue"
	"github.com/chetasparekh/gotube-lite/worker/internal/processor"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	queue       *queue.Queue
	processor   *processor.VideoProcessor
	concurrency int
	consumerID  string
}

func NewConsumer(q *queue.Queue, proc *processor.VideoProcessor, concurrency int) *Consumer {
	hostname, _ := os.Hostname()
	return &Consumer{
		queue:       q,
		processor:   proc,
		concurrency: concurrency,
		consumerID:  fmt.Sprintf("worker-%s-%d", hostname, os.Getpid()),
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	log.Info().
		Str("consumer_id", c.consumerID).
		Int("concurrency", c.concurrency).
		Msg("starting consumer")

	var wg sync.WaitGroup
	sem := make(chan struct{}, c.concurrency)

	// Periodically reclaim stale messages
	go c.reclaimLoop(ctx)

	// Periodically report queue depth
	go c.metricsLoop(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("consumer shutting down, waiting for in-flight jobs...")
			wg.Wait()
			return nil
		default:
		}

		messages, err := c.queue.Consume(ctx, c.consumerID, 1, 5*time.Second)
		if err != nil {
			log.Error().Err(err).Msg("consume error")
			time.Sleep(2 * time.Second)
			continue
		}

		for _, msg := range messages {
			jobMsg, err := c.queue.ParseMessage(msg)
			if err != nil {
				log.Error().Err(err).Str("msg_id", msg.ID).Msg("failed to parse message")
				_ = c.queue.Ack(ctx, msg.ID)
				continue
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(msgID string, job *queue.JobMessage) {
				defer wg.Done()
				defer func() { <-sem }()

				logger := log.With().
					Str("job_id", job.JobID.String()).
					Str("video_id", job.VideoID.String()).
					Logger()

				logger.Info().Msg("processing job")
				if err := c.processor.Process(ctx, job.JobID, job.VideoID); err != nil {
					logger.Error().Err(err).Msg("job processing failed")
				} else {
					logger.Info().Msg("job processing succeeded")
				}

				if err := c.queue.Ack(ctx, msgID); err != nil {
					logger.Error().Err(err).Msg("failed to ack message")
				}
			}(msg.ID, jobMsg)
		}
	}
}

func (c *Consumer) reclaimLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := c.queue.ReclaimStale(ctx, c.consumerID, 5*time.Minute, 10)
			if err != nil {
				log.Warn().Err(err).Msg("reclaim error")
				continue
			}
			for _, msg := range messages {
				jobMsg, err := c.queue.ParseMessage(msg)
				if err != nil {
					_ = c.queue.Ack(ctx, msg.ID)
					continue
				}
				log.Info().Str("job_id", jobMsg.JobID.String()).Msg("reclaiming stale job")
				if err := c.processor.Process(ctx, jobMsg.JobID, jobMsg.VideoID); err != nil {
					log.Error().Err(err).Msg("reclaimed job failed")
				}
				_ = c.queue.Ack(ctx, msg.ID)
			}
		}
	}
}

func (c *Consumer) metricsLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			depth, err := c.queue.Len(ctx)
			if err == nil {
				metrics.QueueDepth.Set(float64(depth))
			}
		}
	}
}
