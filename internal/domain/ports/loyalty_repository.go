package ports

import (
	"context"
	"errors"

	"ahorrapp/internal/domain/entities"
)

// ErrAlreadyAwarded is returned by LoyaltyRepository.AwardForReceipt when
// a loyalty_transactions row already exists for the given receipt_id. The
// award use case swallows this and treats it as a no-op (FR-004).
var ErrAlreadyAwarded = errors.New("loyalty: already awarded for this receipt")

// ContributorStats summarizes what the user has contributed to the
// community price database. Used for the profile badge in the mobile app.
type ContributorStats struct {
	ReceiptsConfirmed int `json:"receipts_confirmed"`
	PriceObservations int `json:"price_observations"`
	UniqueStores      int `json:"unique_stores"`
	UniqueProducts    int `json:"unique_products"`
}

// LoyaltyRepository is the port for awarding points and reading a user's
// loyalty balance and history. Implementations MUST make AwardForReceipt
// idempotent on receipt_id at the data layer (unique partial index on
// loyalty_transactions.receipt_id).
type LoyaltyRepository interface {
	// AwardForReceipt inserts one loyalty_transactions row tied to receiptID
	// (unless points==0 && reason==daily_limit_reached, in which case the
	// row still references the receipt so the user sees the explanation),
	// atomically increments users.points by points, and returns
	// ErrAlreadyAwarded when receiptID already has an award row.
	AwardForReceipt(ctx context.Context, userID, receiptID string, points int, reason string) error

	// DailyGrantCount returns the number of distinct receipt_id rows for
	// the user on the current UTC day with points > 0. Used to apply the
	// LOYALTY_DAILY_AWARD_CAP rule (FR-005).
	DailyGrantCount(ctx context.Context, userID string) (int, error)

	// Balance returns the cached users.points value.
	Balance(ctx context.Context, userID string) (int, error)

	// History returns the latest `limit` movements in created_at DESC order.
	History(ctx context.Context, userID string, limit int) ([]entities.LoyaltyTransaction, error)

	// ContributorStats returns aggregate contribution counts for the user:
	// confirmed receipts, price observations, unique stores, unique
	// products. Used to surface the "contribuidor" badge in the profile.
	ContributorStats(ctx context.Context, userID string) (ContributorStats, error)
}
