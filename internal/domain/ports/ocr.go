package ports

import "context"

type RawOCRResult struct {
	RawText    string
	Lines      []string
	Confidence *float64
}

type OCRProvider interface {
	Extract(ctx context.Context, imageRef string) (RawOCRResult, error)
}
