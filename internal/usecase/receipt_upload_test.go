package usecase

import (
	"context"
	"testing"

	"ahorrapp/internal/domain/entities"
)

type uploadRepoStub struct {
	existing *entities.Receipt
	created  bool
}

func (s *uploadRepoStub) FindByUserAndImageHash(context.Context, string, string) (*entities.Receipt, error) {
	return s.existing, nil
}

func (s *uploadRepoStub) CreatePendingReceipt(context.Context, string, string, string) (*entities.Receipt, error) {
	s.created = true
	return &entities.Receipt{ID: "new-id", Status: entities.ReceiptStatusPending}, nil
}

func (s *uploadRepoStub) GetByIDForUser(context.Context, string, string) (*entities.EditableSummary, error) {
	panic("not used")
}
func (s *uploadRepoStub) GetByID(context.Context, string) (*entities.Receipt, error) {
	panic("not used")
}
func (s *uploadRepoStub) MarkNeedsReview(context.Context, string, entities.EditableSummary) error {
	panic("not used")
}
func (s *uploadRepoStub) ConfirmReceipt(context.Context, string, string, entities.ConfirmPayload, []entities.PriceObservation) error {
	panic("not used")
}
func (s *uploadRepoStub) ResolveOrCreateStore(context.Context, entities.StoreSummary) (string, bool, error) {
	panic("not used")
}
func (s *uploadRepoStub) NormalizeProduct(context.Context, string) (string, string, error) {
	return "p-1", "milk", nil
}
func (s *uploadRepoStub) ListByUser(context.Context, string, int, int) ([]entities.ReceiptListItem, error) {
	return nil, nil
}

type storageStub struct{}

func (s storageStub) Upload(context.Context, string, []byte) (string, error) {
	return "http://minio/object", nil
}

type queueStub struct{ enqueued bool }

func (q *queueStub) Enqueue(context.Context, string) error   { q.enqueued = true; return nil }
func (q *queueStub) Dequeue(context.Context) (string, error) { panic("not used") }

func TestReceiptUploadUseCase_DuplicateReturnsExisting(t *testing.T) {
	repo := &uploadRepoStub{existing: &entities.Receipt{ID: "existing-id", Status: entities.ReceiptStatusPending}}
	queue := &queueStub{}
	uc := NewReceiptUploadUseCase(repo, storageStub{}, queue)

	out, err := uc.Execute(context.Background(), UploadInput{UserID: "u-1", Data: []byte("image")})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !out.Duplicate || out.ReceiptID != "existing-id" {
		t.Fatalf("expected duplicate existing-id, got %+v", out)
	}
	if repo.created {
		t.Fatalf("did not expect new receipt creation")
	}
	if queue.enqueued {
		t.Fatalf("did not expect queue enqueue for duplicate")
	}
}
