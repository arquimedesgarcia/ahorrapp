package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
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

func (h *ReceiptHandler) uploadReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data := make([]byte, 0, 1<<20)
	buf := make([]byte, 4096)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	out, err := h.upload.Execute(r.Context(), usecase.UploadInput{UserID: userID, Data: data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ReceiptHandler) getReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	out, err := h.get.Execute(r.Context(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if out == nil {
		http.Error(w, "receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ReceiptHandler) confirmReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var payload entities.ConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if err := h.confirm.Execute(r.Context(), id, userID, payload); err != nil {
		http.Error(w, fmt.Sprintf("confirm receipt: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userIDFromRequest(r *http.Request) string {
	if fromHeader := r.Header.Get("X-User-ID"); fromHeader != "" {
		return fromHeader
	}
	if fromCtx, ok := r.Context().Value("user_id").(string); ok {
		return fromCtx
	}
	return ""
}
