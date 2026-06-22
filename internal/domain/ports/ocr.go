package ports

import "context"

type OCRProvider interface {
	Extract(ctx context.Context, imageRef string) (string, error)
}
