package usecase

import (
	"context"
	"fmt"

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

// ReceiptListUseCase returns the most recent receipts for a user, paged.
// Defaults: 20 per page, 0 offset. The mobile reads ?limit= and ?offset=
// when the user scrolls but the MVP does not need infinite scroll yet.
type ReceiptListUseCase struct {
	repo ports.ReceiptRepository
}

func NewReceiptListUseCase(repo ports.ReceiptRepository) *ReceiptListUseCase {
	return &ReceiptListUseCase{repo: repo}
}

func (u *ReceiptListUseCase) Execute(ctx context.Context, userID string, limit, offset int) ([]entities.ReceiptListItem, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.ListByUser(ctx, userID, limit, offset)
}
