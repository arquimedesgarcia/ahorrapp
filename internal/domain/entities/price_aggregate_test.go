package entities

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPriceAggregateJSONTags(t *testing.T) {
	branch := "Downtown"
	agg := PriceAggregate{
		ProductID:      "prod-1",
		StoreID:        "store-1",
		StoreName:      "Central Market",
		Branch:         &branch,
		Currency:       "USD",
		AveragePrice:   1.25,
		MinPrice:       1.10,
		SampleCount:    45,
		LastObservedAt: time.Date(2025, 6, 28, 14, 30, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2025, 6, 28, 14, 31, 0, 0, time.UTC),
	}

	data, err := json.Marshal(agg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)

	if !strings.Contains(jsonStr, `"store_id":"store-1"`) {
		t.Errorf("expected store_id in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"store_name":"Central Market"`) {
		t.Errorf("expected store_name in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"average_price":1.25`) {
		t.Errorf("expected average_price in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"sample_count":45`) {
		t.Errorf("expected sample_count in JSON, got: %s", jsonStr)
	}
}

func TestPriceAggregateNilBranch(t *testing.T) {
	agg := PriceAggregate{
		StoreID:  "s1",
		Branch:   nil,
		Currency: "Bs.",
	}

	data, err := json.Marshal(agg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), `"branch"`) {
		t.Errorf("expected branch to be omitted when nil, got: %s", string(data))
	}
}

func TestProductSearchResultJSONTags(t *testing.T) {
	result := ProductSearchResult{
		ProductID:   "p1",
		ProductName: "Arroz Blanco",
		Unit:        nil,
		BestPrices: map[string]*PriceAggregate{
			"USD": {
				StoreID:      "s1",
				StoreName:    "Central Market",
				Currency:     "USD",
				AveragePrice: 1.25,
				SampleCount:  10,
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)

	if !strings.Contains(jsonStr, `"product_id":"p1"`) {
		t.Errorf("expected product_id in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"best_prices"`) {
		t.Errorf("expected best_prices in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"USD"`) {
		t.Errorf("expected USD key in best_prices, got: %s", jsonStr)
	}
}
