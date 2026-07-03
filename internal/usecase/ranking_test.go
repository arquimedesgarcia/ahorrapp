package usecase

import (
	"context"
	"testing"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

func TestRankingUseCase_GetProductRanking_OrdersByAveragePrice(t *testing.T) {
	repo := &stubRankingRepo{
		ranking: []entities.PriceAggregate{
			{StoreID: "s1", StoreName: "Store A", Currency: "USD", AveragePrice: 1.00, SampleCount: 5},
			{StoreID: "s2", StoreName: "Store B", Currency: "USD", AveragePrice: 1.50, SampleCount: 3},
			{StoreID: "s3", StoreName: "Store C", Currency: "USD", AveragePrice: 1.20, SampleCount: 4},
		},
		productName: "Test Product",
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.GetProductRanking(context.Background(), "p1", ports.RankingQueryOptions{})
	if err != nil {
		t.Fatalf("GetProductRanking: %v", err)
	}

	if resp.ProductName != "Test Product" {
		t.Errorf("expected product name 'Test Product', got '%s'", resp.ProductName)
	}

	usaRanking := resp.CurrencyRankings["USD"]
	if len(usaRanking) != 3 {
		t.Fatalf("expected 3 stores in USD ranking, got %d", len(usaRanking))
	}

	if usaRanking[0].StoreName != "Store A" {
		t.Errorf("expected cheapest store first (Store A at 1.00), got %s", usaRanking[0].StoreName)
	}
}

func TestRankingUseCase_GetProductRanking_CurrencyIsolation(t *testing.T) {
	repo := &stubRankingRepo{
		ranking: []entities.PriceAggregate{
			{StoreID: "s1", StoreName: "Store A", Currency: "USD", AveragePrice: 1.00},
			{StoreID: "s2", StoreName: "Store B", Currency: "Bs.", AveragePrice: 100.00},
		},
		productName: "Test Product",
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.GetProductRanking(context.Background(), "p1", ports.RankingQueryOptions{})
	if err != nil {
		t.Fatalf("GetProductRanking: %v", err)
	}

	if len(resp.CurrencyRankings["USD"]) != 1 {
		t.Errorf("expected 1 store in USD group")
	}
	if len(resp.CurrencyRankings["Bs."]) != 1 {
		t.Errorf("expected 1 store in Bs. group")
	}

	for currency := range resp.CurrencyRankings {
		for _, entry := range resp.CurrencyRankings[currency] {
			if entry.Currency != currency {
				t.Errorf("currency mismatch: group %s contains entry with currency %s", currency, entry.Currency)
			}
		}
	}
}

func TestRankingUseCase_GetProductRanking_EmptyRanking(t *testing.T) {
	repo := &stubRankingRepo{
		ranking:     []entities.PriceAggregate{},
		productName: "Empty Product",
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.GetProductRanking(context.Background(), "p1", ports.RankingQueryOptions{})
	if err != nil {
		t.Fatalf("GetProductRanking: %v", err)
	}

	if len(resp.CurrencyRankings) != 0 {
		t.Errorf("expected empty currency_rankings, got %d currencies", len(resp.CurrencyRankings))
	}
}

func TestRankingUseCase_SearchProducts_ReturnsResults(t *testing.T) {
	repo := &stubRankingRepo{
		search: []entities.ProductSearchResult{
			{ProductID: "p1", ProductName: "Arroz Blanco", BestPrices: map[string]*entities.PriceAggregate{
				"USD": {StoreName: "Store A", Currency: "USD", AveragePrice: 1.25},
			}},
		},
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.SearchProducts(context.Background(), "arroz")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].ProductName != "Arroz Blanco" {
		t.Errorf("expected 'Arroz Blanco', got '%s'", resp.Results[0].ProductName)
	}
}

func TestRankingUseCase_SearchProducts_EmptyResults(t *testing.T) {
	repo := &stubRankingRepo{
		search: []entities.ProductSearchResult{},
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.SearchProducts(context.Background(), "xyzabc")
	if err != nil {
		t.Fatalf("SearchProducts: %v", err)
	}

	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

func TestRankingUseCase_SearchProducts_ShortQuery(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewRankingUseCase(repo)

	_, err := uc.SearchProducts(context.Background(), "ab")
	if err == nil {
		t.Fatal("expected error for query < 3 chars, got nil")
	}
}

func TestRankingUseCase_SearchProducts_EmptyQuery(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewRankingUseCase(repo)

	_, err := uc.SearchProducts(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

func TestRankingUseCase_GetProductRanking_ProximityOptions(t *testing.T) {
	lat := 10.5
	long := -66.9
	repo := &stubRankingRepo{
		ranking: []entities.PriceAggregate{
			{StoreID: "s1", StoreName: "Near Store", Currency: "USD", AveragePrice: 1.00, DistanceKm: &[]float64{0.5}[0]},
			{StoreID: "s2", StoreName: "Far Store", Currency: "USD", AveragePrice: 1.50, DistanceKm: &[]float64{50.0}[0]},
		},
		productName: "Test Product",
	}
	uc := NewRankingUseCase(repo)

	opts := ports.RankingQueryOptions{Lat: &lat, Long: &long}
	resp, err := uc.GetProductRanking(context.Background(), "p1", opts)
	if err != nil {
		t.Fatalf("GetProductRanking with proximity: %v", err)
	}

	if len(resp.CurrencyRankings["USD"]) != 2 {
		t.Fatalf("expected 2 stores, got %d", len(resp.CurrencyRankings["USD"]))
	}
}

func TestRankingUseCase_SearchProductsLegacy(t *testing.T) {
	repo := &stubRankingRepo{
		search: []entities.ProductSearchResult{
			{ProductID: "p1", ProductName: "Arroz", BestPrices: map[string]*entities.PriceAggregate{
				"USD": {StoreID: "s1", StoreName: "Central Market", Currency: "USD", AveragePrice: 1.25, SampleCount: 45},
			}},
		},
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.SearchProductsLegacy(context.Background(), "arroz")
	if err != nil {
		t.Fatalf("SearchProductsLegacy: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	r := resp.Results[0]
	if r.ProductName != "Arroz" {
		t.Errorf("expected product name 'Arroz', got '%s'", r.ProductName)
	}
	if len(r.Stores) != 1 {
		t.Fatalf("expected 1 store, got %d", len(r.Stores))
	}
	if r.Stores[0].StoreName != "Central Market" {
		t.Errorf("expected store name 'Central Market', got '%s'", r.Stores[0].StoreName)
	}
}

func TestRankingUseCase_SearchProductsLegacy_EmptyQuery(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewRankingUseCase(repo)

	_, err := uc.SearchProductsLegacy(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

func TestRankingUseCase_StaleStoreExcluded(t *testing.T) {
	repo := &stubRankingRepo{
		ranking: []entities.PriceAggregate{
			{StoreID: "s1", StoreName: "Fresh Store", Currency: "USD", AveragePrice: 1.00, SampleCount: 5},
		},
		productName: "Test Product",
	}
	uc := NewRankingUseCase(repo)

	resp, err := uc.GetProductRanking(context.Background(), "p1", ports.RankingQueryOptions{})
	if err != nil {
		t.Fatalf("GetProductRanking: %v", err)
	}

	for _, entries := range resp.CurrencyRankings {
		for _, e := range entries {
			if e.SampleCount == 0 {
				t.Errorf("store with sample_count=0 should not appear in ranking: %s", e.StoreName)
			}
		}
	}
}

func TestPriceAggregateRecompute_AgeThresholdPassed(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	observations := []entities.PriceObservation{
		{ProductID: "p1", StoreID: "s1", Currency: "USD", ObservedAt: time.Now()},
	}

	if err := uc.Execute(context.Background(), observations, 180); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(repo.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(repo.calls))
	}
	if repo.calls[0].ageThresholdDays != 180 {
		t.Errorf("expected ageThreshold 180, got %d", repo.calls[0].ageThresholdDays)
	}
}
