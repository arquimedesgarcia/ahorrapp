package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

// validCurrencies is the closed set enforced by the price_aggregates
// and receipt_items CHECK constraints (see migrations/000004 and the
// 000003 schema). Validating here gives a clear 400 error instead of
// letting the DB reject with an opaque constraint violation.
var validCurrencies = map[string]struct{}{
	"USD": {},
	"Bs.": {},
}

func isValidCurrency(c string) bool {
	_, ok := validCurrencies[c]
	return ok
}

type ReceiptConfirmUseCase struct {
	repo                  ports.ReceiptRepository
	events                ports.ReceiptEvents
	recompute             *PriceAggregateRecomputeUseCase
	loyalty               *LoyaltyAwardUseCase
	firstObs              ports.FirstObservationChecker
	priceAgeThresholdDays int
}

func NewReceiptConfirmUseCase(
	repo ports.ReceiptRepository,
	events ports.ReceiptEvents,
	recompute *PriceAggregateRecomputeUseCase,
	loyalty *LoyaltyAwardUseCase,
	firstObs ports.FirstObservationChecker,
	priceAgeThresholdDays int,
) *ReceiptConfirmUseCase {
	if priceAgeThresholdDays < 1 {
		priceAgeThresholdDays = 90
	}
	return &ReceiptConfirmUseCase{
		repo:                  repo,
		events:                events,
		recompute:             recompute,
		loyalty:               loyalty,
		firstObs:              firstObs,
		priceAgeThresholdDays: priceAgeThresholdDays,
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
		normalizedCurrency := strings.TrimSpace(*item.Currency)
		if !isValidCurrency(normalizedCurrency) {
			return ConfirmResult{}, fmt.Errorf("item currency %q is not supported (expected USD or Bs.)", normalizedCurrency)
		}
		productID, _, err := u.repo.NormalizeProduct(ctx, item.RawText)
		if err != nil {
			return ConfirmResult{}, err
		}
		observations = append(observations, entities.PriceObservation{
			ProductID:  productID,
			StoreID:    storeID,
			UnitPrice:  *item.UnitPrice,
			Currency:   normalizedCurrency,
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
		// Recompute failure MUST NOT fail the confirm flow: the receipt
		// is already persisted as CONFIRMED, the price_observations are
		// already stored, and the next confirm will fix the aggregates.
		// Returning an error here would leave the mobile app retrying
		// against a row that is already CONFIRMED, producing duplicate
		// observations. Log and continue.
		if err := u.recompute.Execute(ctx, observations, u.priceAgeThresholdDays); err != nil {
			log.Printf("receipt confirm: recompute aggregates for %s failed: %v (price_aggregates will be stale until next confirm)", receiptID, err)
		}
	}

	var pointsEarned int
	var reasons []string
	if u.loyalty != nil {
		// Build the AwardInput directly from what we already know in
		// memory: the receipt was just transitioned to CONFIRMED inside
		// the ConfirmReceipt tx, so the status is known and the
		// user_id is the same one passed into this use case. This
		// eliminates the post-commit GetByID round trip (and the small
		// window in which it could return a row whose status was
		// reverted by a concurrent operation).
		awardInput := AwardInput{
			Receipt: entities.Receipt{
				ID:     receiptID,
				UserID: userID,
				Status: entities.ReceiptStatusConfirmed,
			},
			Payload:                   payload,
			StoreCreatedNow:           storeCreatedNow,
			FirstObservationPairCount: firstObservationPairCount,
		}
		pointsEarned, reasons = u.loyalty.AwardForReceipt(ctx, awardInput)
	}

	if u.events != nil {
		if err := u.events.EmitReceiptConfirmed(ctx, receiptID, userID, len(observations)); err != nil {
			return ConfirmResult{PointsEarned: pointsEarned, Reasons: reasons}, err
		}
	}

	return ConfirmResult{PointsEarned: pointsEarned, Reasons: reasons}, nil
}
