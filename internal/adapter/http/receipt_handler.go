package httpapi

import (
	"context"
	"net/http"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type ReceiptUploadCommand interface {
	Execute(ctx context.Context, in usecase.UploadInput) (usecase.UploadResult, error)
}

type ReceiptGetQuery interface {
	Execute(ctx context.Context, receiptID, userID string) (*entities.EditableSummary, error)
}

type ReceiptConfirmCommand interface {
	Execute(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload) error
}

type ReceiptHandler struct {
	upload  ReceiptUploadCommand
	get     ReceiptGetQuery
	confirm ReceiptConfirmCommand
}

func NewReceiptHandler(upload ReceiptUploadCommand, get ReceiptGetQuery, confirm ReceiptConfirmCommand) *ReceiptHandler {
	return &ReceiptHandler{upload: upload, get: get, confirm: confirm}
}

func (h *ReceiptHandler) RegisterRoutes(r chi.Router) {
	r.Post("/receipts", h.uploadReceipt)
	r.Get("/receipts/{id}", h.getReceipt)
	r.Post("/receipts/{id}/confirm", h.confirmReceipt)
}

func userIDFromRequest(r *http.Request) string {
	if fromHeader := r.Header.Get("X-User-ID"); fromHeader != "" {
		return fromHeader
	}
	if fromCtx, ok := r.Context().Value(userIDCtxKey).(string); ok {
		return fromCtx
	}
	return ""
}
