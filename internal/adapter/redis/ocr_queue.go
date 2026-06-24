package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type OCRQueue struct {
	client *redis.Client
	key    string
}

func NewOCRQueue(client *redis.Client, key string) *OCRQueue {
	if key == "" {
		key = "ocr:jobs"
	}
	return &OCRQueue{client: client, key: key}
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
