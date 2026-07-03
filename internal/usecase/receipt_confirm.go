package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ReceiptConfirmUseCase struct {
	repo      ports.ReceiptRepository
	events    ports.ReceiptEvents
	recompute *PriceAggregateRecomputeUseCase
	loyalty   *LoyaltyAwardUseCase
	firstObs  ports.FirstObservationChecker
}

func NewReceiptConfirmUseCase(
	repo ports.ReceiptRepository,
	events ports.ReceiptEvents,
	recompute *PriceAggregateRecomputeUseCase,
	loyalty *LoyaltyAwardUseCase,
	firstObs ports.FirstObservationChecker,
) *ReceiptConfirmUseCase {
	return &ReceiptConfirmUseCase{
		repo:      repo,
		events:    events,
		recompute: recompute,
		loyalty:   loyalty,
		firstObs:  firstObs,
	}
}

// ConfirmResult is what Execute returns: the points the loyalty layer
// actually awarded and the human-readable reasons. Echoed back to the
// mobile client so the success screen can show "+18 pts (first product +
// complete data)" instead of a blank "+0".
type ConfirmResult struct {
	PointsEarned int
	Reasons      []string
}

func (u *ReceiptConfirmUseCase) Execute(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload) (ConfirmResult, error) {
	if userID == "" || receiptID == "" {
		return ConfirmResult{}, fmt.Errorf("user id and receipt id are required")
	}
	if _, err := time.Parse("2006-01-02", payload.PurchaseDate); err != nil {
		return ConfirmResult{}, fmt.Errorf("purchase_date must use format YYYY-MM-DD")
	}
	if len(payload.Items) == 0 {
		return ConfirmResult{}, fmt.Errorf("at least one item is required")
	}

	// Resolve the store BEFORE ConfirmReceipt commits so we can signal the
	// loyalty award whether this confirmation created a brand-new store.
	// ConfirmReceipt internally re-resolves the store within its tx but
	// finds the row already inserted (idempotent), so no duplicate store
	// row is created.
	storeID, storeCreatedNow, err := u.repo.ResolveOrCreateStore(ctx, payload.Store)
	if err != nil {
		return ConfirmResult{}, err
	}

	now := time.Now().UTC()
	observations := make([]entities.PriceObservation, 0, len(payload.Items))

	// Pre-compute first-observation (per distinct product-store pair) BEFORE
	// ConfirmReceipt inserts the new price_observations, eliminating the
	// race described in spec R-04. The set is built in the same order as
	// observations so the dedup-by-set semantics is the same regardless of
	// input ordering.
	firstObsPairs := make(map[string]struct{})

	for _, item := range payload.Items {
		if strings.TrimSpace(item.RawText) == "" {
			return ConfirmResult{}, fmt.Errorf("item raw_text is required")
		}
		if item.UnitPrice == nil {
			return ConfirmResult{}, fmt.Errorf("item unit_price is required")
		}
		if item.Currency == nil || strings.TrimSpace(*item.Currency) == "" {
			return ConfirmResult{}, fmt.Errorf("item currency is required")
		}
		productID, _, err := u.repo.NormalizeProduct(ctx, item.RawText)
		if err != nil {
			return ConfirmResult{}, err
		}
		observations = append(observations, entities.PriceObservation{
			ProductID:  productID,
			StoreID:    storeID,
			UnitPrice:  *item.UnitPrice,
			Currency:   strings.TrimSpace(*item.Currency),
			ObservedAt: now,
			ReceiptID:  receiptID,
		})

		// Detect first-observation ONLY if needed (loyalty not nil). The
		// check is keyed on (productID, storeID) and deduplicated across
		// the items in this confirmation; the award use case grants the
		// bonus once per distinct first pair.
		if u.loyalty != nil && u.firstObs != nil {
			pairKey := productID + "|" + storeID
			if _, dup := firstObsPairs[pairKey]; dup {
				continue
			}
			firstObsPairs[pairKey] = struct{}{}
		}
	}

	// Resolve all first-observation checks against the pre-persistence
	// state (no row for this pair in price_observations yet).
	firstObservationPairCount := 0
	if u.loyalty != nil && u.firstObs != nil {
		for pairKey := range firstObsPairs {
			parts := strings.SplitN(pairKey, "|", 2)
			if len(parts) != 2 {
				continue
			}
			known, err := u.firstObs.PreviouslyObserved(ctx, parts[0], parts[1])
			if err != nil {
				// Non-fatal: skip the bonus rather than failing the confirm.
				continue
			}
			if !known {
				firstObservationPairCount++
			}
		}
	}

	if err := u.repo.ConfirmReceipt(ctx, receiptID, userID, payload, observations); err != nil {
		return ConfirmResult{}, err
	}

	if u.recompute != nil {
		if err := u.recompute.Execute(ctx, observations, 90); err != nil {
			return ConfirmResult{}, fmt.Errorf("recompute aggregates: %w", err)
		}
	}

	var pointsEarned int
	var reasons []string
	if u.loyalty != nil {
		// Fetch the freshly-confirmed receipt so the award sees the
		// CONFIRMED status that the persistence layer just set.
		if receipt, err := u.repo.GetByID(ctx, receiptID); err == nil && receipt != nil {
			awardInput := AwardInput{
				Receipt:                   *receipt,
				Payload:                   payload,
				StoreCreatedNow:           storeCreatedNow,
				FirstObservationPairCount: firstObservationPairCount,
			}
			pointsEarned, reasons = u.loyalty.AwardForReceipt(ctx, awardInput)
		}
	}

	if u.events != nil {
		if err := u.events.EmitReceiptConfirmed(ctx, receiptID, userID, len(observations)); err != nil {
			return ConfirmResult{PointsEarned: pointsEarned, Reasons: reasons}, err
		}
	}

	return ConfirmResult{PointsEarned: pointsEarned, Reasons: reasons}, nil
}
