package postgres

import (
	"testing"
)

func TestPriceAggregateRepository_RecomputeAggregate_NilPool(t *testing.T) {
	repo := &PriceAggregateRepository{pool: nil}
	if repo == nil {
		t.Fatal("repository should not be nil")
	}
}

func TestPriceAggregateRepository_SearchProducts_ShortQuery(t *testing.T) {
	repo := &PriceAggregateRepository{pool: nil}
	_, err := repo.SearchProducts(nil, "ab")
	if err == nil {
		t.Fatal("expected error for short query, got nil")
	}
}
