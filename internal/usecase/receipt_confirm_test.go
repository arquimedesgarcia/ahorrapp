package usecase

import (
	"context"
	"testing"

	"ahorrapp/internal/domain/entities"
)

type confirmRepoStub struct{}

func (s confirmRepoStub) FindByUserAndImageHash(context.Context, string, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s confirmRepoStub) CreatePendingReceipt(context.Context, string, string, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s confirmRepoStub) GetByIDForUser(context.Context, string, string) (*entities.EditableSummary, error) {
	panic("not used")
}
func (s confirmRepoStub) GetByID(_ context.Context, id string) (*entities.Receipt, error) {
	return &entities.Receipt{
		ID:     id,
		UserID: "u-1",
		Status: entities.ReceiptStatusConfirmed,
	}, nil
}
func (s confirmRepoStub) ListByUser(context.Context, string, int, int) ([]entities.ReceiptListItem, error) {
	return nil, nil
}
func (s confirmRepoStub) MarkNeedsReview(context.Context, string, entities.EditableSummary) error {
	panic("not used")
}
func (s confirmRepoStub) ConfirmReceipt(context.Context, string, string, entities.ConfirmPayload, []entities.PriceObservation) error {
	return nil
}
func (s confirmRepoStub) ResolveOrCreateStore(context.Context, entities.StoreSummary) (string, bool, error) {
	return "store-1", false, nil
}
func (s confirmRepoStub) NormalizeProduct(context.Context, string) (string, string, error) {
	return "product-1", "milk", nil
}

type eventsStub struct{}

func (e eventsStub) EmitReceiptConfirmed(context.Context, string, string, int) error { return nil }

func TestReceiptConfirmUseCase_RejectsItemWithoutCurrency(t *testing.T) {
	recomputeUC := NewPriceAggregateRecomputeUseCase(&stubRankingRepo{})
	uc := NewReceiptConfirmUseCase(confirmRepoStub{}, eventsStub{}, recomputeUC, nil, nil)

	_, err := uc.Execute(context.Background(), "r-1", "u-1", entities.ConfirmPayload{
		Store:        entities.StoreSummary{Name: "Store"},
		PurchaseDate: "2026-06-24",
		Total:        2.0,
		Items: []entities.EditableItem{{
			RawText:   "LECHE",
			UnitPrice: ptrFloat(1.0),
		}},
	})
	if err == nil {
		t.Fatalf("expected error for missing currency")
	}
}

func TestReceiptConfirmUseCase_ReturnsPointsEarned(t *testing.T) {
	recomputeUC := NewPriceAggregateRecomputeUseCase(&stubRankingRepo{})
	loyaltyRepo := &fakeLoyaltyRepo{}
	loyaltyUC := NewLoyaltyAwardUseCase(loyaltyRepo, 10, 5, 3, 0)
	uc := NewReceiptConfirmUseCase(confirmRepoStub{}, eventsStub{}, recomputeUC, loyaltyUC, nil)

	cur := "Bs"
	result, err := uc.Execute(context.Background(), "r-1", "u-1", entities.ConfirmPayload{
		Store:        entities.StoreSummary{Name: "Store"},
		PurchaseDate: "2026-06-24",
		Total:        100.0,
		Items: []entities.EditableItem{{
			RawText:   "LECHE",
			UnitPrice: ptrFloat(1.0),
			Currency:  &cur,
		}},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.PointsEarned <= 0 {
		t.Errorf("expected points_earned > 0, got %d (reasons=%v)", result.PointsEarned, result.Reasons)
	}
	if len(result.Reasons) == 0 {
		t.Errorf("expected at least one reason")
	}
	// Verify the loyalty repo got an award call
	if len(loyaltyRepo.awardedCalls) != 1 {
		t.Errorf("expected 1 loyalty award call, got %d", len(loyaltyRepo.awardedCalls))
	}
}

func ptrFloat(v float64) *float64 { return &v }
