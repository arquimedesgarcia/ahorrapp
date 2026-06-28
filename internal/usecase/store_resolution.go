package usecase

import (
	"context"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type StoreResolver struct {
	repo ports.ReceiptRepository
}

func NewStoreResolver(repo ports.ReceiptRepository) *StoreResolver {
	return &StoreResolver{repo: repo}
}

func (r *StoreResolver) Resolve(ctx context.Context, store entities.StoreSummary) (string, error) {
	return r.repo.ResolveOrCreateStore(ctx, store)
}
