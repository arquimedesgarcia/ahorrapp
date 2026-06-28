package postgres

import (
	"context"

	"ahorrapp/internal/domain/entities"
)

type PriceObservationRepository struct {
	receiptRepo *ReceiptRepository
}

func NewPriceObservationRepository(receiptRepo *ReceiptRepository) *PriceObservationRepository {
	return &PriceObservationRepository{receiptRepo: receiptRepo}
}

func (r *PriceObservationRepository) PersistWithReceiptConfirmation(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload, observations []entities.PriceObservation) error {
	return r.receiptRepo.ConfirmReceipt(ctx, receiptID, userID, payload, observations)
}
