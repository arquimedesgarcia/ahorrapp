package httpapi

import (
	"net/http"
	"strconv"

	"ahorrapp/internal/domain/entities"

	"github.com/go-chi/chi/v5"
)

func (h *ReceiptHandler) getReceipt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "receipt id required", http.StatusBadRequest)
		return
	}

	summary, err := h.get.Execute(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "fetch receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if summary == nil {
		http.Error(w, "receipt not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *ReceiptHandler) listReceipts(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		http.Error(w, "missing user id", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.list.Execute(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "list receipts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []entities.ReceiptListItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}
