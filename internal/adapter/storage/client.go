package storage

import (
	"context"
	"fmt"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Upload(ctx context.Context, objectName string, data []byte) (string, error) {
	_ = ctx
	_ = data
	return "", fmt.Errorf("storage adapter not implemented for skeleton: %s", objectName)
}
