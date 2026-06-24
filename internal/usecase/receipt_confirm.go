package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type ReceiptConfirmUseCase struct {
	repo   ports.ReceiptRepository
	events ports.ReceiptEvents
}

func NewReceiptConfirmUseCase(repo ports.ReceiptRepository, events ports.ReceiptEvents) *ReceiptConfirmUseCase {
	return &ReceiptConfirmUseCase{repo: repo, events: events}
}

func (u *ReceiptConfirmUseCase) Execute(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload) error {
	if userID == "" || receiptID == "" {
		return fmt.Errorf("user id and receipt id are required")
	}
	if _, err := time.Parse("2006-01-02", payload.PurchaseDate); err != nil {
		return fmt.Errorf("purchase_date must use format YYYY-MM-DD")
	}
	if len(payload.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	observations := make([]entities.PriceObservation, 0, len(payload.Items))
	storeID, err := u.repo.ResolveOrCreateStore(ctx, payload.Store)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, item := range payload.Items {
		if strings.TrimSpace(item.RawText) == "" {
			return fmt.Errorf("item raw_text is required")
		}
		if item.UnitPrice == nil {
			return fmt.Errorf("item unit_price is required")
		}
		if item.Currency == nil || strings.TrimSpace(*item.Currency) == "" {
			return fmt.Errorf("item currency is required")
		}
		productID, _, err := u.repo.NormalizeProduct(ctx, item.RawText)
		if err != nil {
			return err
		}
		observations = append(observations, entities.PriceObservation{
			ProductID:  productID,
			StoreID:    storeID,
			UnitPrice:  *item.UnitPrice,
			Currency:   strings.TrimSpace(*item.Currency),
			ObservedAt: now,
			ReceiptID:  receiptID,
		})
	}

	if err := u.repo.ConfirmReceipt(ctx, receiptID, userID, payload, observations); err != nil {
		return err
	}

	if u.events != nil {
		if err := u.events.EmitReceiptConfirmed(ctx, receiptID, userID, len(observations)); err != nil {
			return err
		}
	}

	return nil
}
