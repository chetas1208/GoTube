package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	StreamName    = "gotube:jobs"
	ConsumerGroup = "gotube-workers"
)

type JobMessage struct {
	JobID   uuid.UUID `json:"job_id"`
	VideoID uuid.UUID `json:"video_id"`
	JobType string    `json:"job_type"`
}

type Queue struct {
	rdb *redis.Client
}

func NewQueue(rdb *redis.Client) *Queue {
	return &Queue{rdb: rdb}
}

func (q *Queue) CreateConsumerGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(ctx, StreamName, ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("create consumer group: %w", err)
	}
	return nil
}

func (q *Queue) Enqueue(ctx context.Context, msg JobMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}
	return q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{
			"payload": string(data),
		},
	}).Err()
}

func (q *Queue) Consume(ctx context.Context, consumerName string, count int64, blockDuration time.Duration) ([]redis.XMessage, error) {
	streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumerName,
		Streams:  []string{StreamName, ">"},
		Count:    count,
		Block:    blockDuration,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

func (q *Queue) Ack(ctx context.Context, messageID string) error {
	return q.rdb.XAck(ctx, StreamName, ConsumerGroup, messageID).Err()
}

func (q *Queue) ParseMessage(msg redis.XMessage) (*JobMessage, error) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid message payload")
	}
	var job JobMessage
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job message: %w", err)
	}
	return &job, nil
}

func (q *Queue) Len(ctx context.Context) (int64, error) {
	return q.rdb.XLen(ctx, StreamName).Result()
}

// ReclaimStale re-processes messages that were claimed but never acknowledged.
func (q *Queue) ReclaimStale(ctx context.Context, consumerName string, minIdle time.Duration, count int64) ([]redis.XMessage, error) {
	messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   StreamName,
		Group:    ConsumerGroup,
		Consumer: consumerName,
		MinIdle:  minIdle,
		Start:    "0-0",
		Count:    count,
	}).Result()
	if err != nil {
		return nil, err
	}
	log.Debug().Int("reclaimed", len(messages)).Msg("reclaimed stale messages")
	return messages, nil
}
