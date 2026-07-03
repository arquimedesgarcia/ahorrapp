package ports

import (
	"context"

	"ahorrapp/internal/domain/entities"
)

type RankingQueryOptions struct {
	Lat      *float64
	Long     *float64
	RadiusKm *float64
}

func (o RankingQueryOptions) HasLocation() bool {
	return o.Lat != nil && o.Long != nil
}

func (o RankingQueryOptions) HasRadius() bool {
	return o.RadiusKm != nil
}

type RankingRepository interface {
	RecomputeAggregate(ctx context.Context, productID, storeID, currency string, ageThresholdDays int) error
	RecomputeAll(ctx context.Context, ageThresholdDays int) error
	GetProductRanking(ctx context.Context, productID string, opts RankingQueryOptions) ([]entities.PriceAggregate, error)
	SearchProducts(ctx context.Context, query string) ([]entities.ProductSearchResult, error)
	GetProductName(ctx context.Context, productID string) (string, error)
}
