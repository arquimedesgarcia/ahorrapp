package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/usecase"
)

type uploadSequenceStub struct {
	count int
}

func (s *uploadSequenceStub) Execute(context.Context, usecase.UploadInput) (usecase.UploadResult, error) {
	s.count++
	if s.count == 1 {
		return usecase.UploadResult{ReceiptID: "r-dup", Status: "PENDING", Duplicate: false}, nil
	}
	return usecase.UploadResult{ReceiptID: "r-dup", Status: "PENDING", Duplicate: true}, nil
}

type staticGetStub struct{}

func (s staticGetStub) Execute(context.Context, string, string) (*entities.EditableSummary, error) {
	return &entities.EditableSummary{ReceiptID: "r-dup"}, nil
}

type staticConfirmStub struct{}

func (s staticConfirmStub) Execute(context.Context, string, string, entities.ConfirmPayload) (usecase.ConfirmResult, error) {
	return usecase.ConfirmResult{}, nil
}

func TestReceiptUploadHandler_HappyAndDuplicate(t *testing.T) {
	upload := &uploadSequenceStub{}
	h := NewReceiptHandler(upload, staticGetStub{}, listNoopStub{}, staticConfirmStub{})
	router := newTestRouter(NewHealthHandler(fakeHealthUseCase{}), h.RegisterRoutes)

	first := doUploadRequest(t, router)
	if first.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", first.Code)
	}
	var firstPayload usecase.UploadResult
	if err := json.Unmarshal(first.Body.Bytes(), &firstPayload); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if firstPayload.Duplicate {
		t.Fatalf("expected first upload non-duplicate")
	}

	second := doUploadRequest(t, router)
	if second.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", second.Code)
	}
	var secondPayload usecase.UploadResult
	if err := json.Unmarshal(second.Body.Bytes(), &secondPayload); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if !secondPayload.Duplicate {
		t.Fatalf("expected duplicate response on second upload")
	}
	if secondPayload.ReceiptID != firstPayload.ReceiptID {
		t.Fatalf("expected same receipt id for duplicate")
	}
}

func doUploadRequest(t *testing.T, router http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("image", "receipt.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte("img-data")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close form writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/receipts", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-User-ID", "u-dup")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}
