package usecase

import (
	"context"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ProfileUseCase struct {
	users   ports.UserRepository
	loyalty ports.LoyaltyRepository
}

func NewProfileUseCase(users ports.UserRepository, loyalty ports.LoyaltyRepository) *ProfileUseCase {
	return &ProfileUseCase{users: users, loyalty: loyalty}
}

type PointsResponse struct {
	TotalPoints        int                    `json:"total_points"`
	Level              string                 `json:"level"`
	Contributor        ContributorOut         `json:"contributor"`
	RecentTransactions []RecentTransactionOut `json:"recent_transactions"`
}

type RecentTransactionOut struct {
	ID        string `json:"id"`
	Points    int    `json:"points"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

func (u *ProfileUseCase) GetPoints(ctx context.Context, userID string) (*PointsResponse, error) {
	points, err := u.users.GetPoints(ctx, userID)
	if err != nil {
		return nil, err
	}

	txs, err := u.users.RecentTransactions(ctx, userID, 10)
	if err != nil {
		txs = nil
	}

	stats, err := u.loyalty.ContributorStats(ctx, userID)
	if err != nil {
		stats = ports.ContributorStats{}
	}

	resp := &PointsResponse{
		TotalPoints: points,
		Level:       levelFor(points),
		Contributor: ContributorOut{
			ReceiptsConfirmed: stats.ReceiptsConfirmed,
			PriceObservations: stats.PriceObservations,
			UniqueStores:      stats.UniqueStores,
			UniqueProducts:    stats.UniqueProducts,
		},
	}
	resp.RecentTransactions = toRecentTxOuts(txs)
	return resp, nil
}

func toRecentTxOuts(txs []entities.LoyaltyTransaction) []RecentTransactionOut {
	if len(txs) == 0 {
		return []RecentTransactionOut{}
	}
	out := make([]RecentTransactionOut, 0, len(txs))
	for _, tx := range txs {
		out = append(out, RecentTransactionOut{
			ID:        tx.ID,
			Points:    tx.Points,
			Reason:    tx.Reason,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return out
}
