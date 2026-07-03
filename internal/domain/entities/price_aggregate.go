package entities

import "time"

type PriceAggregate struct {
	ProductID      string    `json:"-"`
	StoreID        string    `json:"store_id"`
	StoreName      string    `json:"store_name"`
	Branch         *string   `json:"branch,omitempty"`
	Currency       string    `json:"currency"`
	AveragePrice   float64   `json:"average_price"`
	MinPrice       float64   `json:"min_price"`
	SampleCount    int       `json:"sample_count"`
	LastObservedAt time.Time `json:"last_observed_at"`
	UpdatedAt      time.Time `json:"-"`
	DistanceKm     *float64  `json:"distance_km,omitempty"`
}

type ProductSearchResult struct {
	ProductID   string                     `json:"product_id"`
	ProductName string                     `json:"product_name"`
	Unit        *string                    `json:"unit"`
	BestPrices  map[string]*PriceAggregate `json:"best_prices"`
}
