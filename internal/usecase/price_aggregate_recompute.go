package usecase

import (
	"context"
	"fmt"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type PriceAggregateRecomputeUseCase struct {
	repo ports.RankingRepository
}

func NewPriceAggregateRecomputeUseCase(repo ports.RankingRepository) *PriceAggregateRecomputeUseCase {
	return &PriceAggregateRecomputeUseCase{repo: repo}
}

func (u *PriceAggregateRecomputeUseCase) Execute(ctx context.Context, observations []entities.PriceObservation, ageThresholdDays int) error {
	if len(observations) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	for _, obs := range observations {
		key := fmt.Sprintf("%s|%s|%s", obs.ProductID, obs.StoreID, obs.Currency)
		if seen[key] {
			continue
		}
		seen[key] = true
		if err := u.repo.RecomputeAggregate(ctx, obs.ProductID, obs.StoreID, obs.Currency, ageThresholdDays); err != nil {
			return fmt.Errorf("recompute aggregate for %s: %w", key, err)
		}
	}
	return nil
}
