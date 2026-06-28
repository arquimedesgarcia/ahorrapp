package httpapi

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/usecase"
)

type uploadStub struct{}

func (s uploadStub) Execute(context.Context, usecase.UploadInput) (usecase.UploadResult, error) {
	return usecase.UploadResult{ReceiptID: "r-1", Status: "PENDING"}, nil
}

type getStub struct{}

func (s getStub) Execute(context.Context, string, string) (*entities.EditableSummary, error) {
	return &entities.EditableSummary{ReceiptID: "r-1"}, nil
}

type confirmStub struct{}

func (s confirmStub) Execute(context.Context, string, string, entities.ConfirmPayload) error {
	return nil
}

func TestReceiptUploadHandler_AcceptsMultipart(t *testing.T) {
	h := NewReceiptHandler(uploadStub{}, getStub{}, confirmStub{})
	router := newTestRouter(NewHealthHandler(fakeHealthUseCase{}), h.RegisterRoutes)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "receipt.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte("img-data")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/receipts", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-User-ID", "u-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}
