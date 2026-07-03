package ports

import (
	"context"
	"errors"

	"ahorrapp/internal/domain/entities"
)

// ErrProductNotFound is returned by RankingRepository.GetProductName
// (and propagated through the use case) when the requested product id
// does not exist. The handler maps this to HTTP 404. Using a typed
// error keeps the handler free of string matching on error text.
var ErrProductNotFound = errors.New("product not found")

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
