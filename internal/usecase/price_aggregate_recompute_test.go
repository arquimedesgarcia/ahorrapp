package usecase

import (
	"context"
	"testing"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type stubRankingRepo struct {
	calls       []recomputeCall
	ranking     []entities.PriceAggregate
	search      []entities.ProductSearchResult
	productName string
}

type recomputeCall struct {
	productID        string
	storeID          string
	currency         string
	ageThresholdDays int
}

func (s *stubRankingRepo) RecomputeAggregate(ctx context.Context, productID, storeID, currency string, ageThresholdDays int) error {
	s.calls = append(s.calls, recomputeCall{productID, storeID, currency, ageThresholdDays})
	return nil
}

func (s *stubRankingRepo) RecomputeAll(ctx context.Context, ageThresholdDays int) error {
	return nil
}

func (s *stubRankingRepo) GetProductRanking(ctx context.Context, productID string, opts ports.RankingQueryOptions) ([]entities.PriceAggregate, error) {
	return s.ranking, nil
}

func (s *stubRankingRepo) SearchProducts(ctx context.Context, query string) ([]entities.ProductSearchResult, error) {
	return s.search, nil
}

func (s *stubRankingRepo) GetProductName(ctx context.Context, productID string) (string, error) {
	return s.productName, nil
}

func TestPriceAggregateRecompute_DeduplicatesTriples(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	observations := []entities.PriceObservation{
		{ProductID: "p1", StoreID: "s1", Currency: "USD", UnitPrice: 1.00, ObservedAt: time.Now()},
		{ProductID: "p1", StoreID: "s1", Currency: "USD", UnitPrice: 1.20, ObservedAt: time.Now()},
		{ProductID: "p1", StoreID: "s1", Currency: "USD", UnitPrice: 1.40, ObservedAt: time.Now()},
	}

	if err := uc.Execute(context.Background(), observations, 90); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(repo.calls) != 1 {
		t.Errorf("expected 1 RecomputeAggregate call, got %d", len(repo.calls))
	}
}

func TestPriceAggregateRecompute_CurrencyIsolation(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	observations := []entities.PriceObservation{
		{ProductID: "p1", StoreID: "s1", Currency: "USD", UnitPrice: 1.00, ObservedAt: time.Now()},
		{ProductID: "p1", StoreID: "s1", Currency: "Bs.", UnitPrice: 100.00, ObservedAt: time.Now()},
	}

	if err := uc.Execute(context.Background(), observations, 90); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(repo.calls) != 2 {
		t.Errorf("expected 2 RecomputeAggregate calls (one per currency), got %d", len(repo.calls))
	}

	currencies := map[string]bool{}
	for _, call := range repo.calls {
		currencies[call.currency] = true
	}
	if !currencies["USD"] || !currencies["Bs."] {
		t.Errorf("expected both USD and Bs. calls, got: %+v", currencies)
	}
}

func TestPriceAggregateRecompute_NewProduct(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	observations := []entities.PriceObservation{
		{ProductID: "new-product", StoreID: "s1", Currency: "USD", UnitPrice: 2.50, ObservedAt: time.Now()},
	}

	if err := uc.Execute(context.Background(), observations, 90); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(repo.calls) != 1 {
		t.Errorf("expected 1 call for new product, got %d", len(repo.calls))
	}
}

func TestPriceAggregateRecompute_EmptyObservations(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	if err := uc.Execute(context.Background(), nil, 90); err != nil {
		t.Fatalf("Execute with empty list: %v", err)
	}

	if len(repo.calls) != 0 {
		t.Errorf("expected 0 calls for empty observations, got %d", len(repo.calls))
	}
}

func TestPriceAggregateRecompute_MultipleTriples(t *testing.T) {
	repo := &stubRankingRepo{}
	uc := NewPriceAggregateRecomputeUseCase(repo)

	observations := []entities.PriceObservation{
		{ProductID: "p1", StoreID: "s1", Currency: "USD", ObservedAt: time.Now()},
		{ProductID: "p1", StoreID: "s2", Currency: "USD", ObservedAt: time.Now()},
		{ProductID: "p2", StoreID: "s1", Currency: "Bs.", ObservedAt: time.Now()},
	}

	if err := uc.Execute(context.Background(), observations, 90); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(repo.calls) != 3 {
		t.Errorf("expected 3 unique triple calls, got %d", len(repo.calls))
	}
}
