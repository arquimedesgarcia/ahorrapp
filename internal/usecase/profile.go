package usecase

import (
	"context"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ProfileUseCase struct {
	users ports.UserRepository
}

func NewProfileUseCase(users ports.UserRepository) *ProfileUseCase {
	return &ProfileUseCase{users: users}
}

type PointsResponse struct {
	TotalPoints        int                    `json:"total_points"`
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

	resp := &PointsResponse{TotalPoints: points}
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
