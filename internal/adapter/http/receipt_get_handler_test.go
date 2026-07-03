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

func (s confirmNoopStub) Execute(context.Context, string, string, entities.ConfirmPayload) (usecase.ConfirmResult, error) {
	return usecase.ConfirmResult{}, nil
}

type listNoopStub struct{}

func (s listNoopStub) Execute(context.Context, string, int, int) ([]entities.ReceiptListItem, error) {
	return []entities.ReceiptListItem{}, nil
}

func TestReceiptGetHandler_ReturnsEditableSummary(t *testing.T) {
	h := NewReceiptHandler(uploadNoopStub{}, getEditableStub{}, listNoopStub{}, confirmNoopStub{})
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

func TestReceiptListHandler_ReturnsItems(t *testing.T) {
	list := &listWithItemsStub{items: []entities.ReceiptListItem{
		{ID: "r-1", Status: entities.ReceiptStatusConfirmed, StoreName: "FARMATODO", ItemCount: 3, Total: ptrFloat(1656.83)},
		{ID: "r-2", Status: entities.ReceiptStatusNeedsReview, StoreName: "Unknown", ItemCount: 0},
	}}
	h := NewReceiptHandler(uploadNoopStub{}, getEditableStub{}, list, confirmNoopStub{})
	router := newTestRouter(NewHealthHandler(fakeHealthUseCase{}), h.RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/receipts", nil)
	req.Header.Set("X-User-ID", "u-1")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	var resp struct {
		Items []entities.ReceiptListItem `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].StoreName != "FARMATODO" {
		t.Errorf("first item store = %q, want FARMATODO", resp.Items[0].StoreName)
	}
}

type listWithItemsStub struct {
	items []entities.ReceiptListItem
}

func (s *listWithItemsStub) Execute(_ context.Context, _ string, _, _ int) ([]entities.ReceiptListItem, error) {
	return s.items, nil
}

func ptrFloat(v float64) *float64 { return &v }
