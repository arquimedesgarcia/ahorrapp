package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/usecase"
)

type uploadNoopStub struct{}

func (s uploadNoopStub) Execute(context.Context, usecase.UploadInput) (usecase.UploadResult, error) {
	return usecase.UploadResult{}, nil
}

type getEditableStub struct{}

func (s getEditableStub) Execute(context.Context, string, string) (*entities.EditableSummary, error) {
	return &entities.EditableSummary{
		ReceiptID: "r-1",
		Status:    entities.ReceiptStatusNeedsReview,
		Store:     entities.StoreSummary{Name: "Central Market"},
		Items: []entities.EditableItem{{
			RawText: "ARROZ",
		}},
	}, nil
}

type confirmNoopStub struct{}

func (s confirmNoopStub) Execute(context.Context, string, string, entities.ConfirmPayload) error {
	return nil
}

func TestReceiptGetHandler_ReturnsEditableSummary(t *testing.T) {
	h := NewReceiptHandler(uploadNoopStub{}, getEditableStub{}, confirmNoopStub{})
	router := newTestRouter(NewHealthHandler(fakeHealthUseCase{}), h.RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/receipts/r-1", nil)
	req.Header.Set("X-User-ID", "u-1")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var got entities.EditableSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != entities.ReceiptStatusNeedsReview {
		t.Fatalf("expected NEEDS_REVIEW, got %s", got.Status)
	}
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
}
