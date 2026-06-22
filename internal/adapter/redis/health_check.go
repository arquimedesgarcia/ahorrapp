package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Checker struct {
	client *redis.Client
}

func NewChecker(client *redis.Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) Name() string {
	return "redis"
}

func (c *Checker) Check(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
