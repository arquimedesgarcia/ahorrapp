package usecase

import (
	"context"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"
)

type processRepoStub struct {
	marked bool
	saved  entities.EditableSummary
}

func (s *processRepoStub) FindByUserAndImageHash(context.Context, string, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s *processRepoStub) CreatePendingReceipt(context.Context, string, string, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s *processRepoStub) GetByIDForUser(context.Context, string, string) (*entities.EditableSummary, error) {
	panic("not used")
}
func (s *processRepoStub) GetByID(context.Context, string) (*entities.Receipt, error) {
	return &entities.Receipt{ID: "r-1", ImageURL: "http://img"}, nil
}
func (s *processRepoStub) MarkNeedsReview(_ context.Context, _ string, summary entities.EditableSummary) error {
	s.marked = true
	s.saved = summary
	return nil
}
func (s *processRepoStub) ConfirmReceipt(context.Context, string, string, entities.ConfirmPayload, []entities.PriceObservation) error {
	panic("not used")
}
func (s *processRepoStub) ResolveOrCreateStore(context.Context, entities.StoreSummary) (string, error) {
	panic("not used")
}
func (s *processRepoStub) NormalizeProduct(context.Context, string) (string, string, error) {
	panic("not used")
}

type ocrStub struct{ text string }

func (o ocrStub) Extract(context.Context, string) (ports.RawOCRResult, error) {
	return ports.RawOCRResult{RawText: o.text}, nil
}

func TestReceiptProcessUseCase_TransitionsToNeedsReview(t *testing.T) {
	repo := &processRepoStub{}
	uc := NewReceiptProcessUseCase(repo, ocrStub{text: "SUPERMARKET CENTRAL\nDATE 2026-06-21\nARROZ 1 x 2.40 USD"})

	if err := uc.Execute(context.Background(), "r-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !repo.marked {
		t.Fatalf("expected MarkNeedsReview call")
	}
	if repo.saved.Status != entities.ReceiptStatusNeedsReview {
		t.Fatalf("expected NEEDS_REVIEW, got %s", repo.saved.Status)
	}
}

func TestReceiptProcessUseCase_UnreadableFallback(t *testing.T) {
	repo := &processRepoStub{}
	uc := NewReceiptProcessUseCase(repo, ocrStub{text: "### ??? ###"})

	if err := uc.Execute(context.Background(), "r-1"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.saved.Store.Name == "" {
		t.Fatalf("expected fallback store name")
	}
}
