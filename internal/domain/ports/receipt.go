package ports

import (
	"context"

	"ahorrapp/internal/domain/entities"
)

type ReceiptRepository interface {
	FindByUserAndImageHash(ctx context.Context, userID, imageHash string) (*entities.Receipt, error)
	CreatePendingReceipt(ctx context.Context, userID, imageURL, imageHash string) (*entities.Receipt, error)
	GetByIDForUser(ctx context.Context, receiptID, userID string) (*entities.EditableSummary, error)
	GetByID(ctx context.Context, receiptID string) (*entities.Receipt, error)
	MarkNeedsReview(ctx context.Context, receiptID string, summary entities.EditableSummary) error
	ConfirmReceipt(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload, observations []entities.PriceObservation) error
	ResolveOrCreateStore(ctx context.Context, store entities.StoreSummary) (string, error)
	NormalizeProduct(ctx context.Context, rawName string) (string, string, error)
}

type OCRQueue interface {
	Enqueue(ctx context.Context, receiptID string) error
	Dequeue(ctx context.Context) (string, error)
}

type ReceiptEvents interface {
	EmitReceiptConfirmed(ctx context.Context, receiptID, userID string, observations int) error
}
