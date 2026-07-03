package usecase

import (
	"context"
	"fmt"
	"strings"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type RankingUseCase struct {
	repo ports.RankingRepository
}

func NewRankingUseCase(repo ports.RankingRepository) *RankingUseCase {
	return &RankingUseCase{repo: repo}
}

type ProductRankingResponse struct {
	ProductID        string                               `json:"product_id"`
	ProductName      string                               `json:"product_name"`
	CurrencyRankings map[string][]entities.PriceAggregate `json:"currency_rankings"`
}

type SearchResponse struct {
	Results []entities.ProductSearchResult `json:"results"`
}

func (u *RankingUseCase) GetProductRanking(ctx context.Context, productID string, opts ports.RankingQueryOptions) (*ProductRankingResponse, error) {
	entries, err := u.repo.GetProductRanking(ctx, productID, opts)
	if err != nil {
		return nil, fmt.Errorf("get product ranking: %w", err)
	}

	name, err := u.repo.GetProductName(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get product name: %w", err)
	}

	rankings := make(map[string][]entities.PriceAggregate)
	for _, entry := range entries {
		rankings[entry.Currency] = append(rankings[entry.Currency], entry)
	}

	return &ProductRankingResponse{
		ProductID:        productID,
		ProductName:      name,
		CurrencyRankings: rankings,
	}, nil
}

func (u *RankingUseCase) SearchProducts(ctx context.Context, query string) (*SearchResponse, error) {
	normalized := strings.TrimSpace(query)
	if len(normalized) < 3 {
		return nil, fmt.Errorf("search query must be at least 3 characters")
	}

	results, err := u.repo.SearchProducts(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}

	return &SearchResponse{Results: results}, nil
}

func (u *RankingUseCase) SearchProductsLegacy(ctx context.Context, query string) (*LegacySearchResponse, error) {
	normalized := strings.TrimSpace(query)
	if len(normalized) < 1 {
		return nil, fmt.Errorf("query parameter 'q' is required")
	}

	results, err := u.repo.SearchProducts(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}

	legacyResults := make([]LegacyProductResult, 0, len(results))
	for _, r := range results {
		stores := make([]LegacyStoreEntry, 0)
		for _, best := range r.BestPrices {
			if best == nil {
				continue
			}
			stores = append(stores, LegacyStoreEntry{
				StoreID:      best.StoreID,
				StoreName:    best.StoreName,
				Branch:       best.Branch,
				AveragePrice: best.AveragePrice,
				Currency:     best.Currency,
				SampleCount:  best.SampleCount,
			})
		}
		legacyResults = append(legacyResults, LegacyProductResult{
			ProductID:   r.ProductID,
			ProductName: r.ProductName,
			Unit:        r.Unit,
			Stores:      stores,
		})
	}

	return &LegacySearchResponse{Results: legacyResults}, nil
}

type LegacySearchResponse struct {
	Results []LegacyProductResult `json:"results"`
}

type LegacyProductResult struct {
	ProductID   string             `json:"product_id"`
	ProductName string             `json:"product_name"`
	Unit        *string            `json:"unit"`
	Stores      []LegacyStoreEntry `json:"stores"`
}

type LegacyStoreEntry struct {
	StoreID      string  `json:"store_id"`
	StoreName    string  `json:"store_name"`
	Branch       *string `json:"branch,omitempty"`
	AveragePrice float64 `json:"average_price"`
	Currency     string  `json:"currency"`
	SampleCount  int     `json:"sample_count"`
}
