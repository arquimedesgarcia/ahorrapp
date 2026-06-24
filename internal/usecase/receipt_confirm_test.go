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
func (s confirmRepoStub) GetByID(context.Context, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s confirmRepoStub) MarkNeedsReview(context.Context, string, entities.EditableSummary) error {
	panic("not used")
}
func (s confirmRepoStub) ConfirmReceipt(context.Context, string, string, entities.ConfirmPayload, []entities.PriceObservation) error {
	return nil
}
func (s confirmRepoStub) ResolveOrCreateStore(context.Context, entities.StoreSummary) (string, error) {
	return "store-1", nil
}
func (s confirmRepoStub) NormalizeProduct(context.Context, string) (string, string, error) {
	return "product-1", "milk", nil
}

type eventsStub struct{}

func (e eventsStub) EmitReceiptConfirmed(context.Context, string, string, int) error { return nil }

func TestReceiptConfirmUseCase_RejectsItemWithoutCurrency(t *testing.T) {
	uc := NewReceiptConfirmUseCase(confirmRepoStub{}, eventsStub{})

	err := uc.Execute(context.Background(), "r-1", "u-1", entities.ConfirmPayload{
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

func ptrFloat(v float64) *float64 { return &v }
