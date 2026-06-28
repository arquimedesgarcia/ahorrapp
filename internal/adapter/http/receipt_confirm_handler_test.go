package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/usecase"
)

type uploadDummyStub struct{}

func (s uploadDummyStub) Execute(context.Context, usecase.UploadInput) (usecase.UploadResult, error) {
	return usecase.UploadResult{}, nil
}

type getDummyStub struct{}

func (s getDummyStub) Execute(context.Context, string, string) (*entities.EditableSummary, error) {
	return &entities.EditableSummary{ReceiptID: "r-1"}, nil
}

type confirmConditionalStub struct{}

func (s confirmConditionalStub) Execute(_ context.Context, _ string, _ string, payload entities.ConfirmPayload) error {
	for i, item := range payload.Items {
		if item.Currency == nil || *item.Currency == "" {
			return fmt.Errorf("items[%d].currency missing", i)
		}
	}
	return nil
}

func TestReceiptConfirmHandler_SuccessAndMissingCurrency(t *testing.T) {
	h := NewReceiptHandler(uploadDummyStub{}, getDummyStub{}, confirmConditionalStub{})
	router := newTestRouter(NewHealthHandler(fakeHealthUseCase{}), h.RegisterRoutes)

	goodBody := []byte(`{"store":{"name":"Store"},"purchase_date":"2026-06-24","total":1.0,"items":[{"raw_text":"PAN","quantity":1,"unit_price":1.0,"currency":"USD"}]}`)
	goodReq := httptest.NewRequest(http.MethodPost, "/api/v1/receipts/r-1/confirm", bytes.NewReader(goodBody))
	goodReq.Header.Set("X-User-ID", "u-1")
	goodReq.Header.Set("Content-Type", "application/json")
	goodRR := httptest.NewRecorder()
	router.ServeHTTP(goodRR, goodReq)
	if goodRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", goodRR.Code)
	}

	badBody := []byte(`{"store":{"name":"Store"},"purchase_date":"2026-06-24","total":1.0,"items":[{"raw_text":"PAN","quantity":1,"unit_price":1.0}]}`)
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/receipts/r-1/confirm", bytes.NewReader(badBody))
	badReq.Header.Set("X-User-ID", "u-1")
	badReq.Header.Set("Content-Type", "application/json")
	badRR := httptest.NewRecorder()
	router.ServeHTTP(badRR, badReq)
	if badRR.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", badRR.Code)
	}
}
