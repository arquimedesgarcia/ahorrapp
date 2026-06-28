package usecase

import (
	"context"

	"ahorrapp/internal/domain/ports"
)

type ProductNormalizer struct {
	repo ports.ReceiptRepository
}

func NewProductNormalizer(repo ports.ReceiptRepository) *ProductNormalizer {
	return &ProductNormalizer{repo: repo}
}

func (n *ProductNormalizer) Normalize(ctx context.Context, rawName string) (string, string, error) {
	return n.repo.NormalizeProduct(ctx, rawName)
}
