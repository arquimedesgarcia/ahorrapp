package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OCRQueue struct {
	client *redis.Client
	key    string
	dlqKey string
}

func NewOCRQueue(client *redis.Client, key string) *OCRQueue {
	if key == "" {
		key = "ocr:jobs"
	}
	return &OCRQueue{client: client, key: key, dlqKey: key + ":dlq"}
}

func (q *OCRQueue) Enqueue(ctx context.Context, receiptID string) error {
	return q.client.LPush(ctx, q.key, receiptID).Err()
}

func (q *OCRQueue) Dequeue(ctx context.Context) (string, error) {
	out, err := q.client.BRPop(ctx, 5e9, q.key).Result()
	if err != nil {
		return "", err
	}
	if len(out) < 2 {
		return "", fmt.Errorf("invalid queue payload")
	}
	return out[1], nil
}

func (q *OCRQueue) RequeueWithBackoff(ctx context.Context, receiptID string, attempt int, reason string) error {
	if attempt < 1 {
		attempt = 1
	}
	backoff := time.Duration(attempt*attempt) * time.Second
	t := time.NewTimer(backoff)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
	}

	if err := q.client.LPush(ctx, q.key, receiptID).Err(); err != nil {
		return err
	}
	if reason != "" {
		_ = q.client.HSet(ctx, q.key+":retry-reasons", receiptID, reason).Err()
	}
	return nil
}

func (q *OCRQueue) SendToDeadLetter(ctx context.Context, receiptID, reason string) error {
	if err := q.client.LPush(ctx, q.dlqKey, receiptID).Err(); err != nil {
		return err
	}
	if reason != "" {
		_ = q.client.HSet(ctx, q.dlqKey+":reasons", receiptID, reason).Err()
	}
	return nil
}
