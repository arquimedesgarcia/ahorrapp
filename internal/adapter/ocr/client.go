package ocr

import (
	"context"
	"fmt"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Extract(ctx context.Context, imageRef string) (string, error) {
	_ = ctx
	return "", fmt.Errorf("ocr adapter not implemented for skeleton: %s", imageRef)
}
