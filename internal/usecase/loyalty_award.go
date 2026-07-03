package usecase

import (
	"context"
	"log"
	"strings"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

// AwardInput is the pre-computed bundle passed to LoyaltyAwardUseCase.
// The confirm use case computes FirstObservationPairCount and
// StoreCreatedNow BEFORE the receipt's price_observations are persisted
// (so the check sees the world before this confirmation, eliminating the
// race described in spec R-04). The award use case itself is intentionally
// free of any DB read other than the daily-cap count + the award insert.
type AwardInput struct {
	Receipt                   entities.Receipt
	Payload                   entities.ConfirmPayload
	StoreCreatedNow           bool
	FirstObservationPairCount int
}

type LoyaltyAwardUseCase struct {
	repo                  ports.LoyaltyRepository
	basePoints            int
	firstObservationBonus int
	dataCompletionBonus   int
	dailyAwardCap         int
}

func NewLoyaltyAwardUseCase(
	repo ports.LoyaltyRepository,
	basePoints, firstObsBonus, dataCompletionBonus, dailyCap int,
) *LoyaltyAwardUseCase {
	return &LoyaltyAwardUseCase{
		repo:                  repo,
		basePoints:            basePoints,
		firstObservationBonus: firstObsBonus,
		dataCompletionBonus:   dataCompletionBonus,
		dailyAwardCap:         dailyCap,
	}
}

// AwardForReceipt awards base + bonuses for one confirmed receipt. It
// is the single source of the rule: daily cap > base > bonuses > reasons.
// Awarding never aborts the confirm flow; errors are logged and swallowed
// (per plan R-07). Returns the points actually awarded and the reason
// codes so the confirm handler can echo them back to the mobile app.
func (u *LoyaltyAwardUseCase) AwardForReceipt(ctx context.Context, in AwardInput) (int, []string) {
	if in.Receipt.Status != entities.ReceiptStatusConfirmed {
		return 0, nil
	}

	count, err := u.repo.DailyGrantCount(ctx, in.Receipt.UserID)
	if err != nil {
		log.Printf("loyalty: daily grant count for user %s failed: %v (awarding base anyway)", in.Receipt.UserID, err)
		count = 0
	}
	capped := u.dailyAwardCap > 0 && count >= u.dailyAwardCap
	if capped {
		_ = u.persist(ctx, in.Receipt, 0, []string{ReasonDailyLimitReached})
		return 0, []string{ReasonDailyLimitReached}
	}

	reasons := []string{ReasonReceiptConfirmed}
	points := u.basePoints

	if in.FirstObservationPairCount > 0 {
		reasons = append(reasons, ReasonFirstObservationProduct)
		points += u.firstObservationBonus
	}
	if in.StoreCreatedNow {
		reasons = append(reasons, ReasonFirstObservationStore)
		points += u.firstObservationBonus
	}
	if dataCompleted(in.Payload) {
		reasons = append(reasons, ReasonDataCompletion)
		points += u.dataCompletionBonus
	}

	_ = u.persist(ctx, in.Receipt, points, reasons)
	return points, reasons
}

func (u *LoyaltyAwardUseCase) persist(ctx context.Context, receipt entities.Receipt, points int, reasons []string) error {
	reason := strings.Join(reasons, ";")
	if err := u.repo.AwardForReceipt(ctx, receipt.UserID, receipt.ID, points, reason); err != nil {
		if err == ports.ErrAlreadyAwarded {
			return nil
		}
		log.Printf("loyalty: award for receipt %s failed: %v (swallowed, receipt already confirmed)", receipt.ID, err)
		return nil
	}
	return nil
}

// dataCompleted returns true when the user filled every optional field on
// the editable summary: purchase_date, total > 0, and on every item the
// quantity, unit_price, and currency. (unit_price and currency are
// validated elsewhere on the confirm path; this function is intentionally
// self-contained so award rules are independent of validation rules.)
func dataCompleted(p entities.ConfirmPayload) bool {
	if strings.TrimSpace(p.PurchaseDate) == "" {
		return false
	}
	if p.Total <= 0 {
		return false
	}
	for _, item := range p.Items {
		if item.Quantity == nil || item.UnitPrice == nil || item.Currency == nil || strings.TrimSpace(*item.Currency) == "" {
			return false
		}
	}
	return true
}
