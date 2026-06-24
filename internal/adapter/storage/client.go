package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func NewClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   minioClient,
		bucket:   bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

func (c *Client) Upload(ctx context.Context, objectName string, data []byte) (string, error) {
	if strings.TrimSpace(objectName) == "" {
		return "", fmt.Errorf("object name is required")
	}
	if len(data) == 0 {
		return "", fmt.Errorf("image payload is empty")
	}

	err := c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, bucketErr := c.client.BucketExists(ctx, c.bucket)
		if bucketErr != nil {
			return "", bucketErr
		}
		if !exists {
			return "", err
		}
	}

	_, err = c.client.PutObject(ctx, c.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return "", err
	}

	scheme := "http"
	if c.useSSL {
		scheme = "https"
	}
	u := url.URL{Scheme: scheme, Host: c.endpoint, Path: fmt.Sprintf("/%s/%s", c.bucket, objectName)}
	return u.String(), nil
}
