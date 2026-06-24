package usecase

import (
	"context"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ReceiptGetUseCase struct {
	repo ports.ReceiptRepository
}

func NewReceiptGetUseCase(repo ports.ReceiptRepository) *ReceiptGetUseCase {
	return &ReceiptGetUseCase{repo: repo}
}

func (u *ReceiptGetUseCase) Execute(ctx context.Context, receiptID, userID string) (*entities.EditableSummary, error) {
	return u.repo.GetByIDForUser(ctx, receiptID, userID)
}
