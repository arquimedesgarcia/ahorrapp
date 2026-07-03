package usecase

import (
	"context"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

// LoyaltyQueryUseCase builds the response for GET /api/v1/me/loyalty.
// Balance is the stored users.points cache; History is the latest
// loyaltyHistoryLimit (defined in loyalty_reasons.go) movements in
// created_at DESC order. Contributor surfaces aggregate counts so the
// mobile profile can render a "contribuidor" badge without a second
// round-trip.
type LoyaltyQueryUseCase struct {
	repo ports.LoyaltyRepository
}

func NewLoyaltyQueryUseCase(repo ports.LoyaltyRepository) *LoyaltyQueryUseCase {
	return &LoyaltyQueryUseCase{repo: repo}
}

type LoyaltyResponse struct {
	Balance     int                  `json:"balance"`
	Level       string               `json:"level"`
	Contributor ContributorOut       `json:"contributor"`
	History     []LoyaltyMovementOut `json:"history"`
}

type ContributorOut struct {
	ReceiptsConfirmed int `json:"receipts_confirmed"`
	PriceObservations int `json:"price_observations"`
	UniqueStores      int `json:"unique_stores"`
	UniqueProducts    int `json:"unique_products"`
}

type LoyaltyMovementOut struct {
	ID        string  `json:"id"`
	Points    int     `json:"points"`
	Reason    string  `json:"reason"`
	CreatedAt string  `json:"created_at"`
	ReceiptID *string `json:"receipt_id,omitempty"`
}

func (u *LoyaltyQueryUseCase) GetLoyalty(ctx context.Context, userID string) (*LoyaltyResponse, error) {
	balance, err := u.repo.Balance(ctx, userID)
	if err != nil {
		return nil, err
	}
	movements, err := u.repo.History(ctx, userID, loyaltyHistoryLimit)
	if err != nil {
		return nil, err
	}
	stats, err := u.repo.ContributorStats(ctx, userID)
	if err != nil {
		// Non-fatal: profile still works without contributor badge.
		stats = ports.ContributorStats{}
	}
	out := make([]LoyaltyMovementOut, 0, len(movements))
	for _, m := range movements {
		out = append(out, toMovementOut(m))
	}
	return &LoyaltyResponse{
		Balance: balance,
		Level:   levelFor(balance),
		Contributor: ContributorOut{
			ReceiptsConfirmed: stats.ReceiptsConfirmed,
			PriceObservations: stats.PriceObservations,
			UniqueStores:      stats.UniqueStores,
			UniqueProducts:    stats.UniqueProducts,
		},
		History: out,
	}, nil
}

// levelFor maps a point balance to a contributor tier. The thresholds are
// intentionally small for the MVP: even one confirmed receipt lands the
// user in "Bronce" so the badge feels earned quickly. The spec is in
// specs/006-loyalty-points (to be tightened once the value proposition is
// validated with real users).
func levelFor(points int) string {
	switch {
	case points >= 500:
		return "Platino"
	case points >= 200:
		return "Oro"
	case points >= 50:
		return "Plata"
	default:
		return "Bronce"
	}
}

func toMovementOut(m entities.LoyaltyTransaction) LoyaltyMovementOut {
	return LoyaltyMovementOut{
		ID:        m.ID,
		Points:    m.Points,
		Reason:    m.Reason,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ReceiptID: m.ReceiptID,
	}
}
